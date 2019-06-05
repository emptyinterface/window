package sysinfo

import (
	"sync"
	"time"
)

type (
	StatSeries struct {
		stats   []*Stat
		stats_i int
		me      sync.Mutex
	}
)

func NewStatSeries(n int) *StatSeries {
	return &StatSeries{
		stats: make([]*Stat, n),
		me:    sync.Mutex{},
	}
}

func (ss *StatSeries) Size() int {
	ss.me.Lock()
	defer ss.me.Unlock()
	return len(ss.stats)
}

func (ss *StatSeries) Len() int {
	ss.me.Lock()
	defer ss.me.Unlock()
	if ss.stats_i < len(ss.stats) {
		return ss.stats_i
	}
	return len(ss.stats)
}

func (ss *StatSeries) Add(stat *Stat) {
	ss.me.Lock()
	defer ss.me.Unlock()
	stat.Duration = stat.UpTime.Total
	ss.stats[ss.stats_i%len(ss.stats)] = stat
	ss.stats_i++
}

func (ss *StatSeries) Since(ts time.Time) []*Stat {

	var set []*Stat

	ss.me.Lock()
	defer ss.me.Unlock()

	for i, l := ss.stats_i, ss.stats_i+len(ss.stats); i < l; i++ {
		stat := ss.stats[i%len(ss.stats)]
		if stat != nil && stat.Timestamp.After(ts) {
			set = append(set, stat)
		}
	}

	return set

}

// return stats in order 0 == oldest, len == newest
func (ss *StatSeries) Last(n int) []*Stat {

	var set []*Stat

	ss.me.Lock()
	defer ss.me.Unlock()

	// if we still haven't wrapped around just copy out the stats
	if n > ss.stats_i {
		set = append(set, ss.stats[:ss.stats_i]...)
	} else {
		for i := ss.stats_i - n; i < ss.stats_i; i++ {
			set = append(set, ss.stats[i%len(ss.stats)])
		}
	}

	return set

}
