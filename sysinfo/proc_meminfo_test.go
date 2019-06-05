package sysinfo

import (
	"reflect"
	"testing"
)

func TestParseMemInfo(t *testing.T) {

	const output = `MemTotal:        1016652 kB
MemFree:          330788 kB
MemAvailable:     638772 kB
Buffers:             324 kB
Cached:           441384 kB
SwapCached:            0 kB
Active:           331508 kB
Inactive:         214468 kB
Active(anon):     122820 kB
Inactive(anon):    39184 kB
Active(file):     208688 kB
Inactive(file):   175284 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:             0 kB
SwapFree:              0 kB
Dirty:                 4 kB
Writeback:             0 kB
AnonPages:        104300 kB
Mapped:            18032 kB
Shmem:             57736 kB
Slab:             106816 kB
SReclaimable:      73272 kB
SUnreclaim:        33544 kB
KernelStack:        1040 kB
PageTables:         3608 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:      508324 kB
Committed_AS:     266220 kB
VmallocTotal:   34359738367 kB
VmallocUsed:        3492 kB
VmallocChunk:   34359734271 kB
HardwareCorrupted:     0 kB
AnonHugePages:     65536 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:       45056 kB
DirectMap2M:     1003520 kB
`

	expected := MemInfo{MemTotal: 0x3e0d3000, MemFree: 0x14309000, MemAvailable: 0x26fcd000, Buffers: 0x51000, Cached: 0x1af0a000, SwapCached: 0x0, Active: 0x143bd000, Inactive: 0xd171000, ActiveAnon: 0x77f1000, InactiveAnon: 0x2644000, ActiveFile: 0xcbcc000, InactiveFile: 0xab2d000, Unevictable: 0x0, Mlocked: 0x0, SwapTotal: 0x0, SwapFree: 0x0, Dirty: 0x1000, Writeback: 0x0, AnonPages: 0x65db000, Mapped: 0x119c000, Shmem: 0x3862000, Slab: 0x6850000, SReclaimable: 0x478e000, SUnreclaim: 0x20c2000, KernelStack: 0x104000, PageTables: 0x386000, NFS_Unstable: 0x0, Bounce: 0x0, WritebackTmp: 0x0, CommitLimit: 0x1f069000, Committed_AS: 0x103fb000, VmallocTotal: 0x1ffffffffc00, VmallocUsed: 0x369000, VmallocChunk: 0x1fffffbffc00, HardwareCorrupted: 0x0, AnonHugePages: 0x4000000, HugePages_Total: 0x0, HugePages_Free: 0x0, HugePages_Rsvd: 0x0, HugePages_Surp: 0x0, Hugepagesize: 0x200000, DirectMap4k: 0x2c00000, DirectMap2M: 0x3d400000, DirectMap1G: 0x0}

	stat := NewStat()

	if err := (&stat.MemInfo).Parse([]byte(output)); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stat.MemInfo, expected) {
		t.Error("parse mismatch")
		dumpDiff(expected, stat.MemInfo)
	}
}
