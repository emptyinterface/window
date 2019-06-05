package sysinfo

import (
	"reflect"
	"testing"
)

func TestParseLoadAvg(t *testing.T) {

	const output = `0.00 0.03 0.05 1/117 16915`

	expected := LoadAvg{Last1Min: 0, Last5Min: 0.03, Last15Min: 0.05, ProcessRunning: 0x1, ProcessTotal: 0x75, LastPID: 0x4213}

	stat := NewStat()

	if err := (&stat.LoadAvg).Parse([]byte(output)); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stat.LoadAvg, expected) {
		t.Error("parse mismatch")
		dumpDiff(expected, stat.LoadAvg)
	}

}
