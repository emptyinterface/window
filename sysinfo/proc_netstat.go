package sysinfo

import (
	"bufio"
	"bytes"
	"reflect"
	"strconv"
	"strings"
)

type (
	NetStat struct {
		// TcpExt
		SyncookiesSent            uint64
		SyncookiesRecv            uint64
		SyncookiesFailed          uint64
		EmbryonicRsts             uint64
		PruneCalled               uint64
		RcvPruned                 uint64
		OfoPruned                 uint64
		OutOfWindowIcmps          uint64
		LockDroppedIcmps          uint64
		ArpFilter                 uint64
		TW                        uint64
		TWRecycled                uint64
		TWKilled                  uint64
		PAWSPassive               uint64
		PAWSActive                uint64
		PAWSEstab                 uint64
		DelayedACKs               uint64
		DelayedACKLocked          uint64
		DelayedACKLost            uint64
		ListenOverflows           uint64
		ListenDrops               uint64
		TCPPrequeued              uint64
		TCPDirectCopyFromBacklog  uint64
		TCPDirectCopyFromPrequeue uint64
		TCPPrequeueDropped        uint64
		TCPHPHits                 uint64
		TCPHPHitsToUser           uint64
		TCPPureAcks               uint64
		TCPHPAcks                 uint64
		TCPRenoRecovery           uint64
		TCPSackRecovery           uint64
		TCPSACKReneging           uint64
		TCPFACKReorder            uint64
		TCPSACKReorder            uint64
		TCPRenoReorder            uint64
		TCPTSReorder              uint64
		TCPFullUndo               uint64
		TCPPartialUndo            uint64
		TCPDSACKUndo              uint64
		TCPLossUndo               uint64
		TCPLoss                   uint64
		TCPLostRetransmit         uint64
		TCPRenoFailures           uint64
		TCPSackFailures           uint64
		TCPLossFailures           uint64
		TCPFastRetrans            uint64
		TCPForwardRetrans         uint64
		TCPSlowStartRetrans       uint64
		TCPTimeouts               uint64
		TCPLossProbes             uint64
		TCPLossProbeRecovery      uint64
		TCPRenoRecoveryFail       uint64
		TCPSackRecoveryFail       uint64
		TCPSchedulerFailed        uint64
		TCPRcvCollapsed           uint64
		TCPDSACKOldSent           uint64
		TCPDSACKOfoSent           uint64
		TCPDSACKRecv              uint64
		TCPDSACKOfoRecv           uint64
		TCPAbortOnSyn             uint64
		TCPAbortOnData            uint64
		TCPAbortOnClose           uint64
		TCPAbortOnMemory          uint64
		TCPAbortOnTimeout         uint64
		TCPAbortOnLinger          uint64
		TCPAbortFailed            uint64
		TCPMemoryPressures        uint64
		TCPSACKDiscard            uint64
		TCPDSACKIgnoredOld        uint64
		TCPDSACKIgnoredNoUndo     uint64
		TCPSpuriousRTOs           uint64
		TCPMD5NotFound            uint64
		TCPMD5Unexpected          uint64
		TCPSackShifted            uint64
		TCPSackMerged             uint64
		TCPSackShiftFallback      uint64
		TCPBacklogDrop            uint64
		TCPMinTTLDrop             uint64
		TCPDeferAcceptDrop        uint64
		IPReversePathFilter       uint64
		TCPTimeWaitOverflow       uint64
		TCPReqQFullDoCookies      uint64
		TCPReqQFullDrop           uint64
		TCPRetransFail            uint64
		TCPRcvCoalesce            uint64
		TCPOFOQueue               uint64
		TCPOFODrop                uint64
		TCPOFOMerge               uint64
		TCPChallengeACK           uint64
		TCPSYNChallenge           uint64
		TCPFastOpenActive         uint64
		TCPFastOpenActiveFail     uint64
		TCPFastOpenPassive        uint64
		TCPFastOpenPassiveFail    uint64
		TCPFastOpenListenOverflow uint64
		TCPFastOpenCookieReqd     uint64
		TCPSpuriousRtxHostQueues  uint64
		BusyPollRxPackets         uint64
		TCPAutoCorking            uint64
		TCPFromZeroWindowAdv      uint64
		TCPToZeroWindowAdv        uint64
		TCPWantZeroWindowAdv      uint64
		TCPSynRetrans             uint64
		TCPOrigDataSent           uint64
		// IpExt
		InNoRoutes      uint64
		InTruncatedPkts uint64
		InMcastPkts     uint64
		OutMcastPkts    uint64
		InBcastPkts     uint64
		OutBcastPkts    uint64
		InOctets        uint64
		OutOctets       uint64
		InMcastOctets   uint64
		OutMcastOctets  uint64
		InBcastOctets   uint64
		OutBcastOctets  uint64
		InCsumErrors    uint64
		InNoECTPkts     uint64
		InECT1Pkts      uint64
		InECT0Pkts      uint64
		InCEPkts        uint64
	}
)

