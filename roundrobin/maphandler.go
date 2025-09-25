package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/common"
	"github.com/fkgi/gsmap/tcap"
)

var client http.Client

func handleIncomingDialog(t *tcap.Transaction, cp []gsmap.Component) ([]gsmap.Component, tcap.ComponentHandler) {
	n, v := getContextName(t.GetContext())
	if *verbose {
		log.Println("[INFO]", "Rx new dialog", n, v)
	}

	if iv, ok := cp[0].(gsmap.Invoke); ok {
		t.LastInvokeID = iv.GetInvokeID()
	}

	jsondata, e := writeToJSON(nil, &t.CdPA, cp)
	if e != nil {
		log.Println("[ERROR]", "failed to unmarshal JSON:", e)
		return []gsmap.Component{
				&common.SystemFailure{InvokeID: t.LastInvokeID}},
			nil
	}

	r, e := client.Post(backend+"/mapmsg/v1/"+n+"/"+v,
		"application/json", bytes.NewBuffer(jsondata))
	if e != nil {
		log.Println("[ERROR]", "failed to access to backend:", e)
		return []gsmap.Component{
				&common.SystemFailure{InvokeID: t.LastInvokeID}},
			nil
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case http.StatusOK, http.StatusCreated:
	case http.StatusServiceUnavailable:
		return nil, nil
	case http.StatusNotAcceptable:
		return []gsmap.Component{}, nil
	default:
		log.Println("[ERROR]", "error from backend:", r.StatusCode, r.Status)
		return []gsmap.Component{
				&common.SystemFailure{InvokeID: t.LastInvokeID}},
			nil
	}

	jsondata, e = io.ReadAll(r.Body)
	if e != nil {
		log.Println("[ERROR]", "failed to get data from backend:", e)
		return []gsmap.Component{
				&common.SystemFailure{InvokeID: t.LastInvokeID}},
			nil
	}
	_, _, cp, e = readFromJSON(jsondata, t.LastInvokeID)
	if e != nil {
		log.Println("[ERROR]", "failed to unmarshal JSON:", e)
		return []gsmap.Component{
				&common.SystemFailure{}},
			nil
	}

	if r.StatusCode == http.StatusOK {
		return cp, nil
	}

	path := r.Header.Get("Location")
	return cp, func(t *tcap.Transaction, cp []gsmap.Component, e error) {
		following(t, cp, e, path)
	}
}

func following(t *tcap.Transaction, cp []gsmap.Component, e error, path string) {
	for {
		if e == io.EOF {
			var req *http.Request
			var jsondata []byte
			if jsondata, e = writeToJSON(nil, &t.CdPA, cp); e != nil {
				req, e = http.NewRequest(http.MethodDelete, backend+path, nil)
			} else if req, e = http.NewRequest(
				http.MethodDelete, backend+path, bytes.NewBuffer(jsondata)); e != nil {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, e = http.NewRequest(http.MethodDelete, backend+path, nil)
			}

			if e != nil {
				r, _ := client.Do(req)
				r.Body.Close()
			}
			return
		} else if e != nil {
			if req, e := http.NewRequest(http.MethodDelete, backend+path, nil); e != nil {
				r, _ := client.Do(req)
				r.Body.Close()
			}
			return
		}

		if iv, ok := cp[0].(gsmap.Invoke); ok {
			t.LastInvokeID = iv.GetInvokeID()
		}

		var jsondata []byte
		jsondata, e = writeToJSON(nil, &t.CdPA, cp)
		if e != nil {
			t.End(&common.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}

		var r *http.Response
		r, e = client.Post(backend+path, "application/json", bytes.NewBuffer(jsondata))
		if e != nil {
			t.End(&common.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}
		defer r.Body.Close()

		switch r.StatusCode {
		case http.StatusOK, http.StatusCreated:
		case http.StatusServiceUnavailable:
			t.Discard()
			return
		case http.StatusNotAcceptable:
			t.Reject()
			return
		default:
			t.End(&common.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}

		jsondata, e = io.ReadAll(r.Body)
		if e != nil {
			t.End(&common.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}
		_, _, cp, e = readFromJSON(jsondata, t.LastInvokeID)
		if e != nil {
			t.End(&common.SystemFailure{InvokeID: t.LastInvokeID})
			return
		}

		if r.StatusCode == http.StatusOK {
			t.End(cp...)
			return
		}
		cp, e = t.Continue(cp...)
	}
}
