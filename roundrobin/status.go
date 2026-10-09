package main

import (
	"fmt"
	"net/http"
	"runtime"
	"syscall"

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

	var usage syscall.Rusage
	if e := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); e == nil {
		fmt.Fprintln(w, "# HELP cpu_seconds_total Total user and system CPU time spent in seconds")
		fmt.Fprintln(w, "# TYPE cpu_seconds_total counter")
		fmt.Fprintf(w, "cpu_seconds_total %.4f\n",
			float64(usage.Utime.Sec)+float64(usage.Utime.Usec)/1e6+
				float64(usage.Stime.Sec)+float64(usage.Stime.Usec)/1e6)
	}

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
	}

	m := <-txMetrics
	if len(m) != 0 {
		fmt.Fprintln(w, "# HELP roundrobin_tx_msu Total number of outgoing MSU")
		fmt.Fprintln(w, "# TYPE roundrobin_tx_msu counter")
		for k, v := range m {
			fmt.Fprintf(w, "roundrobin_tx_msu{message=\"%s\"} %d\n", k, v)
		}
	}
	txMetrics <- m
	m = <-rxMetrics
	if len(m) != 0 {
		fmt.Fprintln(w, "# HELP roundrobin_rx_msu Total number of outgoing MSU")
		fmt.Fprintln(w, "# TYPE roundrobin_rx_msu counter")
		for k, v := range m {
			fmt.Fprintf(w, "roundrobin_rx_msu{message=\"%s\"} %d\n", k, v)
		}
	}
	rxMetrics <- m
}

/*
type statFmt struct {
	S string    `json:"roundrobin_as_state"`
	P []peerFmt `json:"peer"`
}
type peerFmt struct {
	S string `json:"state,omitempty"`
	A string `json:"address"`
}

func conStateHandler(w http.ResponseWriter, r *http.Request) {
	asps := tcap.EndPoint.ListASPs()
	if len(asps) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	st := statFmt{
		S: tcap.EndPoint.State(),
		P: []peerFmt{}}
	for _, a := range asps {
		st.P = append(st.P, peerFmt{
			S: a.State(),
			A: a.RemoteAddr().String(),
		})
	}

	if jd, e := json.Marshal(st); e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jd)
	}
}

type statistics struct {
	RxInvoke     uint64 `json:"rx_invoke"`
	TxInvoke     uint64 `json:"tx_invoke"`
	RxResult     uint64 `json:"rx_result"`
	TxResult     uint64 `json:"tx_result"`
	RxResultLast uint64 `json:"rx_resultlast"`
	TxResultLast uint64 `json:"tx_resultlast"`
	RxError      uint64 `json:"rx_error"`
	TxError      uint64 `json:"tx_error"`
	RxAbort      uint64 `json:"rx_abort"`
	TxAbort      uint64 `json:"tx_abort"`
}

var stats = make(chan statistics, 1)

func init() {
	stats <- statistics{}
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	s := <-stats
	stats <- s

	if jd, e := json.Marshal(s); e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jd)
	}
}
*/

func count(m tcap.Message, rx bool) {
	// s := <-stats
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
			/*
				switch c.(type) {
				case gsmap.Invoke:
					if rx {
						s.RxInvoke++
					} else {
						s.TxInvoke++
					}
				case gsmap.ReturnResult:
					if rx {
						s.RxResult++
					} else {
						s.TxResult++
					}
				case gsmap.ReturnResultLast:
					if rx {
						s.RxResultLast++
					} else {
						s.TxResultLast++
					}
				case gsmap.ReturnError:
					if rx {
						s.RxError++
					} else {
						s.TxError++
					}
				}
			*/
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
		/*
			if rx {
				s.RxAbort++
			} else {
				s.TxAbort++
			}
		*/
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
	// stats <- s
}
