package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/tcap"
)

var client http.Client

func handleIncomingDialog(t *tcap.Transaction, cp []gsmap.Component) (
	[]gsmap.Component, tcap.Dialogue, tcap.ComponentHandler) {

	n, v := getContextName(t.GetContext())
	traceRxDialog(n, v)
	path := "/mapmsg/v1/" + n + "/" + v

	var internalErr []gsmap.Component
	if len(cp) == 0 {
		internalErr = []gsmap.Component{}
	} else if iv, ok := cp[0].(gsmap.Invoke); ok {
		t.LastInvokeID = iv.GetInvokeID()
		internalErr = []gsmap.Component{&gsmap.SystemFailure{InvokeID: t.LastInvokeID}}
	}

	txjson, e := writeToJSON(nil, &t.CdPA, cp)
	if e != nil {
		e = errors.New("failed to marshal JSON: " + e.Error())
		traceTxHttpRequest(path, nil, 0, nil, e)
		return internalErr, nil, nil
	}

	r, e := client.Post(backend+path, "application/json", bytes.NewBuffer(txjson))
	if e != nil {
		e = errors.New("failed to access to backend: " + e.Error())
		traceTxHttpRequest(path, txjson, 0, nil, e)
		return internalErr, nil, nil
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case http.StatusOK, http.StatusCreated:
		// End / Continue
	case http.StatusMovedPermanently:
		// Fallback
		if loc := r.Header.Get("Location"); loc == "" {
			e = errors.New("missing Location header in backend response")
			traceTxHttpRequest(path, txjson, r.StatusCode, nil, e)
			return nil, &tcap.ABRT{Source: tcap.SvcUser}, nil
		} else if _, a, ok := strings.Cut(loc, "mapmsg/v1/"); !ok || len(a) == 0 {
			e = errors.New("invalid Location header in backend response: " + loc)
			traceTxHttpRequest(path, txjson, r.StatusCode, nil, e)
			return nil, &tcap.ABRT{Source: tcap.SvcUser}, nil
		} else if n, v, ok = strings.Cut(a, "/"); !ok || len(v) == 0 {
			e = errors.New("invalid Location header in backend response: " + loc)
			traceTxHttpRequest(path, txjson, r.StatusCode, nil, e)
			return nil, &tcap.ABRT{Source: tcap.SvcUser}, nil
		} else {
			v, _, _ = strings.Cut(v, "/")
		}
		if ctx := getContext(n, v); ctx == 0 {
			e = errors.New("unknown context in backend response: " + n + "/" + v)
			traceTxHttpRequest(path, txjson, r.StatusCode, nil, e)
			return nil, &tcap.ABRT{Source: tcap.SvcUser}, nil
		} else {
			traceTxHttpRequest(path, txjson, r.StatusCode, nil, nil)
			return nil, &tcap.AARE{
				Context:   ctx,
				Result:    tcap.RejectPermanent,
				ResultSrc: tcap.SrcUsrACNameNotSupported}, nil
		}
	case http.StatusServiceUnavailable:
		// Discard
		traceTxHttpRequest(path, txjson, r.StatusCode, nil, nil)
		return nil, nil, nil
	case http.StatusNotAcceptable:
		// Abort
		traceTxHttpRequest(path, txjson, r.StatusCode, nil, nil)
		return nil, &tcap.ABRT{Source: tcap.SvcUser}, nil
	case http.StatusNoContent:
		// EmptyResult(obsolete)
		traceTxHttpRequest(path, txjson, r.StatusCode, nil, nil)
		return []gsmap.Component{tcap.EmptyResult{InvokeID: cp[0].GetInvokeID()}}, nil, nil
	default:
		e = errors.New("error from backend: " + r.Status)
		traceTxHttpRequest(path, txjson, r.StatusCode, nil, e)
		return internalErr, nil, nil
	}

	rxjson, e := io.ReadAll(r.Body)
	if e != nil {
		e = errors.New("failed to get data from backend: " + e.Error())
		traceTxHttpRequest(path, txjson, r.StatusCode, nil, e)
		return internalErr, nil, nil
	}
	if len(rxjson) == 0 {
		cp = []gsmap.Component{}
	} else if rxjson, e = compact(rxjson); e != nil {
		traceTxHttpRequest(path, txjson, r.StatusCode, rxjson, e)
		return internalErr, nil, nil
	} else if _, _, cp, e = readFromJSON(rxjson, t.LastInvokeID); e != nil {
		e = errors.New("failed to unmarshal JSON: " + e.Error())
		traceTxHttpRequest(path, txjson, r.StatusCode, rxjson, e)
		return internalErr, nil, nil
	}

	traceTxHttpRequest(path, txjson, r.StatusCode, rxjson, nil)
	if r.StatusCode == http.StatusOK {
		return cp, nil, nil
	} else {
		l := r.Header.Get("Location")
		f := func(t *tcap.Transaction, cp []gsmap.Component, e error) {
			following(t, cp, e, l)
		}
		return cp, nil, f
	}
}

