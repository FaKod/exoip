package exoip

import (
	"time"
)

func (engine *Engine) switchToBackup() {
	Logger.Warning("switching to back-up state")
	engine.ReleaseNic(engine.NicId)
}

func (engine *Engine) switchToMaster() {
	Logger.Warning("switching to master state")
	engine.ObtainNic(engine.NicId)
}

func (engine *Engine) performStateTransition(state State) {

	if engine.State == state {
		if state == StateMaster {
			engine.SetNicRatioCounter = engine.SetNicRatioCounter - 1
			if engine.SetNicRatioCounter <= 0 {
				engine.SetNicRatioCounter = engine.SetNicRatio
				engine.ObtainNic(engine.NicId)
			}
		}
		return
	}

	engine.State = state

	if state == StateBackup {
		engine.switchToBackup()
	} else {
		engine.switchToMaster()
	}
}

func (engine *Engine) checkState() {

	time.Sleep(skew)

	now := CurrentTimeMillis()

	if now <= engine.InitHoldOff {
		return
	}

	deadPeers := make([]*Peer, 0)
	bestAdvertisement := true

	for _, peer := range engine.Peers {
		if engine.peerIsNewlyDead(now, peer) {
			deadPeers = append(deadPeers, peer)
		} else {
			if engine.backupOf(peer) {
				bestAdvertisement = false
			}
		}
	}

	if bestAdvertisement == false {
		engine.performStateTransition(StateBackup)
	} else {
		engine.performStateTransition(StateMaster)
	}

	for _, peer := range deadPeers {
		engine.handleDeadPeer(peer)
	}
}
