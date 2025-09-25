package main

import (
	"fmt"
	"net/http"

	"github.com/fkgi/gsmap/xua"
)

const constatFmt = `{
	"state": "%s",
	"local": {
		"address": "%s",
		"gt": "%s"
	},
	"peer": {
		"address": "%s"
	}
}`

func conStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(constatFmt,
		asp.State(), asp.LocalAddr(), xua.LocalAddr.GlobalTitle, asp.RemoteAddr())))
}

var (
	rxInvoke     uint64
	txInvoke     uint64
	rxResult     uint64
	txResult     uint64
	rxResultLast uint64
	txResultLast uint64
	rxError      uint64
	txError      uint64
	rxAbort      uint64
	txAbort      uint64
)

const statsFmt = `{
	"rx_invoke": %d,
	"tx_result": %d,
	"tx_resultlast": %d,
	"tx_error": %d,
	"tx_abort": %d,
	"tx_invoke": %d,
	"rx_result": %d,
	"rx_reusltlast": %d,
	"rx_error": %d,
	"rx_abort": %d
}`

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(statsFmt,
		rxInvoke, txResult, txResultLast, txError, txAbort,
		txInvoke, rxResult, rxResultLast, rxError, rxAbort)))
}
