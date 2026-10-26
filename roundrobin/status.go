package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/fkgi/gsmap/tcap"
	"github.com/fkgi/gsmap/xua"
)

var rxMetrics = make(chan map[string]uint64, 1)
var txMetrics = make(chan map[string]uint64, 1)

func init() {
	rxMetrics <- map[string]uint64{}
	txMetrics <- map[string]uint64{}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	sysmetricsHandler(w)

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Fprintln(w, "# HELP heap_alloc_bytes Number of heap bytes allocated and still in use")
	fmt.Fprintln(w, "# TYPE heap_alloc_bytes gauge")
	fmt.Fprintf(w, "heap_alloc_bytes %d\n", ms.Alloc)

	fmt.Fprintln(w, "# HELP roundrobin_active_worker_count Count of active worker")
	fmt.Fprintln(w, "# TYPE roundrobin_active_worker gauge")
	fmt.Fprintf(w, "roundrobin_active_worker %d\n", xua.ActiveWorkerCount())

	fmt.Fprintln(w, "# HELP roundrobin_as_state Status of AS(down=0, inactive=1, active=2)")
	fmt.Fprintln(w, "# TYPE roundrobin_as_state gauge")
	switch tcap.EndPoint.State() {
	case "down":
		fmt.Fprintln(w, "roundrobin_as_state 0")
	case "inactive":
		fmt.Fprintln(w, "roundrobin_as_state 1")
	case "active":
		fmt.Fprintln(w, "roundrobin_as_state 2")
	}

	asps := tcap.EndPoint.ListASPs()
	if len(asps) != 0 {
		fmt.Fprintln(w, "# HELP roundrobin_asp_state Status of ASP(down=0, inactive=1, active=2)")
		fmt.Fprintln(w, "# TYPE roundrobin_asp_state gauge")
	}
	for _, a := range asps {
		switch a.State() {
		case "down":
			fmt.Fprintf(w, "roundrobin_asp_state{address=\"%s\"} 0\n", a.RemoteAddr())
		case "inactive":
			fmt.Fprintf(w, "roundrobin_asp_state{address=\"%s\"} 1\n", a.RemoteAddr())
		case "active":
			fmt.Fprintf(w, "roundrobin_asp_state{address=\"%s\"} 2\n", a.RemoteAddr())
		}
	}

	m := <-txMetrics
	if len(m) != 0 {
		fmt.Fprintln(w, "# HELP roundrobin_tx_msu Total number of outgoing MSU")
		fmt.Fprintln(w, "# TYPE roundrobin_tx_msu counter")
	}
	for k, v := range m {
		fmt.Fprintf(w, "roundrobin_tx_msu{message=\"%s\"} %d\n", k, v)
	}
	txMetrics <- m

	m = <-rxMetrics
	if len(m) != 0 {
		fmt.Fprintln(w, "# HELP roundrobin_rx_msu Total number of incoming MSU")
		fmt.Fprintln(w, "# TYPE roundrobin_rx_msu counter")
	}
	for k, v := range m {
		fmt.Fprintf(w, "roundrobin_rx_msu{message=\"%s\"} %d\n", k, v)
	}
	rxMetrics <- m
}

func count(m tcap.Message, rx bool) {
	switch msg := m.(type) {
	case *tcap.TcBegin, *tcap.TcContinue, *tcap.TcEnd:
		cs := msg.Components()
		if len(cs) == 0 {
			n := ""
			switch m.(type) {
			case *tcap.TcBegin:
				n = "EmptyBegin"
			case *tcap.TcContinue:
				n = "EmptyContinue"
			case *tcap.TcEnd:
				n = "EmptyEnd"
			}
			if rx {
				mx := <-rxMetrics
				mx[n] = mx[n] + 1
				rxMetrics <- mx
			} else {
				mx := <-txMetrics
				mx[n] = mx[n] + 1
				txMetrics <- mx
			}
		}
		for _, c := range cs {
			n := c.Name()
			if rx {
				mx := <-rxMetrics
				mx[n] = mx[n] + 1
				rxMetrics <- mx
			} else {
				mx := <-txMetrics
				mx[n] = mx[n] + 1
				txMetrics <- mx
			}
		}
	case *tcap.TcAbort:
		n := "pAbort"
		if _, d := msg.Cause(); d != nil {
			n = "uAbort"
		}
		if rx {
			mx := <-rxMetrics
			mx[n] = mx[n] + 1
			rxMetrics <- mx
		} else {
			mx := <-txMetrics
			mx[n] = mx[n] + 1
			txMetrics <- mx
		}
	}
}
