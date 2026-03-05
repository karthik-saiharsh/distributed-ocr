package swim

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const defaultPort = 7946 // well-known SWIM/memberlist port

// SWIMService is the top-level coordinator for the SWIM protocol.
// It owns the MembershipList, the UDP network layer, and the gossip engine.
type SWIMService struct {
	Self    Node
	Members *MembershipList

	net    *Network
	gossip *gossip
	ctx    context.Context // Wails runtime context for event emission

	// Hooks for logging and external reactions to membership changes
	NotifyAlive   func(Node)
	NotifySuspect func(Node)
	NotifyDead    func(Node)
}

// NewSWIMService constructs a SWIMService for the local node.
// selfID should be a unique identifier (e.g. a UUID or hostname).
// ip is the LAN IP of this machine; port is the UDP port to listen on.
func NewSWIMService(selfID, ip string, port int) *SWIMService {
	if port == 0 {
		port = defaultPort
	}
	svc := &SWIMService{
		Self: Node{
			ID:          selfID,
			IP:          ip,
			Port:        port,
			Status:      StatusAlive,
			LastUpdated: time.Now(),
		},
		Members: NewMembershipList(),
	}

	// Default hooks print to log but can be overridden.
	svc.NotifyAlive = func(n Node) { log.Printf("[Hook] Node %s (%s) is Alive", n.ID, n.IP) }
	svc.NotifySuspect = func(n Node) { log.Printf("[Hook] Node %s (%s) is Suspect", n.ID, n.IP) }
	svc.NotifyDead = func(n Node) { log.Printf("[Hook] Node %s (%s) is Dead", n.ID, n.IP) }

	svc.net = newNetwork(svc)
	svc.gossip = newGossip(svc)
	return svc
}

// Start begins listening for UDP packets and starts the gossip heartbeat.
// ctx must be the Wails application context (used for runtime.EventsEmit).
// The service registers itself in the MembershipList so it appears in the UI.
func (svc *SWIMService) Start(ctx context.Context) error {
	svc.ctx = ctx

	// Register self so the UI shows the local node immediately.
	self := svc.Self
	svc.Members.Set(&self)

	if err := svc.net.StartListener(svc.Self.Port); err != nil {
		return fmt.Errorf("swim: cannot start listener: %w", err)
	}

	svc.gossip.Start()
	log.Printf("[SWIM] service started — self: %s (%s:%d)", svc.Self.ID, svc.Self.IP, svc.Self.Port)
	return nil
}

// Stop gracefully shuts down the gossip engine and closes the UDP socket.
func (svc *SWIMService) Stop() {
	svc.gossip.Stop()
	svc.net.Close()
	log.Println("[SWIM] service stopped")
}

// GetClusterNodes is the Wails-bound method called by the frontend on load.
// It returns a stable snapshot of all currently known nodes.
func (svc *SWIMService) GetClusterNodes() []Node {
	return svc.Members.GetAll()
}

// KillNode stops the local heartbeat, simulating a node failure.
// Useful for testing how other nodes react to a dead peer.
func (svc *SWIMService) KillNode() {
	svc.gossip.Stop()
	svc.net.Close()
	log.Println("[SWIM] KillNode called — local node is now silent")
}

// emitUpdate fires a "cluster:update" Wails event with the current node list.
// Called internally whenever the membership state changes.
func (svc *SWIMService) emitUpdate() {
	if svc.ctx == nil {
		return // Wails context not yet available (startup race guard)
	}
	nodes := svc.Members.GetAll()
	runtime.EventsEmit(svc.ctx, "cluster:update", nodes)
}

// selfAddr returns the "ip:port" string for this node.
func (svc *SWIMService) selfAddr() string {
	return fmt.Sprintf("%s:%d", svc.Self.IP, svc.Self.Port)
}

// cancelSuspectTimer delegates to the gossip engine.
// Called by the network layer when an ACK arrives.
func (svc *SWIMService) cancelSuspectTimer(nodeID string) {
	svc.gossip.cancelSuspectTimer(nodeID)
}

// mergeMemberList delegates to the gossip engine.
// Called by the network layer on every incoming PING or ACK.
func (svc *SWIMService) mergeMemberList(members []Node) {
	svc.gossip.mergeMemberList(members)
}

