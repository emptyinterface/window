package sysinfo

import (
	"reflect"
	"testing"
)

func TestParseNetStat(t *testing.T) {

	const output = `TcpExt: SyncookiesSent SyncookiesRecv SyncookiesFailed EmbryonicRsts PruneCalled RcvPruned OfoPruned OutOfWindowIcmps LockDroppedIcmps ArpFilter TW TWRecycled TWKilled PAWSPassive PAWSActive PAWSEstab DelayedACKs DelayedACKLocked DelayedACKLost ListenOverflows ListenDrops TCPPrequeued TCPDirectCopyFromBacklog TCPDirectCopyFromPrequeue TCPPrequeueDropped TCPHPHits TCPHPHitsToUser TCPPureAcks TCPHPAcks TCPRenoRecovery TCPSackRecovery TCPSACKReneging TCPFACKReorder TCPSACKReorder TCPRenoReorder TCPTSReorder TCPFullUndo TCPPartialUndo TCPDSACKUndo TCPLossUndo TCPLostRetransmit TCPRenoFailures TCPSackFailures TCPLossFailures TCPFastRetrans TCPForwardRetrans TCPSlowStartRetrans TCPTimeouts TCPLossProbes TCPLossProbeRecovery TCPRenoRecoveryFail TCPSackRecoveryFail TCPSchedulerFailed TCPRcvCollapsed TCPDSACKOldSent TCPDSACKOfoSent TCPDSACKRecv TCPDSACKOfoRecv TCPAbortOnData TCPAbortOnClose TCPAbortOnMemory TCPAbortOnTimeout TCPAbortOnLinger TCPAbortFailed TCPMemoryPressures TCPSACKDiscard TCPDSACKIgnoredOld TCPDSACKIgnoredNoUndo TCPSpuriousRTOs TCPMD5NotFound TCPMD5Unexpected TCPSackShifted TCPSackMerged TCPSackShiftFallback TCPBacklogDrop TCPMinTTLDrop TCPDeferAcceptDrop IPReversePathFilter TCPTimeWaitOverflow TCPReqQFullDoCookies TCPReqQFullDrop TCPRetransFail TCPRcvCoalesce TCPOFOQueue TCPOFODrop TCPOFOMerge TCPChallengeACK TCPSYNChallenge TCPFastOpenActive TCPFastOpenPassive TCPFastOpenPassiveFail TCPFastOpenListenOverflow TCPFastOpenCookieReqd TCPSpuriousRtxHostQueues BusyPollRxPackets
TcpExt: 0 0 9 9 19 0 0 0 0 0 93 0 0 0 0 0 242515 56 8940 0 0 198 0 192 0 351523 0 2974290 3109 0 365 0 0 0 0 0 0 0 0 32 0 0 6 0 369 22 10 231 1079 909 0 23 0 192 8940 0 0 0 0 0 0 1 0 0 0 0 0 0 5 0 0 3 4 1188 0 0 0 0 0 0 0 6 440152 2889 0 0 0 0 0 0 0 0 0 45 0
IpExt: InNoRoutes InTruncatedPkts InMcastPkts OutMcastPkts InBcastPkts OutBcastPkts InOctets OutOctets InMcastOctets OutMcastOctets InBcastOctets OutBcastOctets InCsumErrors
IpExt: 2 0 0 0 0 0 1176894455 1304988547 0 0 0 0 0
`
	expected := NetStat{SyncookiesSent: 0x0, SyncookiesRecv: 0x0, SyncookiesFailed: 0x9, EmbryonicRsts: 0x9, PruneCalled: 0x13, RcvPruned: 0x0, OfoPruned: 0x0, OutOfWindowIcmps: 0x0, LockDroppedIcmps: 0x0, ArpFilter: 0x0, TW: 0x5d, TWRecycled: 0x0, TWKilled: 0x0, PAWSPassive: 0x0, PAWSActive: 0x0, PAWSEstab: 0x0, DelayedACKs: 0x3b353, DelayedACKLocked: 0x38, DelayedACKLost: 0x22ec, ListenOverflows: 0x0, ListenDrops: 0x0, TCPPrequeued: 0xc6, TCPDirectCopyFromBacklog: 0x0, TCPDirectCopyFromPrequeue: 0xc0, TCPPrequeueDropped: 0x0, TCPHPHits: 0x55d23, TCPHPHitsToUser: 0x0, TCPPureAcks: 0x2d6252, TCPHPAcks: 0xc25, TCPRenoRecovery: 0x0, TCPSackRecovery: 0x16d, TCPSACKReneging: 0x0, TCPFACKReorder: 0x0, TCPSACKReorder: 0x0, TCPRenoReorder: 0x0, TCPTSReorder: 0x0, TCPFullUndo: 0x0, TCPPartialUndo: 0x0, TCPDSACKUndo: 0x0, TCPLossUndo: 0x20, TCPLoss: 0x0, TCPLostRetransmit: 0x0, TCPRenoFailures: 0x0, TCPSackFailures: 0x6, TCPLossFailures: 0x0, TCPFastRetrans: 0x171, TCPForwardRetrans: 0x16, TCPSlowStartRetrans: 0xa, TCPTimeouts: 0xe7, TCPLossProbes: 0x437, TCPLossProbeRecovery: 0x38d, TCPRenoRecoveryFail: 0x0, TCPSackRecoveryFail: 0x17, TCPSchedulerFailed: 0x0, TCPRcvCollapsed: 0xc0, TCPDSACKOldSent: 0x22ec, TCPDSACKOfoSent: 0x0, TCPDSACKRecv: 0x0, TCPDSACKOfoRecv: 0x0, TCPAbortOnSyn: 0x0, TCPAbortOnData: 0x0, TCPAbortOnClose: 0x0, TCPAbortOnMemory: 0x0, TCPAbortOnTimeout: 0x1, TCPAbortOnLinger: 0x0, TCPAbortFailed: 0x0, TCPMemoryPressures: 0x0, TCPSACKDiscard: 0x0, TCPDSACKIgnoredOld: 0x0, TCPDSACKIgnoredNoUndo: 0x0, TCPSpuriousRTOs: 0x5, TCPMD5NotFound: 0x0, TCPMD5Unexpected: 0x0, TCPSackShifted: 0x3, TCPSackMerged: 0x4, TCPSackShiftFallback: 0x4a4, TCPBacklogDrop: 0x0, TCPMinTTLDrop: 0x0, TCPDeferAcceptDrop: 0x0, IPReversePathFilter: 0x0, TCPTimeWaitOverflow: 0x0, TCPReqQFullDoCookies: 0x0, TCPReqQFullDrop: 0x0, TCPRetransFail: 0x6, TCPRcvCoalesce: 0x6b758, TCPOFOQueue: 0xb49, TCPOFODrop: 0x0, TCPOFOMerge: 0x0, TCPChallengeACK: 0x0, TCPSYNChallenge: 0x0, TCPFastOpenActive: 0x0, TCPFastOpenActiveFail: 0x0, TCPFastOpenPassive: 0x0, TCPFastOpenPassiveFail: 0x0, TCPFastOpenListenOverflow: 0x0, TCPFastOpenCookieReqd: 0x0, TCPSpuriousRtxHostQueues: 0x2d, BusyPollRxPackets: 0x0, TCPAutoCorking: 0x0, TCPFromZeroWindowAdv: 0x0, TCPToZeroWindowAdv: 0x0, TCPWantZeroWindowAdv: 0x0, TCPSynRetrans: 0x0, TCPOrigDataSent: 0x0, InNoRoutes: 0x2, InTruncatedPkts: 0x0, InMcastPkts: 0x0, OutMcastPkts: 0x0, InBcastPkts: 0x0, OutBcastPkts: 0x0, InOctets: 0x4625fbf7, OutOctets: 0x4dc88b83, InMcastOctets: 0x0, OutMcastOctets: 0x0, InBcastOctets: 0x0, OutBcastOctets: 0x0, InCsumErrors: 0x0, InNoECTPkts: 0x0, InECT1Pkts: 0x0, InECT0Pkts: 0x0, InCEPkts: 0x0}

	stat := NewStat()

	if err := (&stat.NetStat).Parse([]byte(output)); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stat.NetStat, expected) {
		t.Error("parse mismatch")
		dumpDiff(expected, stat.NetStat)
	}
}
