package sysinfo

import (
	"reflect"
	"testing"
)

func TestParseUpTime(t *testing.T) {

	const output = `2686521.22 2653557.24`

	expected := UpTime{Total: 2686521220000000, Idle: 2653557240000000}

	stat := NewStat()

	if err := (&stat.UpTime).Parse([]byte(output)); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stat.UpTime, expected) {
		t.Error("parse mismatch")
		dumpDiff(expected, stat.UpTime)
	}
}
