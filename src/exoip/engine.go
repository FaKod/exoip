package exoip

import (
	"time"
	"github.com/pyr/egoscale/src/egoscale"
)

const SkewMillis = 100
const Skew time.Duration = 100 * time.Millisecond

func NewEngine(vmid string, nicid string, client *egoscale.Client, interval int,
	vhid int, prio int, dead_ratio int, peers []string) *Engine {

	sendbuf := make([]byte, 2)
	sendbuf[0] = byte(vhid)
	sendbuf[1] = byte(prio)
	engine := Engine{
		DeadRatio: dead_ratio,
		Interval: interval,
		Priority: sendbuf[1],
		VHID: sendbuf[0],
		SendBuf: sendbuf,
		Peers: make([]*Peer, 0),
		State: StateBackup,
		ExoVM: vmid,
		ExoIP: nicid,
		Exo: client,
		InitHoldOff: CurrentTimeMillis() + (1000 * int64(dead_ratio) * int64(interval)) + SkewMillis,
	}
	for _, p := range(peers) {
		engine.Peers = append(engine.Peers, NewPeer(p))
	}
	return &engine
}
