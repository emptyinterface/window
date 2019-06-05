package sysinfo

import (
	"bufio"
	"bytes"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type (
	MemInfo struct {
		MemTotal          uint64
		MemFree           uint64
		MemAvailable      uint64
		Buffers           uint64
		Cached            uint64
		SwapCached        uint64
		Active            uint64
		Inactive          uint64
		ActiveAnon        uint64
		InactiveAnon      uint64
		ActiveFile        uint64
		InactiveFile      uint64
		Unevictable       uint64
		Mlocked           uint64
		SwapTotal         uint64
		SwapFree          uint64
		Dirty             uint64
		Writeback         uint64
		AnonPages         uint64
		Mapped            uint64
		Shmem             uint64
		Slab              uint64
		SReclaimable      uint64
		SUnreclaim        uint64
		KernelStack       uint64
		PageTables        uint64
		NFS_Unstable      uint64
		Bounce            uint64
		WritebackTmp      uint64
		CommitLimit       uint64
		Committed_AS      uint64
		VmallocTotal      uint64
		VmallocUsed       uint64
		VmallocChunk      uint64
		HardwareCorrupted uint64
		AnonHugePages     uint64
		HugePages_Total   uint64
		HugePages_Free    uint64
		HugePages_Rsvd    uint64
		HugePages_Surp    uint64
		Hugepagesize      uint64
		DirectMap4k       uint64
		DirectMap2M       uint64
		DirectMap1G       uint64
	}
)

const MemInfoCommand = `cat /proc/meminfo`

func (_ *MemInfo) Command() string {
	return MemInfoCommand
}

func (m *MemInfo) Parse(b []byte) error {

	// output:
	// MemTotal:        1016652 kB
	// MemFree:          330788 kB
	// MemAvailable:     638772 kB
	// Buffers:             324 kB
	// Cached:           441384 kB
	// SwapCached:            0 kB
	// Active:           331508 kB
	// Inactive:         214468 kB
	// Active(anon):     122820 kB
	// Inactive(anon):    39184 kB
	// Active(file):     208688 kB
	// Inactive(file):   175284 kB
	// Unevictable:           0 kB
	// Mlocked:               0 kB
	// SwapTotal:             0 kB
	// SwapFree:              0 kB
	// Dirty:                 4 kB
	// Writeback:             0 kB
	// AnonPages:        104300 kB
	// Mapped:            18032 kB
	// Shmem:             57736 kB
	// Slab:             106816 kB
	// SReclaimable:      73272 kB
	// SUnreclaim:        33544 kB
	// KernelStack:        1040 kB
	// PageTables:         3608 kB
	// NFS_Unstable:          0 kB
	// Bounce:                0 kB
	// WritebackTmp:          0 kB
	// CommitLimit:      508324 kB
	// Committed_AS:     266220 kB
	// VmallocTotal:   34359738367 kB
	// VmallocUsed:        3492 kB
	// VmallocChunk:   34359734271 kB
	// HardwareCorrupted:     0 kB
	// AnonHugePages:     65536 kB
	// HugePages_Total:       0
	// HugePages_Free:        0
	// HugePages_Rsvd:        0
	// HugePages_Surp:        0
	// Hugepagesize:       2048 kB
	// DirectMap4k:       45056 kB
	// DirectMap2M:     1003520 kB

	v := reflect.ValueOf(m).Elem()
	IsSpaceOrColon := func(r rune) bool { return unicode.IsSpace(r) || r == ':' }
	s := bufio.NewScanner(bytes.NewReader(b))

	for s.Scan() {

		fields := strings.FieldsFunc(s.Text(), IsSpaceOrColon)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		value, _ := strconv.ParseUint(fields[1], 10, 64)

		// store value in bytes
		if len(fields) == 3 && strings.ToLower(fields[2]) == "kb" {
			value *= 1024
		}

		switch name {
		case "Active(anon)":
			v.FieldByName("ActiveAnon").SetUint(value)
		case "Inactive(anon)":
			v.FieldByName("InactiveAnon").SetUint(value)
		case "Active(file)":
			v.FieldByName("ActiveFile").SetUint(value)
		case "Inactive(file)":
			v.FieldByName("InactiveFile").SetUint(value)
		default:
			if f := v.FieldByName(fields[0]); f.CanSet() {
				f.SetUint(value)
			}
		}

	}

	return s.Err()

}