// StartScan probes every usable host on the actual local subnet concurrently
// by sending a SWIM PING to each address. The subnet is derived from the real
// interface CIDR (e.g. /23, /24, /16) — not hardcoded. Any machine that has
// the app open will receive the PING and immediately reply with an ACK, which
// the existing receive loop processes: the peer is added to the MembershipList
// and a "cluster:update" event is emitted to refresh the frontend.
//
// Subnets larger than /16 are capped at /16 (65 534 hosts) to stay practical.
// If interface detection fails a /24 derived from the local IP is used as a
// safe fallback.
func (svc *SWIMService) StartScan() {
	network, err := localNetworkForIP(svc.Self.IP)
	if err != nil {
		log.Printf("[SWIM] StartScan: %v — falling back to /24", err)
		// Fallback: carve a /24 out of the local IP string.
		parts := strings.SplitN(svc.Self.IP, ".", 4)
		if len(parts) != 4 {
			log.Printf("[SWIM] StartScan: cannot parse local IP %q", svc.Self.IP)
			return
		}
		_, network, err = net.ParseCIDR(strings.Join(parts[:3], ".") + ".0/24")
		if err != nil {
			return
		}
	}

	targets := subnetHosts(network, svc.Self.IP)
	log.Printf("[SWIM] StartScan: probing %d hosts on %s (port %d)",
		len(targets), network.String(), svc.Self.Port)

	ping := Message{
		Type:       MsgPing,
		SenderID:   svc.Self.ID,
		SenderAddr: svc.selfAddr(),
		Members:    svc.Members.GetAll(),
	}

	var wg sync.WaitGroup
	for _, hostIP := range targets {
		addr := fmt.Sprintf("%s:%d", hostIP, svc.Self.Port)
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			_ = svc.net.SendMessage(a, ping) // offline hosts simply won't ACK
		}(addr)
	}

	// Also probe local development ports on localhost to support single-machine cluster testing
	for p := 7946; p <= 7956; p += 2 {
		addr := fmt.Sprintf("127.0.0.1:%d", p)
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			_ = svc.net.SendMessage(a, ping)
		}(addr)

		addrLAN := fmt.Sprintf("%s:%d", svc.Self.IP, p)
		wg.Add(1)
		go func(a string) {
			defer wg.Done()
			_ = svc.net.SendMessage(a, ping)
		}(addrLAN)
	}

	wg.Wait()
	log.Printf("[SWIM] StartScan complete — %s", network.String())
}

// localNetworkForIP walks the machine's network interfaces and returns the
// *net.IPNet whose CIDR contains localIP. This gives the real subnet mask
// (e.g. /23, /22) rather than assuming /24.
func localNetworkForIP(localIP string) (*net.IPNet, error) {
	target := net.ParseIP(localIP)
	if target == nil {
		return nil, fmt.Errorf("invalid local IP %q", localIP)
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("cannot list interface addresses: %w", err)
	}
	for _, addr := range addrs {
		cidr, ok := addr.(*net.IPNet)
		if !ok || cidr.IP.To4() == nil {
			continue
		}
		if cidr.Contains(target) {
			return cidr, nil
		}
	}
	return nil, fmt.Errorf("no IPv4 interface found containing %s", localIP)
}

// subnetHosts enumerates every usable host address in network, excluding
// selfIP, the network address, and the broadcast address.
// Subnets with a prefix shorter than /16 are treated as /16 to cap the scan
// at 65 534 addresses and avoid runaway probing on large corporate networks.
func subnetHosts(network *net.IPNet, selfIP string) []string {
	ip4 := network.IP.To4()
	if ip4 == nil {
		return nil
	}
	ones, _ := network.Mask.Size()
	if ones < 16 {
		ones = 16 // cap: don't scan more than a /16
	}
	mask := net.CIDRMask(ones, 32)
	base := ipToUint32(ip4.Mask(mask))
	bcast := base | ^ipToUint32([]byte(mask))
	self := ipToUint32(net.ParseIP(selfIP).To4())

	hosts := make([]string, 0, int(bcast-base-1))
	for addr := base + 1; addr < bcast; addr++ {
		if addr == self {
			continue
		}
		hosts = append(hosts, uint32ToIP(addr).String())
	}
	return hosts
}

func ipToUint32(ip net.IP) uint32 {
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func uint32ToIP(n uint32) net.IP {
	return net.IP{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
}

// DetectLocalIP returns the preferred outbound LAN IP of this machine.
// Falls back to "127.0.0.1" if detection fails.
func DetectLocalIP() string {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
