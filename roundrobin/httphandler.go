package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/tcap"
)

func handleOutgoingDialog(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r.PathValue("ac"), r.PathValue("ver"))
	if ctx == 0 {
		httpErr(r.URL.Path, nil, http.StatusNotFound, "unsupported context",
			fmt.Sprintf("context=%s, version=%s", r.PathValue("ac"), r.PathValue("ver")), w)
		return
	}

	txjson, e := io.ReadAll(r.Body)
	defer r.Body.Close()
	if e != nil {
		httpErr(r.URL.Path, nil, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
		return
	}
	if txjson, e = compact(txjson); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
		return
	}
	cdpa, _, cp, e := readFromJSON(txjson, 1)
	if e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", e.Error(), w)
		return
	}

	n, v := getContextName(ctx)
	traceRxDalog(n, v)

	var t *tcap.Transaction
	if t, cp, e = tcap.DialTC(ctx, cdpa, cp...); e != nil && e != io.EOF {
		httpErr(r.URL.Path, txjson, http.StatusInternalServerError,
			"failed to dial TC", e.Error(), w)
		return
	}

	if len(cp) != 0 {
		if iv, ok := cp[0].(gsmap.Invoke); ok {
			t.LastInvokeID = iv.GetInvokeID()
		}
	}

	var id string
	if e != io.EOF {
		tid := t.GetIdentity()
		id = hex.EncodeToString([]byte{
			byte(tid >> 24), byte(tid >> 16), byte(tid >> 8), byte(tid)})
	}

	rxjson, e := writeToJSON(nil, &t.CdPA, cp)
	if e != nil {
		httpErr(r.URL.Path, txjson, http.StatusInternalServerError,
			"unable to marshal to JSON", e.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if id != "" {
		w.Header().Set("Location", "/dialog/"+id)
		w.WriteHeader(http.StatusCreated)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusCreated, rxjson, nil)
	} else {
		w.WriteHeader(http.StatusOK)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusOK, rxjson, nil)
	}
	w.Write(rxjson)
}

func handleContinueDialog(w http.ResponseWriter, r *http.Request) {
	var t *tcap.Transaction
	if b, e := hex.DecodeString(r.PathValue("id")); e != nil {
		w.WriteHeader(http.StatusNotFound)
		traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
		return
	} else if len(b) != 4 {
		w.WriteHeader(http.StatusNotFound)
		traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
		return
	} else if t = tcap.GetTransaction(
		(uint32(b[0]) << 24) | (uint32(b[1]) << 16) |
			(uint32(b[2]) << 8) | (uint32(b[3]))); t == nil {
		w.WriteHeader(http.StatusNotFound)
		traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
		return
	}

	txjson, e := io.ReadAll(r.Body)
	defer r.Body.Close()
	if e != nil {
		httpErr(r.URL.Path, nil, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
		return
	}
	if txjson, e = compact(txjson); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
		return
	}

	_, _, cp, e := readFromJSON(txjson, t.LastInvokeID)
	if e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", e.Error(), w)
		return
	}

	if r.Method == http.MethodDelete {
		t.End(cp...)
		w.WriteHeader(http.StatusNoContent)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusNoContent, nil, nil)
		return
	}

	var id string
	if cp, e = t.Continue(cp...); e == nil {
		tid := t.GetIdentity()
		id = hex.EncodeToString([]byte{
			byte(tid >> 24), byte(tid >> 16), byte(tid >> 8), byte(tid)})
		if iv, ok := cp[0].(gsmap.Invoke); ok {
			t.LastInvokeID = iv.GetInvokeID()
		}
	} else if e != io.EOF {
		httpErr(r.URL.Path, txjson, http.StatusInternalServerError,
			"failed to continue TC", e.Error(), w)
		return
	}

	rxjson, e := writeToJSON(nil, &t.CdPA, cp)
	if e != nil {
		httpErr(r.URL.Path, txjson, http.StatusInternalServerError,
			"unable to marshal to JSON", e.Error(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if id != "" {
		w.Header().Set("Location", "/dialog/"+id)
		w.WriteHeader(http.StatusCreated)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusCreated, rxjson, nil)
	} else {
		w.WriteHeader(http.StatusOK)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusOK, rxjson, nil)
	}
	w.Write(rxjson)
}

func httpErr(
	path string, txj []byte, hcode int,
	title, detail string, w http.ResponseWriter) {

	data, _ := json.Marshal(struct {
		T string `json:"title"`
		D string `json:"detail"`
	}{T: title, D: detail})

	traceRxHttpRequest(path, txj, hcode, data, errors.New(title+": "+detail))

	w.Header().Add("Content-Type", "application/problem+json")
	w.WriteHeader(hcode)
	w.Write(data)
}
