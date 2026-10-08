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

// ComponentHandler handle continue component of active transaction.
type ComponentHandler func(*Transaction, []gsmap.Component, error)

/*
If response Component is nil, the dialog will be discarded (no response).
If response Dialogue is nil, the dialog will success with AARE and continue if ComponentHandler is not nil or end if ComponentHandler is nil.
If response Dialogue is ABRT, the dialog will be abort.
*/
var NewInvoke = func(*Transaction, []gsmap.Component) ([]gsmap.Component, Dialogue, ComponentHandler) {
	return []gsmap.Component{}, &ABRT{Source: SvcProvider}, nil
}

var EndPoint *xua.SignalingPoint
var LocalGT xua.SCCPAddr

type FallbackError struct {
	Context gsmap.AppContext
}

func (e FallbackError) Error() string {
	return fmt.Sprintf("fallback to %016x is required", e.Context)
}

func HandlePayload(ud xua.UnitData) {
	t, v, e := gsmap.ReadTLV(bytes.NewBuffer(ud.Data), 0x00)
	if e != nil {
		if TraceRxMessage != nil {
			TraceRxMessage(nil, fmt.Errorf("invalid data: %v", e))
		}
		return
	}

	switch t {
	case 0x61: // Unidirectional
		msg, e := unmarshalUnidirectional(v)
		if e == nil {
			e = fmt.Errorf("unidirectional is not supported")
		}
		if TraceRxMessage != nil {
			if e != nil {
				e = fmt.Errorf("invalid Unidirectional: %v", e)
			}
			TraceRxMessage(msg, e)
		}

	case 0x62: // Begin
		msg, e := unmarshalTcBegin(v)
		if TraceRxMessage != nil {
			if e != nil {
				e = fmt.Errorf("invalid Begin: %v", e)
			}
			TraceRxMessage(msg, e)
		}
		if e != nil {
			sendAbort(ud.CgPA, msg.otid, TcBadlyFormattedTransactionPortion)
		} else {
			go acceptTC(msg, ud.CgPA)
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
		if TraceRxMessage != nil {
			if e != nil {
				e = fmt.Errorf("invalid End: %v", e)
			}
			TraceRxMessage(msg, e)
		}
		if e == nil {
			t.CdPA = ud.CgPA
			t.rxStack <- msg
			deregister(t)
		}

	case 0x65: // Continue
		if msg, e := unmarshalTcContinue(v); e != nil {
			if TraceRxMessage != nil {
				TraceRxMessage(msg, fmt.Errorf("invalid Continue: %v", e))
			}
			sendAbort(ud.CgPA, msg.otid, TcBadlyFormattedTransactionPortion)
		} else if t := GetTransaction(msg.dtid); t == nil {
			if TraceRxMessage != nil {
				TraceRxMessage(msg, fmt.Errorf("invalid Continue: no active TC"))
			}
			sendAbort(ud.CgPA, msg.otid, TcUnrecognizedTransactionID)
		} else if len(t.rxStack) == cap(t.rxStack) {
			if TraceRxMessage != nil {
				TraceRxMessage(msg, fmt.Errorf("invalid Continue: unexpected response"))
			}
			sendAbort(ud.CgPA, msg.otid, TcResourceLimitation)
			deregister(t)
		} else {
			if TraceRxMessage != nil {
				TraceRxMessage(msg, e)
			}
			t.CdPA = ud.CgPA
			t.rxStack <- msg
		}

	case 0x67: // Abort
		msg, e := unmarshalTcAbort(v)
		var t *Transaction
		if e != nil {
		} else if t = GetTransaction(msg.dtid); t == nil {
			e = fmt.Errorf("no active TC")
		}
		if TraceRxMessage != nil {
			if e != nil {
				e = fmt.Errorf("invalid Abort data: %v", e)
			}
			TraceRxMessage(msg, e)
		}
		if e == nil {
			t.CdPA = ud.CgPA
			t.rxStack <- msg
			deregister(t)
		}
	}
}

func sendAbort(cdpa xua.SCCPAddr, tid uint32, cause Cause) {
	msg := &TcAbort{dtid: tid, pCause: cause}
	if tid != 0 {
		send(cdpa, LocalGT, msg)
	} else if TraceTxMessage != nil {
		TraceTxMessage(msg, fmt.Errorf("tid not defined"))
	}
}

func send(cdpa, cgpa xua.SCCPAddr, msg Message) (e error) {
	if EndPoint == nil {
		e = fmt.Errorf("failed to select destination")
	} else {
		e = EndPoint.Write(xua.UnitData{
			ReturnOnError: false,
			CdPA:          cdpa,
			CgPA:          cgpa,
			Data:          msg.marshalTc()})
	}
	if TraceTxMessage != nil {
		TraceTxMessage(msg, e)
	}
	return
}
