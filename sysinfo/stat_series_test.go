package sysinfo

import (
	"testing"
	"time"
)

func TestStatSeries(t *testing.T) {

	series := NewStatSeries(10)
	now := time.Now()

	for i := 0; i < series.Size(); i++ {
		stat := NewStat()
		stat.Timestamp = now.Add(-time.Duration(series.Size()-i) * time.Minute)
		series.Add(stat)

	}

	last5min := series.Since(now.Add(-5 * time.Minute))

	if len(last5min) != 4 {
		t.Errorf("Expected 4, got %d", len(last5min))
	}

	maxage := now.Add(-5 * time.Minute)

	for _, stat := range last5min {
		if stat.Timestamp.Before(maxage) {
			t.Errorf("expected events after %s, found %s", maxage, stat.Timestamp)
		}
	}

	last5 := series.Last(5)

	if len(last5) != 5 {
		t.Errorf("Expected 5, got %d", len(last5))
	}

	for _, stat := range last5 {
		if stat.Timestamp.Before(maxage) {
			t.Errorf("expected events after %s, found %s", maxage, stat.Timestamp)
		}
	}

}
