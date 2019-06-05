package sysinfo

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/emptyinterface/window/sysinfo/test"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"golang.org/x/crypto/ssh"
)

func NewSSHGzipHandler(t *testing.T, h func(req *ssh.Request, input []byte) (output []byte)) func(req *ssh.Request, channel ssh.Channel) {

	return func(req *ssh.Request, channel ssh.Channel) {

		if reqBashCommand := string(req.Payload); reqBashCommand != RemoteBashCommand {
			t.Errorf("Expected %q, got %q", RemoteBashCommand, reqBashCommand)
		}

		gzr, err := gzip.NewReader(channel)
		if err != nil {
			t.Error(err)
		}
		defer gzr.Close()

		input, err := ioutil.ReadAll(gzr)
		if err != nil {
			t.Error(err)
		}

		gzw := gzip.NewWriter(channel)
		gzw.Write(h(req, input))

		if err := gzw.Close(); err != nil {
			t.Error(err)
		}

		if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}); err != nil {
			t.Error(err)
		}

		if err := channel.Close(); err != nil {
			t.Error(err)
		}

	}
}

func TestSystemInformationCollector(t *testing.T) {

	const (
		TestCommand = `cat /proc/uptime && echo -n ===Jj52dgpmaF=== && cat /proc/stat && echo -n ===Jj52dgpmaF=== && cat /proc/meminfo && echo -n ===Jj52dgpmaF=== && cat /proc/net/netstat && echo -n ===Jj52dgpmaF=== && cat /proc/loadavg && echo -n ===Jj52dgpmaF=== && df -B1 && echo -n ===aRVZeDergP=== && df -i`
		TestOutput1 = `634791.08 5077082.86
===Jj52dgpmaF===cpu  42642 5461 18868 507226813 12965 3 452 22329 0 0
cpu0 2572 1692 1703 63401671 7941 0 1 2594 0 0
cpu1 24496 90 7268 63405853 617 2 450 1961 0 0
cpu2 2351 0 1572 63405250 615 0 0 2980 0 0
cpu3 2857 2 1741 63404991 469 0 0 2917 0 0
cpu4 2317 79 1841 63405587 683 0 0 2816 0 0
cpu5 2468 0 1358 63399196 435 0 0 3206 0 0
cpu6 2826 814 1544 63401298 803 0 0 2965 0 0
cpu7 2751 2780 1838 63402962 1398 0 0 2886 0 0
intr 171722677 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 19715879 643 333184 22 0 2123 0 12371059 57 311223 47 0 1409 0 22013740 144 418108 52 0 1651 0 21838087 118 331739 49 0 5953 0 22182739 65 294326 45 0 2586 0 24016127 46 272845 30 0 2235 0 22538477 56 252966 28 0 2071 0 21345895 60 288944 27 0 1466 0 685 55 178003 96 1140 473248 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 331486280
btime 1445422082
processes 25280
procs_running 1
procs_blocked 0
softirq 11921298 0 4947902 227750 320262 0 0 7610 3758154 111804 2547816
===Jj52dgpmaF===MemTotal:        1022944 kB
MemFree:          166960 kB
MemAvailable:     805256 kB
Buffers:          329512 kB
Cached:           309956 kB
SwapCached:           20 kB
Active:           342116 kB
Inactive:         424112 kB
Active(anon):     112296 kB
Inactive(anon):    17196 kB
Active(file):     229820 kB
Inactive(file):   406916 kB
Unevictable:           0 kB
Mlocked:               0 kB
HighTotal:        313352 kB
HighFree:         139772 kB
LowTotal:         709592 kB
LowFree:           27188 kB
SwapTotal:        524284 kB
SwapFree:         524200 kB
Dirty:                24 kB
Writeback:             0 kB
AnonPages:        126636 kB
Mapped:            86356 kB
Shmem:              2732 kB
Slab:              29496 kB
SReclaimable:      16572 kB
SUnreclaim:        12924 kB
KernelStack:        2328 kB
PageTables:         1992 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:     1035756 kB
Committed_AS:     941652 kB
VmallocTotal:     122880 kB
VmallocUsed:       10924 kB
VmallocChunk:     103968 kB
DirectMap4k:      735224 kB
DirectMap2M:           0 kB
===Jj52dgpmaF===TcpExt: SyncookiesSent SyncookiesRecv SyncookiesFailed EmbryonicRsts PruneCalled RcvPruned OfoPruned OutOfWindowIcmps LockDroppedIcmps ArpFilter TW TWRecycled TWKilled PAWSPassive PAWSActive PAWSEstab DelayedACKs DelayedACKLocked DelayedACKLost ListenOverflows ListenDrops TCPPrequeued TCPDirectCopyFromBacklog TCPDirectCopyFromPrequeue TCPPrequeueDropped TCPHPHits TCPHPHitsToUser TCPPureAcks TCPHPAcks TCPRenoRecovery TCPSackRecovery TCPSACKReneging TCPFACKReorder TCPSACKReorder TCPRenoReorder TCPTSReorder TCPFullUndo TCPPartialUndo TCPDSACKUndo TCPLossUndo TCPLostRetransmit TCPRenoFailures TCPSackFailures TCPLossFailures TCPFastRetrans TCPForwardRetrans TCPSlowStartRetrans TCPTimeouts TCPLossProbes TCPLossProbeRecovery TCPRenoRecoveryFail TCPSackRecoveryFail TCPSchedulerFailed TCPRcvCollapsed TCPDSACKOldSent TCPDSACKOfoSent TCPDSACKRecv TCPDSACKOfoRecv TCPAbortOnData TCPAbortOnClose TCPAbortOnMemory TCPAbortOnTimeout TCPAbortOnLinger TCPAbortFailed TCPMemoryPressures TCPSACKDiscard TCPDSACKIgnoredOld TCPDSACKIgnoredNoUndo TCPSpuriousRTOs TCPMD5NotFound TCPMD5Unexpected TCPSackShifted TCPSackMerged TCPSackShiftFallback TCPBacklogDrop TCPMinTTLDrop TCPDeferAcceptDrop IPReversePathFilter TCPTimeWaitOverflow TCPReqQFullDoCookies TCPReqQFullDrop TCPRetransFail TCPRcvCoalesce TCPOFOQueue TCPOFODrop TCPOFOMerge TCPChallengeACK TCPSYNChallenge TCPFastOpenActive TCPFastOpenActiveFail TCPFastOpenPassive TCPFastOpenPassiveFail TCPFastOpenListenOverflow TCPFastOpenCookieReqd TCPSpuriousRtxHostQueues BusyPollRxPackets TCPAutoCorking TCPFromZeroWindowAdv TCPToZeroWindowAdv TCPWantZeroWindowAdv TCPSynRetrans TCPOrigDataSent TCPHystartTrainDetect TCPHystartTrainCwnd TCPHystartDelayDetect TCPHystartDelayCwnd TCPACKSkippedSynRecv TCPACKSkippedPAWS TCPACKSkippedSeq TCPACKSkippedFinWait2 TCPACKSkippedTimeWait TCPACKSkippedChallenge
TcpExt: 0 0 0 198 0 0 0 0 0 0 4515 0 0 0 0 1 1056 0 1220 0 0 5 0 5 0 283576 0 51574 3603 2 127 0 0 0 0 13 7 10 6 14 2 1 2 6 338 69 61 366 182 7 1 9 0 0 1219 1 113 1 7 3 0 43 0 0 0 0 0 60 1 0 0 376 622 662 0 0 0 0 0 0 0 0 301406 11407 0 1 61 60 0 0 0 0 0 4 0 0 2754 0 0 0 404 65127 126 2368 12 595 4 0 1 0 0 0
IpExt: InNoRoutes InTruncatedPkts InMcastPkts OutMcastPkts InBcastPkts OutBcastPkts InOctets OutOctets InMcastOctets OutMcastOctets InBcastOctets OutBcastOctets InCsumErrors InNoECTPkts InECT1Pkts InECT0Pkts InCEPkts
IpExt: 0 0 0 0 0 0 515091575 64696238 0 0 0 0 0 512075 0 397 103
===Jj52dgpmaF===0.04 0.03 0.05 1/285 25339
===Jj52dgpmaF===Filesystem       1B-blocks       Used  Available Use% Mounted on
/dev/xvda       5184094208 3416477696 1552871424  69% /
devtmpfs         523460608       4096  523456512   1% /dev
none             104751104     266240  104484864   1% /run
none               5242880          0    5242880   0% /run/lock
none             523747328          0  523747328   0% /run/shm
cgroup           523747328          0  523747328   0% /sys/fs/cgroup
/dev/xvdc      19418279936 9646477312 9771802624  50% /storage
===aRVZeDergP===Filesystem      Inodes  IUsed   IFree IUse% Mounted on
/dev/xvda      1308160 123383 1184777   10% /
devtmpfs        127798   1419  126379    2% /dev
none            127868    796  127072    1% /run
none            127868      3  127865    1% /run/lock
none            127868      1  127867    1% /run/shm
cgroup          127868     12  127856    1% /sys/fs/cgroup
/dev/xvdc      2424832  51487 2373345    3% /storage
`
		TestOutput2 = `634793.22 5077099.55
===Jj52dgpmaF===cpu  42665 5461 18907 507228463 12965 3 452 22329 0 0
cpu0 2572 1692 1704 63401884 7941 0 1 2594 0 0
cpu1 24497 90 7269 63406062 617 2 450 1961 0 0
cpu2 2352 0 1576 63405459 615 0 0 2980 0 0
cpu3 2870 2 1755 63405178 469 0 0 2917 0 0
cpu4 2322 79 1853 63405783 683 0 0 2816 0 0
cpu5 2468 0 1359 63399408 435 0 0 3206 0 0
cpu6 2827 814 1547 63401509 803 0 0 2965 0 0
cpu7 2752 2780 1839 63403175 1398 0 0 2886 0 0
intr 171726338 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 19716144 644 333233 22 0 2139 0 12371264 57 311284 47 0 1419 0 22014017 144 418389 52 0 1660 0 21838261 118 331920 49 0 6072 0 22182896 65 294485 45 0 2600 0 24016341 46 272891 30 0 2248 0 22538586 56 252990 28 0 2074 0 21346032 60 288981 27 0 1468 0 685 55 178005 96 1140 473320 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 331489564
btime 1445422082
processes 25463
procs_running 1
procs_blocked 0
softirq 11923786 0 4948895 227789 320297 0 0 7611 3758466 111804 2548924
===Jj52dgpmaF===MemTotal:        1022944 kB
MemFree:          167960 kB
MemAvailable:     806252 kB
Buffers:          329512 kB
Cached:           309956 kB
SwapCached:           20 kB
Active:           342028 kB
Inactive:         424108 kB
Active(anon):     112204 kB
Inactive(anon):    17196 kB
Active(file):     229824 kB
Inactive(file):   406912 kB
Unevictable:           0 kB
Mlocked:               0 kB
HighTotal:        313352 kB
HighFree:         140052 kB
LowTotal:         709592 kB
LowFree:           27908 kB
SwapTotal:        524284 kB
SwapFree:         524200 kB
Dirty:                44 kB
Writeback:             0 kB
AnonPages:        126700 kB
Mapped:            86472 kB
Shmem:              2732 kB
Slab:              29500 kB
SReclaimable:      16568 kB
SUnreclaim:        12932 kB
KernelStack:        2328 kB
PageTables:         1988 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:     1035756 kB
Committed_AS:     941948 kB
VmallocTotal:     122880 kB
VmallocUsed:       10924 kB
VmallocChunk:     103968 kB
DirectMap4k:      735224 kB
DirectMap2M:           0 kB
===Jj52dgpmaF===TcpExt: SyncookiesSent SyncookiesRecv SyncookiesFailed EmbryonicRsts PruneCalled RcvPruned OfoPruned OutOfWindowIcmps LockDroppedIcmps ArpFilter TW TWRecycled TWKilled PAWSPassive PAWSActive PAWSEstab DelayedACKs DelayedACKLocked DelayedACKLost ListenOverflows ListenDrops TCPPrequeued TCPDirectCopyFromBacklog TCPDirectCopyFromPrequeue TCPPrequeueDropped TCPHPHits TCPHPHitsToUser TCPPureAcks TCPHPAcks TCPRenoRecovery TCPSackRecovery TCPSACKReneging TCPFACKReorder TCPSACKReorder TCPRenoReorder TCPTSReorder TCPFullUndo TCPPartialUndo TCPDSACKUndo TCPLossUndo TCPLostRetransmit TCPRenoFailures TCPSackFailures TCPLossFailures TCPFastRetrans TCPForwardRetrans TCPSlowStartRetrans TCPTimeouts TCPLossProbes TCPLossProbeRecovery TCPRenoRecoveryFail TCPSackRecoveryFail TCPSchedulerFailed TCPRcvCollapsed TCPDSACKOldSent TCPDSACKOfoSent TCPDSACKRecv TCPDSACKOfoRecv TCPAbortOnData TCPAbortOnClose TCPAbortOnMemory TCPAbortOnTimeout TCPAbortOnLinger TCPAbortFailed TCPMemoryPressures TCPSACKDiscard TCPDSACKIgnoredOld TCPDSACKIgnoredNoUndo TCPSpuriousRTOs TCPMD5NotFound TCPMD5Unexpected TCPSackShifted TCPSackMerged TCPSackShiftFallback TCPBacklogDrop TCPMinTTLDrop TCPDeferAcceptDrop IPReversePathFilter TCPTimeWaitOverflow TCPReqQFullDoCookies TCPReqQFullDrop TCPRetransFail TCPRcvCoalesce TCPOFOQueue TCPOFODrop TCPOFOMerge TCPChallengeACK TCPSYNChallenge TCPFastOpenActive TCPFastOpenActiveFail TCPFastOpenPassive TCPFastOpenPassiveFail TCPFastOpenListenOverflow TCPFastOpenCookieReqd TCPSpuriousRtxHostQueues BusyPollRxPackets TCPAutoCorking TCPFromZeroWindowAdv TCPToZeroWindowAdv TCPWantZeroWindowAdv TCPSynRetrans TCPOrigDataSent TCPHystartTrainDetect TCPHystartTrainCwnd TCPHystartDelayDetect TCPHystartDelayCwnd TCPACKSkippedSynRecv TCPACKSkippedPAWS TCPACKSkippedSeq TCPACKSkippedFinWait2 TCPACKSkippedTimeWait TCPACKSkippedChallenge
TcpExt: 0 0 0 198 0 0 0 0 0 0 4515 0 0 0 0 2 1058 0 1224 0 0 5 0 5 0 283578 0 51593 3603 2 127 0 0 0 0 13 7 10 6 14 2 1 2 6 338 69 61 366 182 7 1 9 0 0 1223 1 113 1 7 3 0 43 0 0 0 0 0 60 1 0 0 376 622 662 0 0 0 0 0 0 0 0 301411 11410 0 1 61 60 0 0 0 0 0 4 0 0 2755 0 0 0 404 65148 126 2368 12 595 4 0 1 0 0 0
IpExt: InNoRoutes InTruncatedPkts InMcastPkts OutMcastPkts InBcastPkts OutBcastPkts InOctets OutOctets InMcastOctets OutMcastOctets InBcastOctets OutBcastOctets InCsumErrors InNoECTPkts InECT1Pkts InECT0Pkts InCEPkts
IpExt: 0 0 0 0 0 0 515097425 64704555 0 0 0 0 0 512122 0 397 103
===Jj52dgpmaF===0.03 0.03 0.05 1/285 25522
===Jj52dgpmaF===Filesystem       1B-blocks       Used  Available Use% Mounted on
/dev/xvda       5184094208 3416477696 1552871424  69% /
devtmpfs         523460608       4096  523456512   1% /dev
none             104751104     266240  104484864   1% /run
none               5242880          0    5242880   0% /run/lock
none             523747328          0  523747328   0% /run/shm
cgroup           523747328          0  523747328   0% /sys/fs/cgroup
/dev/xvdc      19418279936 9646477312 9771802624  50% /storage
===aRVZeDergP===Filesystem      Inodes  IUsed   IFree IUse% Mounted on
/dev/xvda      1308160 123383 1184777   10% /
devtmpfs        127798   1419  126379    2% /dev
none            127868    796  127072    1% /run
none            127868      3  127865    1% /run/lock
none            127868      1  127867    1% /run/shm
cgroup          127868     12  127856    1% /sys/fs/cgroup
/dev/xvdc      2424832  51487 2373345    3% /storage
`
	)

	var req_num int
	s := test.NewTestSSHExecServer(t, NewSSHGzipHandler(t, func(req *ssh.Request, input []byte) []byte {

		reqCommand := string(input)

		if reqCommand != TestCommand {
			t.Errorf("Expected %q, got %q", TestCommand, reqCommand)
		}

		req_num++
		switch req_num {
		case 1:
			return []byte(TestOutput1)
		case 2:
			return []byte(TestOutput2)
		default:
			t.Errorf("unexpected request #%d", req_num)
		}

		return nil

	}))
	defer s.Close()

	sc := NewSystemInfoCollector(s.Host, &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, 2)

	if err := sc.Poll(); err != nil {
		t.Error(err)
	}

	if err := sc.Poll(); err != nil {
		t.Error(err)
	}

	summary := sc.GetSummary()

	// pretty.Println(summary)

	// zero out timestamps because they will never match
	summary.Timestamp = time.Time{}

	expected := &SystemInfoSummary{
		Timestamp: time.Time{},
		Duration:  2140000000,
		Memory: struct {
			Total         uint64
			System        uint64
			User          uint64
			PercentSystem float64
			PercentUser   float64
		}{Total: 0x3e6f8000, System: 0x2707b000, User: 0xd277000, PercentSystem: 0.6251251290393218, PercentUser: 0.21068210967560297},
		Swap: struct {
			Total        uint64
			Free         uint64
			PercentInUse float64
		}{Total: 0x1ffff000, Free: 0x1ffe5000, PercentInUse: 0.00019836577122323007},
		Disk: Disk{Filesystem: "/dev/xvda", MountPoint: "/", BytesTotal: 0x134ff0000, BytesUsed: 0xcba35000, BytesAvailable: 0x5c8ef000, PercentInUse: 0.6590307889713412, InodesTotal: 0x13f600, InodesUsed: 0x1e1f7, InodesAvailable: 0x121409, PercentInodesInUse: 0.09431797333659492},
		CPU: struct {
			CPUSummary
			ContextSwitches  uint64
			Processes        uint64
			ProcessesRunning uint64
			ProcessesBlocked uint64
			CPUs             []CPUSummary
		}{
			CPUSummary:       CPUSummary{Id: "cpu", PercentInUse: 0.036214953271028034, PercentUser: 0.013434579439252336, PercentNice: 0, PercentSystem: 0.0227803738317757, PercentIdle: 0.9637850467289719, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
			ContextSwitches:  0xcd4,
			Processes:        0xb7,
			ProcessesRunning: 0x1,
			ProcessesBlocked: 0x0,
			CPUs: []CPUSummary{
				{Id: "cpu0", PercentInUse: 0.004672897196261682, PercentUser: 0, PercentNice: 0, PercentSystem: 0.004672897196261682, PercentIdle: 0.9953271028037384, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
				{Id: "cpu1", PercentInUse: 0.009478672985781991, PercentUser: 0.004739336492890996, PercentNice: 0, PercentSystem: 0.004739336492890996, PercentIdle: 0.990521327014218, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
				{Id: "cpu2", PercentInUse: 0.02336448598130841, PercentUser: 0.004672897196261682, PercentNice: 0, PercentSystem: 0.018691588785046728, PercentIdle: 0.9766355140186916, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
				{Id: "cpu3", PercentInUse: 0.1261682242990654, PercentUser: 0.06074766355140187, PercentNice: 0, PercentSystem: 0.06542056074766354, PercentIdle: 0.8738317757009346, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
				{Id: "cpu4", PercentInUse: 0.07981220657276995, PercentUser: 0.023474178403755867, PercentNice: 0, PercentSystem: 0.056338028169014086, PercentIdle: 0.92018779342723, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
				{Id: "cpu5", PercentInUse: 0.004694835680751174, PercentUser: 0, PercentNice: 0, PercentSystem: 0.004694835680751174, PercentIdle: 0.9953051643192489, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
				{Id: "cpu6", PercentInUse: 0.018604651162790697, PercentUser: 0.004651162790697674, PercentNice: 0, PercentSystem: 0.013953488372093023, PercentIdle: 0.9813953488372092, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
				{Id: "cpu7", PercentInUse: 0.009302325581395349, PercentUser: 0.004651162790697674, PercentNice: 0, PercentSystem: 0.004651162790697674, PercentIdle: 0.9906976744186047, PercentIOWait: 0, PercentIRQ: 0, PercentSoftIRQ: 0, PercentSteal: 0, PercentGuest: 0, PercentGuestNice: 0},
			},
		},
		Network: struct {
			BytesIn           uint64
			BytesOut          uint64
			BytesPerSecondIn  uint64
			BytesPerSecondOut uint64
		}{BytesIn: 0x16da, BytesOut: 0x207d, BytesPerSecondIn: 0xaad, BytesPerSecondOut: 0xf2e},
	}

	if !reflect.DeepEqual(summary, expected) {
		t.Error("summary mismatch")
		dumpDiff(expected, summary)
	}

}

func dumpDiff(expected, actual interface{}) {

	expjson, err := json.Marshal(expected)
	if err != nil {
		log.Fatal(err)
	}

	actualjson, err := json.Marshal(actual)
	if err != nil {
		log.Fatal(err)
	}

	d := gojsondiff.New()
	dd, err := d.Compare(expjson, actualjson)
	if err != nil {
		panic(err)
	}

	left := make(map[string]interface{})
	if err := json.Unmarshal(expjson, &left); err != nil {
		panic(err)
	}
	ff := formatter.NewAsciiFormatter(left, formatter.AsciiFormatterConfig{Coloring: true})
	out, err := ff.Format(dd)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

}
