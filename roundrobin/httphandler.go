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
	"github.com/fkgi/gsmap/xua"
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
	cdpa, cgpa, cp, e := readFromJSON(txjson, 1)
	if e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", e.Error(), w)
		return
	}
	if cdpa == nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", "no CdPA", w)
		return
	}
	if cgpa == nil {
		cgpa = &tcap.LocalGT
	}

	n, v := getContextName(ctx)
	traceTxDialog(n, v)

	var t *tcap.Transaction
	if t, cp, e = tcap.DialTC(ctx, *cdpa, *cgpa, false, cp...); e == nil || e == io.EOF {
	} else if fb, ok := e.(tcap.FallbackError); ok {
		if n, v := getContextName(fb.Context); n == "" || v == "" {
			httpErr(r.URL.Path, txjson, http.StatusInternalServerError,
				"fallback to unknown AC is required", e.Error(), w)
		} else {
			w.Header().Set("Location", "/mapmsg/v1/"+n+"/"+v)
			httpErr(r.URL.Path, txjson, http.StatusMovedPermanently,
				"fallback is required", e.Error(), w)
		}
		return
	} else {
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
		// Continue
		w.Header().Set("Location", "/dialog/"+id)
		w.WriteHeader(http.StatusCreated)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusCreated, rxjson, nil)
	} else {
		// End
		w.WriteHeader(http.StatusOK)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusOK, rxjson, nil)
	}
	w.Write(rxjson)
}

func getTransaction(s string) *tcap.Transaction {
	if b, e := hex.DecodeString(s); e != nil {
		return nil
	} else if len(b) != 4 {
		return nil
	} else {
		return tcap.GetTransaction(
			(uint32(b[0]) << 24) | (uint32(b[1]) << 16) |
				(uint32(b[2]) << 8) | (uint32(b[3])))
	}
}

func handleContinueDialogDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if t := getTransaction(r.PathValue("id")); t == nil {
		w.WriteHeader(http.StatusNotFound)
		traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
	} else if txjson, e := io.ReadAll(r.Body); e != nil {
		httpErr(r.URL.Path, nil, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
	} else if len(txjson) == 0 {
		// Discard
		t.Discard()
		w.WriteHeader(http.StatusNoContent)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusNoContent, nil, nil)
	} else if txjson, e = compact(txjson); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
	} else if cdpa, cgpa, cp, e := readFromJSON(txjson, t.LastInvokeID); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", e.Error(), w)
	} else {
		if cdpa != nil {
			t.CdPA = *cdpa
		}
		if cgpa != nil {
			t.CgPA = *cgpa
		}
		// End (obsolate)
		t.End(cp...)
		w.WriteHeader(http.StatusNoContent)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusNoContent, nil, nil)
	}
}

func handleContinueDialogEnd(w http.ResponseWriter, r *http.Request) {
	t := getTransaction(r.PathValue("id"))
	if t == nil {
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

	var cp []gsmap.Component
	var cdpa, cgpa *xua.SCCPAddr
	if len(txjson) == 0 {
	} else if txjson, e = compact(txjson); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
		return
	} else if cdpa, cgpa, cp, e = readFromJSON(txjson, t.LastInvokeID); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", e.Error(), w)
		return
	}
	if cdpa != nil {
		t.CdPA = *cdpa
	}
	if cgpa != nil {
		t.CgPA = *cgpa
	}

	// End
	t.End(cp...)
	w.WriteHeader(http.StatusNoContent)
	traceRxHttpRequest(r.URL.Path, txjson, http.StatusNoContent, nil, nil)
}

