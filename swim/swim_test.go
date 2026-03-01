package swim

import (
	"encoding/json"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Phase 1: MembershipList CRUD
// ---------------------------------------------------------------------------

func TestMembershipListSet(t *testing.T) {
	ml := NewMembershipList()
	n := &Node{ID: "node-1", IP: "192.168.1.10", Port: 7946, Status: StatusAlive, LastUpdated: time.Now()}
	ml.Set(n)

	got, ok := ml.Get("node-1")
	if !ok {
		t.Fatal("expected node-1 to exist after Set")
	}
	if got.IP != "192.168.1.10" {
		t.Errorf("expected IP 192.168.1.10, got %s", got.IP)
	}
}

func TestMembershipListDelete(t *testing.T) {
	ml := NewMembershipList()
	ml.Set(&Node{ID: "node-2", Status: StatusAlive})
	ml.Delete("node-2")

	if _, ok := ml.Get("node-2"); ok {
		t.Fatal("expected node-2 to be deleted")
	}
}

func TestMembershipListGetAll(t *testing.T) {
	ml := NewMembershipList()
	ml.Set(&Node{ID: "a", Status: StatusAlive})
	ml.Set(&Node{ID: "b", Status: StatusSuspect})

	all := ml.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(all))
	}
}

func TestMembershipListLen(t *testing.T) {
	ml := NewMembershipList()
	if ml.Len() != 0 {
		t.Errorf("expected 0 initially, got %d", ml.Len())
	}
	ml.Set(&Node{ID: "x"})
	if ml.Len() != 1 {
		t.Errorf("expected 1, got %d", ml.Len())
	}
}

// ---------------------------------------------------------------------------
// Phase 1: Message marshal / unmarshal
// ---------------------------------------------------------------------------

func TestMessageMarshalRoundTrip(t *testing.T) {
	msg := Message{
		Type:       MsgPing,
		SenderID:   "node-1",
		SenderAddr: "192.168.1.10:7946",
		Members: []Node{
			{ID: "node-2", IP: "192.168.1.11", Port: 7946, Status: StatusAlive, LastUpdated: time.Now()},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got Message
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Type != MsgPing {
		t.Errorf("expected type PING, got %s", got.Type)
	}
	if got.SenderID != "node-1" {
		t.Errorf("expected senderID node-1, got %s", got.SenderID)
	}
	if len(got.Members) != 1 || got.Members[0].ID != "node-2" {
		t.Errorf("unexpected members: %+v", got.Members)
	}
}

// ---------------------------------------------------------------------------
// Phase 3: mergeMemberList (gossip discovery)
// ---------------------------------------------------------------------------

func TestMergeMemberListAddsNewNodes(t *testing.T) {
	svc := newTestService("self-1", "127.0.0.1", 17946)
	g := svc.gossip

	incoming := []Node{
		{ID: "remote-1", IP: "10.0.0.2", Port: 7946, Status: StatusAlive, LastUpdated: time.Now()},
		{ID: "remote-2", IP: "10.0.0.3", Port: 7946, Status: StatusAlive, LastUpdated: time.Now()},
	}
	g.mergeMemberList(incoming)

	if _, ok := svc.Members.Get("remote-1"); !ok {
		t.Error("remote-1 should have been added by merge")
	}
	if _, ok := svc.Members.Get("remote-2"); !ok {
		t.Error("remote-2 should have been added by merge")
	}
}

func TestMergeMemberListUpdatesTimestamp(t *testing.T) {
	svc := newTestService("self-1", "127.0.0.1", 17947)
	g := svc.gossip

	old := time.Now().Add(-10 * time.Second)
	svc.Members.Set(&Node{ID: "peer-1", IP: "10.0.0.2", Port: 7946, Status: StatusAlive, LastUpdated: old})

	newer := time.Now()
	g.mergeMemberList([]Node{
		{ID: "peer-1", IP: "10.0.0.2", Port: 7946, Status: StatusAlive, LastUpdated: newer},
	})

	node, _ := svc.Members.Get("peer-1")
	if !node.LastUpdated.Equal(newer) {
		t.Errorf("expected timestamp to be updated to newer value")
	}
}

func TestMergeMemberListIgnoresSelf(t *testing.T) {
	svc := newTestService("self-1", "127.0.0.1", 17948)
	g := svc.gossip

	// Attempt to overwrite self with a Dead status via gossip.
	g.mergeMemberList([]Node{
		{ID: "self-1", IP: "127.0.0.1", Port: 17948, Status: StatusDead, LastUpdated: time.Now()},
	})

	self, _ := svc.Members.Get("self-1")
	if self != nil && self.Status == StatusDead {
		t.Error("self should never be overwritten by gossip")
	}
}

// ---------------------------------------------------------------------------
// Phase 3: suspect timer
// ---------------------------------------------------------------------------

func TestSuspectTimerCancellation(t *testing.T) {
	svc := newTestService("self-1", "127.0.0.1", 17949)
	g := svc.gossip

	svc.Members.Set(&Node{ID: "peer-1", IP: "10.0.0.2", Port: 7946, Status: StatusAlive, LastUpdated: time.Now()})

	// Arm a timer, then immediately cancel it.
	g.armSuspectTimer("peer-1")
	g.cancelSuspectTimer("peer-1")

	// Wait longer than pingTimeout to confirm the node was NOT marked Suspect.
	time.Sleep(pingTimeout + 50*time.Millisecond)

	node, _ := svc.Members.Get("peer-1")
	if node != nil && node.Status == StatusSuspect {
		t.Error("node should not be Suspect after timer was cancelled")
	}
}

func TestSuspectTimerFires(t *testing.T) {
	svc := newTestService("self-1", "127.0.0.1", 17950)
	g := svc.gossip

	svc.Members.Set(&Node{ID: "peer-2", IP: "10.0.0.3", Port: 7946, Status: StatusAlive, LastUpdated: time.Now()})
	g.armSuspectTimer("peer-2")

	// Wait for the timer to fire.
	time.Sleep(pingTimeout + 100*time.Millisecond)

	node, _ := svc.Members.Get("peer-2")
	if node == nil || node.Status != StatusSuspect {
		t.Errorf("expected peer-2 to be Suspect, got %+v", node)
	}
}

// ---------------------------------------------------------------------------
// Helper: create a SWIMService without starting the network/gossip loops.
// ---------------------------------------------------------------------------

func newTestService(id, ip string, port int) *SWIMService {
	svc := &SWIMService{
		Self: Node{
			ID:          id,
			IP:          ip,
			Port:        port,
			Status:      StatusAlive,
			LastUpdated: time.Now(),
		},
		Members: NewMembershipList(),
	}
	svc.net = newNetwork(svc)
	svc.gossip = newGossip(svc)
	// Register self (mirrors what Start() does).
	self := svc.Self
	svc.Members.Set(&self)
	return svc
}