const NetStatCommand = `cat /proc/net/netstat`

func (_ *NetStat) Command() string {
	return NetStatCommand
}

func (ns *NetStat) Parse(b []byte) error {

	// output:
	// TcpExt: SyncookiesSent SyncookiesRecv SyncookiesFailed EmbryonicRsts PruneCalled RcvPruned OfoPruned OutOfWindowIcmps LockDroppedIcmps ArpFilter TW TWRecycled TWKilled PAWSPassive PAWSActive PAWSEstab DelayedACKs DelayedACKLocked DelayedACKLost ListenOverflows ListenDrops TCPPrequeued TCPDirectCopyFromBacklog TCPDirectCopyFromPrequeue TCPPrequeueDropped TCPHPHits TCPHPHitsToUser TCPPureAcks TCPHPAcks TCPRenoRecovery TCPSackRecovery TCPSACKReneging TCPFACKReorder TCPSACKReorder TCPRenoReorder TCPTSReorder TCPFullUndo TCPPartialUndo TCPDSACKUndo TCPLossUndo TCPLostRetransmit TCPRenoFailures TCPSackFailures TCPLossFailures TCPFastRetrans TCPForwardRetrans TCPSlowStartRetrans TCPTimeouts TCPLossProbes TCPLossProbeRecovery TCPRenoRecoveryFail TCPSackRecoveryFail TCPSchedulerFailed TCPRcvCollapsed TCPDSACKOldSent TCPDSACKOfoSent TCPDSACKRecv TCPDSACKOfoRecv TCPAbortOnData TCPAbortOnClose TCPAbortOnMemory TCPAbortOnTimeout TCPAbortOnLinger TCPAbortFailed TCPMemoryPressures TCPSACKDiscard TCPDSACKIgnoredOld TCPDSACKIgnoredNoUndo TCPSpuriousRTOs TCPMD5NotFound TCPMD5Unexpected TCPSackShifted TCPSackMerged TCPSackShiftFallback TCPBacklogDrop TCPMinTTLDrop TCPDeferAcceptDrop IPReversePathFilter TCPTimeWaitOverflow TCPReqQFullDoCookies TCPReqQFullDrop TCPRetransFail TCPRcvCoalesce TCPOFOQueue TCPOFODrop TCPOFOMerge TCPChallengeACK TCPSYNChallenge TCPFastOpenActive TCPFastOpenPassive TCPFastOpenPassiveFail TCPFastOpenListenOverflow TCPFastOpenCookieReqd TCPSpuriousRtxHostQueues BusyPollRxPackets
	// TcpExt: 0 0 9 9 19 0 0 0 0 0 93 0 0 0 0 0 242515 56 8940 0 0 198 0 192 0 351523 0 2974290 3109 0 365 0 0 0 0 0 0 0 0 32 0 0 6 0 369 22 10 231 1079 909 0 23 0 192 8940 0 0 0 0 0 0 1 0 0 0 0 0 0 5 0 0 3 4 1188 0 0 0 0 0 0 0 6 440152 2889 0 0 0 0 0 0 0 0 0 45 0
	// IpExt: InNoRoutes InTruncatedPkts InMcastPkts OutMcastPkts InBcastPkts OutBcastPkts InOctets OutOctets InMcastOctets OutMcastOctets InBcastOctets OutBcastOctets InCsumErrors
	// IpExt: 2 0 0 0 0 0 1176894455 1304988547 0 0 0 0 0

	v := reflect.ValueOf(ns).Elem()

	s := bufio.NewScanner(bytes.NewReader(b))

	for s.Scan() {

		fields := strings.Fields(s.Text())
		s.Scan()
		values := strings.Fields(s.Text())

		// trim off row title
		if len(fields) > 1 && len(values) > 1 {
			fields, values = fields[1:], values[1:]
		}

		for i, name := range fields {
			if f := v.FieldByName(name); f.CanSet() {
				n, _ := strconv.ParseUint(values[i], 10, 64)
				f.SetUint(n)
			}
		}

	}

	return s.Err()

}
