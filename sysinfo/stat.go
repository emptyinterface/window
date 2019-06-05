package sysinfo

import "time"

type (
	Stat struct {
		Timestamp time.Time
		Duration  time.Duration
		UpTime    UpTime
		CPUInfo   CPUInfo
		MemInfo   MemInfo
		NetStat   NetStat
		LoadAvg   LoadAvg
		DiskInfo  DiskInfo
	}
)

func NewStat() *Stat {
	return &Stat{}
}