func handleContinueDialogContinue(w http.ResponseWriter, r *http.Request) {
	t := getTransaction(r.PathValue("id"))
	if t == nil {
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

	var cp []gsmap.Component
	var cdpa, cgpa *xua.SCCPAddr
	if len(txjson) == 0 {
	} else if txjson, e = compact(txjson); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
		return
	} else if cdpa, cgpa, cp, e = readFromJSON(txjson, t.LastInvokeID); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", e.Error(), w)
		return
	}
	if cdpa != nil {
		t.CdPA = *cdpa
	}
	if cgpa != nil {
		t.CgPA = *cgpa
	}

	// Continue
	var ct bool
	if cp, e = t.Continue(cp...); e == nil {
		// Rx Continue
		ct = true
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
	if ct {
		// Continue
		w.WriteHeader(http.StatusAccepted)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusAccepted, rxjson, nil)
	} else {
		// End
		w.WriteHeader(http.StatusOK)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusOK, rxjson, nil)
	}
	w.Write(rxjson)
}

// obsolate
func handleContinueDialog(w http.ResponseWriter, r *http.Request) {
	t := getTransaction(r.PathValue("id"))
	if t == nil {
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

	var cp []gsmap.Component
	var cdpa, cgpa *xua.SCCPAddr
	if len(txjson) == 0 {
	} else if txjson, e = compact(txjson); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unable to read request body", e.Error(), w)
		return
	} else if cdpa, cgpa, cp, e = readFromJSON(txjson, t.LastInvokeID); e != nil {
		httpErr(r.URL.Path, txjson, http.StatusBadRequest,
			"unexpected JSON data", e.Error(), w)
		return
	}
	if cdpa != nil {
		t.CdPA = *cdpa
	}
	if cgpa != nil {
		t.CgPA = *cgpa
	}

	// Continue
	var id string
	if cp, e = t.Continue(cp...); e == nil {
		// Rx Continue
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
		// Continue
		w.Header().Set("Location", "/dialog/"+id)
		w.WriteHeader(http.StatusCreated)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusCreated, rxjson, nil)
	} else {
		// End
		w.WriteHeader(http.StatusOK)
		traceRxHttpRequest(r.URL.Path, txjson, http.StatusOK, rxjson, nil)
	}
	w.Write(rxjson)
}

func handleContinueDialogAbort(w http.ResponseWriter, r *http.Request) {
	t := getTransaction(r.PathValue("id"))
	if t == nil {
		w.WriteHeader(http.StatusNotFound)
		traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
		return
	}
	// Abort
	t.Reject()
	w.WriteHeader(http.StatusNoContent)
	traceRxHttpRequest(r.URL.Path, nil, http.StatusNoContent, nil, nil)
}

/*
	func handleContinueDialogPost(w http.ResponseWriter, r *http.Request) {
		act := r.PathValue("action")
		switch act {
		case "", "continue", "end", "abort":
		default:
			w.WriteHeader(http.StatusNotFound)
			traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
			return
		}

		t := getTransaction(r.PathValue("id"))
		if t == nil {
			w.WriteHeader(http.StatusNotFound)
			traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
			return
		}

		if act == "abort" {
			// Abort
			t.Reject()
			w.WriteHeader(http.StatusNoContent)
			traceRxHttpRequest(r.URL.Path, nil, http.StatusNoContent, nil, nil)
			return
		}

		txjson, e := io.ReadAll(r.Body)
		defer r.Body.Close()
		if e != nil {
			httpErr(r.URL.Path, nil, http.StatusBadRequest,
				"unable to read request body", e.Error(), w)
			return
		}

		var cp []gsmap.Component
		if len(txjson) == 0 {
		} else if txjson, e = compact(txjson); e != nil {
			httpErr(r.URL.Path, txjson, http.StatusBadRequest,
				"unable to read request body", e.Error(), w)
			return
		} else if _, _, cp, e = readFromJSON(txjson, t.LastInvokeID); e != nil {
			httpErr(r.URL.Path, txjson, http.StatusBadRequest,
				"unexpected JSON data", e.Error(), w)
			return
		}
		if act == "end" {
			// End
			t.End(cp...)
			w.WriteHeader(http.StatusNoContent)
			traceRxHttpRequest(r.URL.Path, txjson, http.StatusNoContent, nil, nil)
			return
		}

		// Continue
		var id string
		if cp, e = t.Continue(cp...); e == nil {
			// Rx Continue
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
*/

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
