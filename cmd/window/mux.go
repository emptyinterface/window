package main

import (
	"net/http"
	"sync/atomic"
	"time"
)

type (
	Mux struct {
		*http.ServeMux
		refresh    func()
		refreshing int64
	}
)

func NewMux(refresh func()) *Mux {
	return &Mux{
		ServeMux: http.NewServeMux(),
		refresh:  refresh,
	}
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if atomic.CompareAndSwapInt64(&m.refreshing, 0, 1) {
		m.refresh()
		go func() {
			time.Sleep(2 * time.Second)
			atomic.StoreInt64(&m.refreshing, 0)
		}()
	}
	m.ServeMux.ServeHTTP(w, r)
}
