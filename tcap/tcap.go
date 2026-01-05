package tcap

import (
	"bytes"
	"fmt"
	"time"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/xua"
)

var (
	Tw = time.Second * 30
)

type ComponentHandler func(*Transaction, []gsmap.Component, error)

/*
If response Component is nil, the dialog will be discarded (no response).
If response Component is empty slice, the dialog will be rejected (TC-Abort).
If response Component has elements, the dialog will be responded with the given components.
*/
var NewInvoke = func(*Transaction, []gsmap.Component) ([]gsmap.Component, gsmap.AppContext, ComponentHandler) {
	return []gsmap.Component{}, 0, nil
}

var EndPoint *xua.SignalingEndpoint
var PeerPointCode uint32

//var SelectASP = func() *xua.ASP {
//	return nil
//}

func HandlePayload(cgpa xua.SCCPAddr, cdpa xua.SCCPAddr, data []byte) {
	t, v, e := gsmap.ReadTLV(bytes.NewBuffer(data), 0x00)
	if e != nil {
		if RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid data: %v", e), data)
		}
		return
	}

	switch t {
	case 0x61: // Unidirectional
		msg, e := unmarshalUnidirectional(v)
		if e == nil {
			e = fmt.Errorf("unidirectional is not supported")
		}
		if TraceMessage != nil {
			TraceMessage(msg, Rx, e)
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid Unidirectional data: %v", e), data)
		}

	case 0x62: // Begin
		msg, e := unmarshalTcBegin(v)
		if TraceMessage != nil {
			TraceMessage(msg, Rx, e)
		}
		if e != nil {
			sendAbort(cgpa, msg.otid, TcBadlyFormattedTransactionPortion)
		} else {
			go acceptTC(msg, cgpa)
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid Begin data: %v", e), data)
		}

	case 0x64: // End
		msg, e := unmarshalTcEnd(v)
		var t *Transaction
		if e != nil {
		} else if t = GetTransaction(msg.dtid); t == nil {
			e = fmt.Errorf("no active TC")
		} else if len(t.rxStack) == cap(t.rxStack) {
			e = fmt.Errorf("unexpected response")
		}
		if TraceMessage != nil {
			TraceMessage(msg, Rx, e)
		}
		if e == nil {
			t.CdPA = cgpa
			t.rxStack <- msg
			t.deregister()
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid End data: %v", e), data)
		}

	case 0x65: // Continue
		if msg, e := unmarshalTcContinue(v); e != nil {
			if TraceMessage != nil {
				TraceMessage(msg, Rx, e)
			}
			sendAbort(cgpa, msg.otid, TcBadlyFormattedTransactionPortion)

			if RxFailureNotify != nil {
				RxFailureNotify(fmt.Errorf("invalid Continue data: %v", e), data)
			}
		} else if t := GetTransaction(msg.dtid); t == nil {
			if TraceMessage != nil {
				TraceMessage(msg, Rx, fmt.Errorf("no active TC"))
			}
			sendAbort(cgpa, msg.otid, TcUnrecognizedTransactionID)
		} else if len(t.rxStack) == cap(t.rxStack) {
			if TraceMessage != nil {
				TraceMessage(msg, Rx, fmt.Errorf("unexpected response"))
			}
			sendAbort(cgpa, msg.otid, TcResourceLimitation)
			t.deregister()
		} else {
			if TraceMessage != nil {
				TraceMessage(msg, Rx, e)
			}
			t.CdPA = cgpa
			t.rxStack <- msg
		}

	case 0x67: // Abort
		msg, e := unmarshalTcAbort(v)
		var t *Transaction
		if e != nil {
		} else if t = GetTransaction(msg.dtid); t == nil {
			e = fmt.Errorf("no active TC")
		}
		if TraceMessage != nil {
			TraceMessage(msg, Rx, e)
		}
		if e == nil {
			t.CdPA = cgpa
			t.rxStack <- msg
			t.deregister()
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid Abort data: %v", e), data)
		}
	}
}

func sendAbort(cdpa xua.SCCPAddr, tid uint32, cause Cause) {
	msg := &TcAbort{
		dtid:   tid,
		pCause: cause,
	}
	if tid == 0 {
		if TraceMessage != nil {
			TraceMessage(msg, Tx, fmt.Errorf("tid not defined"))
		}
		return
	}
	send(cdpa, msg)
}

func send(cdpa xua.SCCPAddr, msg Message) (e error) {
	if EndPoint == nil {
		e = fmt.Errorf("failed to select destination")
	}
	if TraceMessage != nil {
		TraceMessage(msg, Tx, e)
	}
	if EndPoint != nil {
		EndPoint.Write(PeerPointCode, cdpa, msg.marshalTc())
	}
	return
}
