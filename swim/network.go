package swim

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const maxPacketSize = 65507 // max UDP payload size

// Network handles all UDP socket I/O for the SWIM protocol.
type Network struct {
	conn    *net.UDPConn
	service *SWIMService
}

// newNetwork creates a Network bound to the given SWIMService.
func newNetwork(svc *SWIMService) *Network {
	return &Network{service: svc}
}

// StartListener opens a UDP socket on the given port and starts the receive loop.
func (n *Network) StartListener(port int) error {
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("swim: failed to open UDP socket on port %d: %w", port, err)
	}
	n.conn = conn
	log.Printf("[SWIM] UDP listener started on port %d", port)
	go n.receiveLoop()
	return nil
}

// Close shuts down the UDP socket, which also unblocks the receive loop.
func (n *Network) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

// SendMessage marshals msg to JSON and sends it to the given "ip:port" address.
func (n *Network) SendMessage(addr string, msg Message) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return fmt.Errorf("swim: invalid address %q: %w", addr, err)
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("swim: failed to marshal message: %w", err)
	}
	_, err = n.conn.WriteToUDP(data, udpAddr)
	if err != nil {
		return fmt.Errorf("swim: failed to send to %s: %w", addr, err)
	}
	return nil
}

// receiveLoop continuously reads incoming UDP packets and dispatches them.
// It exits when the underlying connection is closed.
func (n *Network) receiveLoop() {
	buf := make([]byte, maxPacketSize)
	for {
		nBytes, remoteAddr, err := n.conn.ReadFromUDP(buf)
		if err != nil {
			// A closed connection is expected on shutdown; log other errors.
			if n.conn != nil {
				log.Printf("[SWIM] receive loop exiting: %v", err)
			}
			return
		}
		var msg Message
		if err := json.Unmarshal(buf[:nBytes], &msg); err != nil {
			log.Printf("[SWIM] failed to unmarshal packet from %s: %v", remoteAddr, err)
			continue
		}
		go n.handleMessage(msg, remoteAddr)
	}
}

// handleMessage routes an incoming message to the appropriate handler.
func (n *Network) handleMessage(msg Message, from *net.UDPAddr) {
	svc := n.service

	// Always merge the gossip payload first.
	svc.mergeMemberList(msg.Members)

	switch msg.Type {
	case MsgPing:
		// Update the sender's last-seen timestamp.
		if node, ok := svc.Members.Get(msg.SenderID); ok {
			node.LastUpdated = time.Now()
			node.Status = StatusAlive
			svc.Members.Set(node)
		} else {
			// First time we see this node — add it.
			newNode := &Node{
				ID:          msg.SenderID,
				IP:          from.IP.String(),
				Port:        from.Port,
				Status:      StatusAlive,
				LastUpdated: time.Now(),
			}
			svc.Members.Set(newNode)
			log.Printf("[SWIM] discovered new node %s (%s)", msg.SenderID, from)
		}
		svc.emitUpdate()

		// Reply with an ACK carrying our own member list.
		ack := Message{
			Type:       MsgAck,
			SenderID:   svc.Self.ID,
			SenderAddr: svc.selfAddr(),
			Members:    svc.Members.GetAll(),
		}
		if err := n.SendMessage(from.String(), ack); err != nil {
			log.Printf("[SWIM] failed to send ACK to %s: %v", from, err)
		}

	case MsgAck:
		// Cancel any pending suspect timer for this sender.
		svc.cancelSuspectTimer(msg.SenderID)

		// Mark the sender as Alive.
		if node, ok := svc.Members.Get(msg.SenderID); ok {
			if node.Status != StatusAlive {
				node.Status = StatusAlive
				node.LastUpdated = time.Now()
				svc.Members.Set(node)
				log.Printf("[SWIM] node %s is now Alive", msg.SenderID)
				svc.emitUpdate()
			} else {
				node.LastUpdated = time.Now()
				svc.Members.Set(node)
			}
		}

	case MsgPingReq:
		// Indirect ping: forward a PING on behalf of the requester.
		// (Placeholder for full PING-REQ support in a future phase.)
		log.Printf("[SWIM] received PING_REQ from %s (not yet fully implemented)", msg.SenderID)

	default:
		log.Printf("[SWIM] unknown message type %q from %s", msg.Type, from)
	}
}
