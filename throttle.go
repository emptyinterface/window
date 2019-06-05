package window

import (
	"fmt"
	"time"
)

type (
	throttle struct {
		concurrent chan struct{}
		rate       chan struct{}
		ticker     *time.Ticker
	}
)

func NewThrottle(concurrent int, rate int, interval time.Duration) *throttle {

	t := &throttle{
		concurrent: make(chan struct{}, concurrent),
		rate:       make(chan struct{}, rate),
		ticker:     time.NewTicker(interval),
	}

	go func() {
		for range t.ticker.C {
			for i := 0; i < cap(t.rate); i++ {
				select {
				case <-t.rate:
				default:
					break
				}
			}
		}
	}()

	return t

}

func (t *throttle) do(name string, f func() error) chan error {

	errchan := make(chan error, 1)

	go func() {
		var err error
		t.concurrent <- struct{}{}
		t.rate <- struct{}{}
		defer func() {
			<-t.concurrent
			// if e := recover(); e != nil {
			// 	errchan <- fmt.Errorf("%s panic: %v", name, e)
			// } else
			if err != nil {
				errchan <- fmt.Errorf("%s error: %v", name, err)
			} else {
				errchan <- nil
			}
		}()
		// start := time.Now()
		err = f()
		// fmt.Fprintf(os.Stderr, "%s completed in %s\n", name, time.Since(start))
	}()

	return errchan

}

func (t *throttle) stop() {
	t.ticker.Stop()
}
