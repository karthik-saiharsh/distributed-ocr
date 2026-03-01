package swim

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	heartbeatInterval = 500 * time.Millisecond // how often we ping a peer
	pingTimeout       = 200 * time.Millisecond // how long to wait for an ACK
)

// gossip manages the heartbeat loop and failure detection timers.
type gossip struct {
	service       *SWIMService
	stopCh        chan struct{}
	suspectMu     sync.Mutex
	suspectTimers map[string]*time.Timer // nodeID -> pending suspect timer
}

// newGossip creates a gossip engine bound to the given service.
func newGossip(svc *SWIMService) *gossip {
	return &gossip{
		service:       svc,
		stopCh:        make(chan struct{}),
		suspectTimers: make(map[string]*time.Timer),
	}
}

// Start launches the heartbeat goroutine. Call Stop() to shut it down.
func (g *gossip) Start() {
	go g.heartbeatLoop()
	log.Printf("[SWIM] gossip engine started (interval=%s, timeout=%s)", heartbeatInterval, pingTimeout)
}

// Stop signals the heartbeat loop to exit.
func (g *gossip) Stop() {
	close(g.stopCh)
}

// heartbeatLoop fires on every heartbeatInterval tick.
func (g *gossip) heartbeatLoop() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			g.tick()
		case <-g.stopCh:
			log.Println("[SWIM] gossip heartbeat stopped")
			return
		}
	}
}

// tick selects a random peer, sends a PING, and arms a suspect timer.
func (g *gossip) tick() {
	peer := g.selectRandomPeer()
	if peer == nil {
		return // no peers yet
	}

	svc := g.service
	ping := Message{
		Type:       MsgPing,
		SenderID:   svc.Self.ID,
		SenderAddr: svc.selfAddr(),
		Members:    svc.Members.GetAll(),
	}

	peerAddr := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	if err := svc.net.SendMessage(peerAddr, ping); err != nil {
		log.Printf("[SWIM] failed to PING %s (%s): %v", peer.ID, peerAddr, err)
		// Treat a send failure as a missed ACK immediately.
		g.armSuspectTimer(peer.ID)
		return
	}

	log.Printf("[SWIM] PING → %s (%s)", peer.ID, peerAddr)
	g.armSuspectTimer(peer.ID)
}

// selectRandomPeer picks one Alive or Suspect node at random, excluding self.
func (g *gossip) selectRandomPeer() *Node {
	all := g.service.Members.GetAll()
	selfID := g.service.Self.ID

	candidates := make([]Node, 0, len(all))
	for _, n := range all {
		if n.ID != selfID && n.Status != StatusDead {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	chosen := candidates[rand.Intn(len(candidates))]
	return &chosen
}

// armSuspectTimer starts a timer for nodeID. If no ACK arrives within
// pingTimeout, the node is marked Suspect.
func (g *gossip) armSuspectTimer(nodeID string) {
	g.suspectMu.Lock()
	defer g.suspectMu.Unlock()

	// Cancel any existing timer for this node before arming a new one.
	if t, ok := g.suspectTimers[nodeID]; ok {
		t.Stop()
	}

	g.suspectTimers[nodeID] = time.AfterFunc(pingTimeout, func() {
		g.onTimeout(nodeID)
	})
}

// cancelSuspectTimer stops and removes the pending timer for nodeID.
// Called by the network layer when an ACK is received.
func (g *gossip) cancelSuspectTimer(nodeID string) {
	g.suspectMu.Lock()
	defer g.suspectMu.Unlock()
	if t, ok := g.suspectTimers[nodeID]; ok {
		t.Stop()
		delete(g.suspectTimers, nodeID)
	}
}

// onTimeout is called when a PING ACK was not received within pingTimeout.
func (g *gossip) onTimeout(nodeID string) {
	svc := g.service
	node, ok := svc.Members.Get(nodeID)
	if !ok {
		return
	}
	if node.Status == StatusAlive {
		node.Status = StatusSuspect
		node.LastUpdated = time.Now()
		svc.Members.Set(node)
		log.Printf("[SWIM] node %s is now SUSPECT (no ACK)", nodeID)
		svc.emitUpdate()
	}
}

// mergeMemberList integrates an incoming gossip payload into the local list.
// New nodes are added; existing nodes get their timestamps updated if the
// incoming record is newer.
func (g *gossip) mergeMemberList(incoming []Node) {
	svc := g.service
	changed := false

	for i := range incoming {
		remote := &incoming[i]
		if remote.ID == svc.Self.ID {
			continue // never overwrite self
		}
		local, exists := svc.Members.Get(remote.ID)
		if !exists {
			// Brand-new node discovered via gossip.
			svc.Members.Set(remote)
			log.Printf("[SWIM] gossip: discovered node %s (%s:%d)", remote.ID, remote.IP, remote.Port)
			changed = true
		} else if remote.LastUpdated.After(local.LastUpdated) {
			// Remote record is fresher — adopt it.
			svc.Members.Set(remote)
			changed = true
		}
	}

	if changed {
		svc.emitUpdate()
	}
}
