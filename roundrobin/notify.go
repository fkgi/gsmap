package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fkgi/gsmap/tcap"
	"github.com/fkgi/gsmap/xua"
)

var traceTxDialog = func(n, v string) {
	log.Println("[INFO]", "Tx new dialog", n, v)
}

var traceRxDialog = func(n, v string) {
	log.Println("[INFO]", "Rx new dialog", n, v)
}

var traceTxHttpRequest = func(
	p string, tx []byte, c int, rx []byte, e error) {
	traceHTTPHandling("Tx", p, tx, c, rx, e)
}
var traceRxHttpRequest = func(
	p string, tx []byte, c int, rx []byte, e error) {
	traceHTTPHandling("Rx", p, tx, c, rx, e)
}

func init() {
	xua.TraceEvent = func(old, new, event string, e error) {
		log.Printf("[INFO] ASP state update: %s->%s by event %s, error=%v",
			old, new, event, e)
	}
	xua.AsStateNotify = func(s string) {
		log.Println("[INFO]", "AS state update:", s)
	}
	tcap.TraceRxMessage = func(m tcap.Message, e error) {
		log.Printf("[INFO] Rx MAP message handling: error=%v\n%s", e, m.String())
		count(m, true)
	}
	tcap.TraceTxMessage = func(m tcap.Message, e error) {
		log.Printf("[INFO] Tx MAP message handling: error=%v\n%s", e, m.String())
		count(m, false)
	}

	xua.DunaNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Rx DUNA for PC=%v", pc)
	}
	xua.DavaNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Rx DAVA for PC=%v", pc)
	}
	xua.DaudNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Tx DAUD for PC=%v", pc)
	}
	xua.SconNotify = func(pc []xua.PointCode, con uint32) {
		log.Printf("[INFO] Rx SCON for PC=%v, congestion level=%d", pc, con)
	}
	xua.DupuNotify = func(pc []xua.PointCode, cause uint16) {
		log.Printf("[INFO] Rx DUPU for PC=%v, cause=%d", pc, cause)
	}
	xua.DrstNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Rx DRST for PC=%v", pc)
	}
}

func traceHTTPHandling(d string, p string, tx []byte, c int, rx []byte, e error) {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s HTTP request handling: error=%v\n", d, e)
	fmt.Fprintln(buf, "| path:  ", p)
	fmt.Fprintln(buf, "| body:  ", strings.TrimSpace(string(tx)))
	fmt.Fprintln(buf, "| result:", c, http.StatusText(c))
	fmt.Fprintln(buf, "| body:  ", strings.TrimSpace(string(rx)))
	log.Print("[INFO] ", buf)
}
