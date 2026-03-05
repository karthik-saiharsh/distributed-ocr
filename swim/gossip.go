package swim

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	ProbeInterval = 500 * time.Millisecond // how often we ping a peer
	ProbeTimeout  = 200 * time.Millisecond // how long to wait for an ACK
	SuspicionMult = 4                      // multiplier for suspect to dead timeout
)

// gossip manages the heartbeat loop and failure detection timers.
type gossip struct {
	service       *SWIMService
	stopCh        chan struct{}
	suspectMu     sync.Mutex
	suspectTimers map[string]*time.Timer // nodeID -> pending suspect timer
	deadTimers    map[string]*time.Timer // nodeID -> pending dead timer
	reclaimTimers map[string]*time.Timer // nodeID -> pending cleanup timer
}

// newGossip creates a gossip engine bound to the given service.
func newGossip(svc *SWIMService) *gossip {
	return &gossip{
		service:       svc,
		stopCh:        make(chan struct{}),
		suspectTimers: make(map[string]*time.Timer),
		deadTimers:    make(map[string]*time.Timer),
		reclaimTimers: make(map[string]*time.Timer),
	}
}

// Start launches the heartbeat goroutine. Call Stop() to shut it down.
func (g *gossip) Start() {
	go g.heartbeatLoop()
	log.Printf("[SWIM] gossip engine started (interval=%s, timeout=%s, suspectMult=%d)", ProbeInterval, ProbeTimeout, SuspicionMult)
}

// Stop signals the heartbeat loop to exit.
func (g *gossip) Stop() {
	close(g.stopCh)
}

// heartbeatLoop fires on every heartbeatInterval tick.
func (g *gossip) heartbeatLoop() {
	ticker := time.NewTicker(ProbeInterval)
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

	// log.Printf("[SWIM] PING → %s (%s)", peer.ID, peerAddr)
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

	g.suspectTimers[nodeID] = time.AfterFunc(ProbeTimeout, func() {
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
	if t, ok := g.deadTimers[nodeID]; ok {
		t.Stop()
		delete(g.deadTimers, nodeID)
	}
	if t, ok := g.reclaimTimers[nodeID]; ok {
		t.Stop()
		delete(g.reclaimTimers, nodeID)
	}
}

// onTimeout is called when a PING ACK was not received within ProbeTimeout.
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

		if svc.NotifySuspect != nil {
			svc.NotifySuspect(*node)
		}

		svc.emitUpdate()

		// Arm the dead timer
		// NARROW LOCK: Only lock for map access, not during callbacks!
		g.suspectMu.Lock()
		if t, ok := g.deadTimers[nodeID]; ok {
			t.Stop()
		}
		g.deadTimers[nodeID] = time.AfterFunc(SuspicionMult*ProbeInterval, func() {
			g.onDeadTimeout(nodeID)
		})
		g.suspectMu.Unlock()
	}
}

// onDeadTimeout is called when the suspect node has not recovered before the timeout
func (g *gossip) onDeadTimeout(nodeID string) {
	svc := g.service
	node, ok := svc.Members.Get(nodeID)
	if !ok {
		return
	}
	if node.Status == StatusSuspect {
		node.Status = StatusDead
		node.LastUpdated = time.Now()
		svc.Members.Set(node)
		log.Printf("[SWIM] node %s is now DEAD (timeout)", nodeID)

		if svc.NotifyDead != nil {
			svc.NotifyDead(*node)
		}

		svc.emitUpdate()

		// Arm the Reclaim timer to actually wipe this node from memory
		// NARROW LOCK: Only lock for map access
		g.suspectMu.Lock()
		if t, ok := g.reclaimTimers[nodeID]; ok {
			t.Stop()
		}
		g.reclaimTimers[nodeID] = time.AfterFunc(10*time.Second, func() {
			g.onReclaimTimeout(nodeID)
		})
		g.suspectMu.Unlock()
	}
}

// onReclaimTimeout permanently scrubs a dead node from the membership map.
func (g *gossip) onReclaimTimeout(nodeID string) {
	svc := g.service

	g.suspectMu.Lock()
	delete(g.reclaimTimers, nodeID)
	g.suspectMu.Unlock()

	node, ok := svc.Members.Get(nodeID)
	if !ok {
		return
	}

	if node.Status == StatusDead {
		svc.Members.Delete(nodeID)
		log.Printf("[SWIM] node %s reclaimed (permanently removed from list)", nodeID)
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
			// OSCILLATION DEFENSE: Determine if someone thinks we are Suspect or Dead
			if remote.Status != StatusAlive && remote.Incarnation >= svc.Self.Incarnation {
				// Bump our incarnation and broadcast we are ALIVE
				g.bumpSelfIncarnation()
			}
			continue // never overwrite self from others
		}

		local, exists := svc.Members.Get(remote.ID)
		if !exists {
			// Brand-new node discovered via gossip.
			svc.Members.Set(remote)
			log.Printf("[SWIM] gossip: discovered node %s (%s:%d)", remote.ID, remote.IP, remote.Port)
			changed = true
		} else {
			// Remote record exists. Prioritize higher Incarnation.
			if remote.Incarnation > local.Incarnation {
				svc.Members.Set(remote)
				changed = true
			} else if remote.Incarnation == local.Incarnation && remote.LastUpdated.After(local.LastUpdated) {
				// Same incarnation, take the freshest wallclock update
				svc.Members.Set(remote)
				changed = true
			}
		}
	}

	if changed {
		svc.emitUpdate()
	}
}

// bumpSelfIncarnation increments the local incarnation number and broadcasts.
func (g *gossip) bumpSelfIncarnation() {
	svc := g.service
	svc.Members.mu.Lock()
	defer svc.Members.mu.Unlock() // Use global Members lock directly for 'Self' consistency

	svc.Self.Incarnation++
	svc.Self.Status = StatusAlive
	svc.Self.LastUpdated = time.Now()

	// Update within map
	svc.Members.members[svc.Self.ID] = &svc.Self

	log.Printf("[SWIM] Self node refuted suspicion! Bumping incarnation to %d", svc.Self.Incarnation)

	// Async emit to avoid deadlock
	go svc.emitUpdate()
}
