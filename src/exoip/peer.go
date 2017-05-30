package exoip

import (
	"fmt"
	"net"

	"github.com/pyr/egoscale/src/egoscale"
)

func newPeer(ego *egoscale.Client, peer string) *Peer {
	addr, err := net.ResolveUDPAddr("udp", peer)
	AssertSuccess(err)
	ip := addr.IP
	conn, err := net.DialUDP("udp", nil, addr)
	AssertSuccess(err)
	peerNic, err := findPeerNic(ego, ip.String())
	AssertSuccess(err)
	return &Peer{IP: ip, NicId: peerNic, LastSeen: 0, Conn: conn, Dead: false}
}

func (engine *Engine) findPeer(addr net.UDPAddr) *Peer {
	for _, p := range engine.Peers {
		if p.IP.Equal(addr.IP) {
			return p
		}
	}
	return nil
}

func (engine *Engine) updatePeer(addr net.UDPAddr, payload *Payload) {

	if !engine.ExoIP.Equal(payload.ExoIP) {
		Logger.Warning("peer sent message for wrong EIP")
		return
	}

	peer := engine.findPeer(addr)
	if peer == nil {
		Logger.Warning("peer not found in configuration")
		return
	}
	peer.Priority = payload.Priority
	peer.NicId = payload.NicId
	peer.LastSeen = CurrentTimeMillis()
}

func (engine *Engine) peerIsNewlyDead(now int64, peer *Peer) bool {
	peerDiff := now - peer.LastSeen
	dead := peerDiff > int64(engine.Interval*engine.DeadRatio)*1000
	if dead != peer.Dead {
		if dead {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, considering dead.", peer.IP, peerDiff))
		} else {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, is now back alive.", peer.IP, peerDiff))
		}
		peer.Dead = dead
		return dead
	}
	return false
}

func (engine *Engine) backupOf(peer *Peer) bool {
	return (!peer.Dead && peer.Priority < engine.Priority)
}

func (engine *Engine) handleDeadPeer(peer *Peer) {
	engine.ReleaseNic(peer.NicId)
}
