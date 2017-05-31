package exoip

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pyr/egoscale/src/egoscale"
)

// DefaultPort default process port
const DefaultPort = 12345
const protoVersion = "0201"
const skewMillis = 100
const skew time.Duration = 100 * time.Millisecond

// Verbose
var Verbose = false

func removeDash(r rune) rune {
	if r == '-' {
		return -1
	}
	return r
}

func convStrToUUID(ustr string) ([]byte, error) {

	if len(ustr) != 36 {
		return nil, fmt.Errorf("NicId %s has wrong length", ustr)
	}

	ustr = strings.ToLower(ustr)
	for _, c := range ustr {
		if !(c >= 'a' && c <= 'f') &&
			!(c >= '0' && c <= '9') &&
			!(c == '-') {
			return nil, errors.New("Bad characters in NicId")
		}
	}
	ustr = strings.Map(removeDash, ustr)
	if len(ustr) != 32 {
		return nil, errors.New("NicId has wrong length")
	}

	ba, err := hex.DecodeString(ustr)
	if err != nil {
		return nil, err
	}
	if len(ba) != 16 {
		return nil, errors.New("Bad NicID byte length")
	}
	return ba, nil
}

func convUUIDToStr(uuidbuf []byte) string {

	str := hex.EncodeToString(uuidbuf)

	hexuuid := fmt.Sprintf("%s-%s-%s-%s-%s",
		str[0:8], str[8:12], str[12:16], str[16:20], str[20:32])
	return hexuuid
}

// NewWatchdogEngine
func NewWatchdogEngine(client *egoscale.Client, ip string, interval int,
	prio int, deadRatio int, peers []string) *Engine {

	mserver, err := FindMetadataServer()
	AssertSuccess(err)
	nicid, err := FetchMyNic(client, mserver)
	uuidbuf, err := convStrToUUID(nicid)
	AssertSuccess(err)
	sendbuf := make([]byte, 24)
	protobuf, err := hex.DecodeString(protoVersion)
	AssertSuccess(err)
	netip := net.ParseIP(ip)
	if netip == nil {
		Logger.Crit("Could not parse IP")
		fmt.Fprintln(os.Stderr, "Could not parse IP")
		os.Exit(1)
	}
	netip = netip.To4()
	if netip == nil {
		Logger.Crit("Unsupported IPv6 Address")
		fmt.Fprintln(os.Stderr, "Unsupported IPv6 Address")
		os.Exit(1)
	}

	netbytes := []byte(netip)

	sendbuf[0] = protobuf[0]
	sendbuf[1] = protobuf[1]
	sendbuf[2] = byte(prio)
	sendbuf[3] = byte(prio)
	sendbuf[4] = netbytes[0]
	sendbuf[5] = netbytes[1]
	sendbuf[6] = netbytes[2]
	sendbuf[7] = netbytes[3]

	for i, b := range uuidbuf {
		sendbuf[i+8] = b
	}

	engine := Engine{
		DeadRatio:   deadRatio,
		Interval:    interval,
		Priority:    sendbuf[2],
		SendBuf:     sendbuf,
		Peers:       make([]*Peer, 0),
		State:       StateBackup,
		NicId:       nicid,
		ExoIP:       netip,
		Exo:         client,
		InitHoldOff: CurrentTimeMillis() + (1000 * int64(deadRatio) * int64(interval)) + skewMillis,
	}
	for _, p := range peers {
		engine.Peers = append(engine.Peers, newPeer(client, p))
	}
	return &engine
}

// NewEngine
func NewEngine(client *egoscale.Client, ip string) *Engine {
	netip := net.ParseIP(ip)
	if netip == nil {
		Logger.Crit("Could not parse IP")
		fmt.Fprintln(os.Stderr, "Could not parse IP")
		os.Exit(1)
	}
	netip = netip.To4()
	if netip == nil {
		Logger.Crit("Unsupported IPv6 Address")
		fmt.Fprintln(os.Stderr, "Unsupported IPv6 Address")
		os.Exit(1)
	}

	engine := Engine{
		ExoIP: netip,
		Exo:   client,
	}
	engine.FetchNicAndVm()
	return &engine
}
