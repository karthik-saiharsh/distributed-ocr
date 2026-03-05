package swim

import (
	"sync"
	"time"
)

// NodeStatus represents the health state of a node in the cluster.
type NodeStatus string

const (
	StatusAlive   NodeStatus = "Alive"
	StatusSuspect NodeStatus = "Suspect"
	StatusDead    NodeStatus = "Dead"
)

// Node represents a single machine participating in the cluster.
type Node struct {
	ID          string     `json:"id"`
	IP          string     `json:"ip"`
	Port        int        `json:"port"`
	Status      NodeStatus `json:"status"`
	Incarnation uint32     `json:"incarnation"`
	LastUpdated time.Time  `json:"lastUpdated"`
}

// MembershipList is a map of all known nodes, keyed by Node ID.
type MembershipList struct {
	mu      sync.RWMutex
	members map[string]*Node
}

// NewMembershipList creates an initialised, empty MembershipList.
func NewMembershipList() *MembershipList {
	return &MembershipList{
		members: make(map[string]*Node),
	}
}

// Set inserts or replaces a node in the list.
func (ml *MembershipList) Set(n *Node) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.members[n.ID] = n
}

// Get returns the node with the given ID and a boolean indicating existence.
func (ml *MembershipList) Get(id string) (*Node, bool) {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	n, ok := ml.members[id]
	return n, ok
}

// Delete removes a node from the list by ID.
func (ml *MembershipList) Delete(id string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	delete(ml.members, id)
}

// GetAll returns a snapshot (copy) of all nodes as a slice.
// Safe to call concurrently; callers receive a stable copy.
func (ml *MembershipList) GetAll() []Node {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	nodes := make([]Node, 0, len(ml.members))
	for _, n := range ml.members {
		nodes = append(nodes, *n)
	}
	return nodes
}

// Len returns the number of known members.
func (ml *MembershipList) Len() int {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return len(ml.members)
}

// MessageType identifies the purpose of a SWIM protocol message.
type MessageType string

const (
	MsgPing    MessageType = "PING"
	MsgAck     MessageType = "ACK"
	MsgPingReq MessageType = "PING_REQ"
)

// Message is the wire format for all SWIM UDP packets.
// Members carries the sender's current membership snapshot (gossip payload).
type Message struct {
	Type       MessageType `json:"type"`
	SenderID   string      `json:"senderID"`
	SenderAddr string      `json:"senderAddr"` // "ip:port" of the sender
	Members    []Node      `json:"members"`    // gossip payload
}
