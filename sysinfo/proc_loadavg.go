package sysinfo

import (
	"bytes"
	"fmt"
)

type (
	LoadAvg struct {
		Last1Min       float64
		Last5Min       float64
		Last15Min      float64
		ProcessRunning uint64
		ProcessTotal   uint64
		LastPID        uint64
	}
)

const loadAvgCommand = `cat /proc/loadavg`

func (_ *LoadAvg) Command() string {
	return loadAvgCommand
}

func (la *LoadAvg) Parse(b []byte) error {

	// output:
	// 0.00 0.03 0.05 1/117 16915

	_, err := fmt.Fscanf(bytes.NewReader(b), "%f %f %f %d/%d %d",
		&la.Last1Min,
		&la.Last5Min,
		&la.Last15Min,
		&la.ProcessRunning,
		&la.ProcessTotal,
		&la.LastPID,
	)

	return err

}
