package sysinfo

import (
	"bytes"
	"fmt"
	"time"
)

type (
	UpTime struct {
		Total time.Duration
		Idle  time.Duration
	}
)

const UpTimeCommand = `cat /proc/uptime`

func (_ *UpTime) Command() string {
	return UpTimeCommand
}

func (u *UpTime) Parse(b []byte) error {

	// output:
	// 2686521.22 2653557.24

	var total, idle float64
	if _, err := fmt.Fscan(bytes.NewReader(b), &total, &idle); err != nil {
		return err
	}

	u.Total = time.Duration(total * float64(time.Second))
	u.Idle = time.Duration(idle * float64(time.Second))

	return nil

}