func following(t *tcap.Transaction, cp []gsmap.Component, e error, path string) {
	for {
		if e == io.EOF {
			// End
			p := path + "/end"
			var req *http.Request
			var jsondata []byte
			if jsondata, e = writeToJSON(nil, &t.CdPA, cp); e != nil {
				req, e = http.NewRequest(http.MethodPost, backend+p, nil)
			} else if req, e = http.NewRequest(
				http.MethodPost, backend+p, bytes.NewBuffer(jsondata)); e != nil {
				req, e = http.NewRequest(http.MethodPost, backend+p, nil)
			} else {
				req.Header.Set("Content-Type", "application/json")
			}

			if e != nil {
				r, _ := client.Do(req)
				r.Body.Close()
				traceTxHttpRequest(p, jsondata, r.StatusCode, nil, e)
			} else {
				traceTxHttpRequest(p, nil, 0, nil, e)
			}
			return
		} else if e != nil {
			// Abort / T.O.
			p := path + "/abort"
			if req, err := http.NewRequest(http.MethodPost, backend+p, nil); err != nil {
				r, _ := client.Do(req)
				r.Body.Close()
				traceTxHttpRequest(p, nil, r.StatusCode, nil, e)
			} else {
				traceTxHttpRequest(p, nil, 0, nil, e)
			}
			return
		}

		// Continue
		p := path + "/continue"
		if iv, ok := cp[0].(gsmap.Invoke); ok {
			t.LastInvokeID = iv.GetInvokeID()
		}

		var txjson, rxjson []byte
		if txjson, e = writeToJSON(nil, &t.CdPA, cp); e != nil {
			e = errors.New("failed to marshal JSON: " + e.Error())
			traceTxHttpRequest(p, nil, 0, nil, e)
			t.End(&gsmap.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}

		var r *http.Response
		if r, e = client.Post(backend+p, "application/json", bytes.NewBuffer(txjson)); e != nil {
			e = errors.New("failed to access to backend: " + e.Error())
			traceTxHttpRequest(p, txjson, 0, nil, e)
			t.End(&gsmap.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}
		defer r.Body.Close()

		switch r.StatusCode {
		case http.StatusOK, http.StatusCreated, http.StatusAccepted:
			// End / Continue
		case http.StatusServiceUnavailable:
			// Discard
			traceTxHttpRequest(p, txjson, r.StatusCode, nil, nil)
			t.Discard()
			return
		case http.StatusNotAcceptable:
			// Abort
			traceTxHttpRequest(p, txjson, r.StatusCode, nil, nil)
			t.Reject()
			return
		default:
			e = errors.New("error from backend: " + r.Status)
			traceTxHttpRequest(p, txjson, r.StatusCode, nil, e)
			t.End(&gsmap.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}

		if rxjson, e = io.ReadAll(r.Body); e != nil {
			e = errors.New("failed to get data from backend: " + e.Error())
			traceTxHttpRequest(p, txjson, r.StatusCode, nil, e)
			t.End(&gsmap.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}
		if len(rxjson) == 0 {
			cp = []gsmap.Component{}
		} else if rxjson, e = compact(rxjson); e != nil {
			traceTxHttpRequest(p, txjson, r.StatusCode, rxjson, e)
			t.End(&gsmap.SystemFailure{InvokeID: t.LastInvokeID})
			return
		} else if _, _, cp, e = readFromJSON(rxjson, t.LastInvokeID); e != nil {
			e = errors.New("failed to unmarshal JSON: " + e.Error())
			traceTxHttpRequest(p, txjson, r.StatusCode, rxjson, e)
			t.End(&gsmap.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}

		traceTxHttpRequest(p, txjson, r.StatusCode, rxjson, nil)
		if r.StatusCode == http.StatusOK {
			// End
			t.End(cp...)
			return
		}
		// Continue
		cp, e = t.Continue(cp...)
	}
}

func compact(j []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	if e := json.Compact(buf, j); e != nil {
		return j, errors.New("not JSON data: " + e.Error())
	} else {
		return buf.Bytes(), nil
	}
}
