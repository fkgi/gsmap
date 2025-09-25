package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/tcap"
	"github.com/fkgi/gsmap/xua"
)

func handleOutgoingDialog(w http.ResponseWriter, r *http.Request) {

	var ctx gsmap.AppContext
	if p := strings.Split(r.URL.Path, "/"); len(p) != 5 {
		log.Println("[INFO]", "invalid path:", r.URL.Path)
		httpErr("invalid path", r.URL.Path,
			http.StatusNotFound, w)
		return
	} else if ctx = getContext(p[3], p[4]); ctx == 0 {
		log.Println("[INFO]", "unsupported context:",
			fmt.Sprintf("context=%s, version=%s", p[3], p[4]))
		httpErr("unsupported context",
			fmt.Sprintf("context=%s, version=%s", p[3], p[4]),
			http.StatusNotFound, w)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	jsondata, e := io.ReadAll(r.Body)
	defer r.Body.Close()
	if e != nil {
		log.Println("[ERROR]", "failed to read request:", e)
		httpErr("unable to read request body", e.Error(),
			http.StatusBadRequest, w)
		return
	}
	cdpa, _, cp, e := readFromJSON(jsondata, 1)
	if e != nil {
		log.Println("[ERROR]", "failed to unmarshal request from JSON:", e)
		httpErr("unexpected JSON data", e.Error(),
			http.StatusBadRequest, w)
		return
	}

	n, v := getContextName(ctx)
	if *verbose {
		log.Println("[INFO]", "Tx new dialog", n, v)
	}

	var t *tcap.Transaction
	if t, cp, e = tcap.DialTC(ctx, cdpa, cp...); e != nil && e != io.EOF {
		log.Println("[ERROR]", "failed to dial TC:", e)
		httpErr("failed to dial TC", e.Error(),
			http.StatusInternalServerError, w)
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
	writeHttpResponse(&t.CdPA, cp, id, w)
}

func handleContinueDialog(w http.ResponseWriter, r *http.Request) {
	var t *tcap.Transaction
	if p := strings.Split(r.URL.Path, "/"); len(p) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if b, e := hex.DecodeString(p[2]); e != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(b) != 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if t = tcap.GetTransaction(
		(uint32(b[0]) << 24) | (uint32(b[1]) << 16) |
			(uint32(b[2]) << 8) | (uint32(b[3]))); t == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.Header().Add("Allow", "POST, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	jsondata, e := io.ReadAll(r.Body)
	defer r.Body.Close()
	if e != nil {
		log.Println("[ERROR]", "failed to read request:", e)
		httpErr("unable to read request body", e.Error(),
			http.StatusBadRequest, w)
		return
	}

	_, _, cp, e := readFromJSON(jsondata, t.LastInvokeID)
	if e != nil {
		log.Println("[ERROR]", "failed to unmarshal request from JSON:", e)
		httpErr("unexpected JSON data", e.Error(),
			http.StatusBadRequest, w)
		return
	}

	if r.Method == http.MethodDelete {
		t.End(cp...)
		w.WriteHeader(http.StatusNoContent)
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
		log.Println("[ERROR]", "failed to continue TC:", e)
		httpErr("failed to continue TC", e.Error(),
			http.StatusInternalServerError, w)
		return
	}

	writeHttpResponse(&t.CdPA, cp, id, w)
}

func writeHttpResponse(cgpa *xua.SCCPAddr, cp []gsmap.Component, id string, w http.ResponseWriter) {
	jsondata, e := writeToJSON(nil, cgpa, cp)
	if e != nil {
		log.Println("[ERROR]", "failed to marshal response to JSON:", e)
		httpErr("unable to marshal to JSON", e.Error(),
			http.StatusInternalServerError, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if id != "" {
		w.Header().Set("Location", "/dialog/"+id)
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write(jsondata)
}

func httpErr(title, detail string, code int, w http.ResponseWriter) {
	data, _ := json.Marshal(struct {
		T string `json:"title"`
		D string `json:"detail"`
	}{T: title, D: detail})

	w.Header().Add("Content-Type", "application/problem+json")
	w.WriteHeader(code)
	w.Write(data)
}
