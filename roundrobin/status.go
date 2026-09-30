package main

import (
	"encoding/json"
	"net/http"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/tcap"
)

type statFmt struct {
	S string    `json:"as_state"`
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

func count(m tcap.Message, rx bool) {
	s := <-stats

	switch msg := m.(type) {
	case *tcap.TcBegin, *tcap.TcContinue, *tcap.TcEnd:
		for _, c := range msg.Components() {
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
		}
	case *tcap.TcAbort:
		if rx {
			s.RxAbort++
		} else {
			s.TxAbort++
		}
	}
	stats <- s
}
