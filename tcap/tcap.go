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

func HandlePayload(ud xua.UnitData) {
	t, v, e := gsmap.ReadTLV(bytes.NewBuffer(ud.Data), 0x00)
	if e != nil {
		if RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid data: %v", e), ud.Data)
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
			TraceRxMessage(msg, e)
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid Unidirectional data: %v", e), ud.Data)
		}

	case 0x62: // Begin
		msg, e := unmarshalTcBegin(v)
		if TraceRxMessage != nil {
			TraceRxMessage(msg, e)
		}
		if e != nil {
			sendAbort(ud.CgPA, msg.otid, TcBadlyFormattedTransactionPortion)
		} else {
			go acceptTC(msg, ud.CgPA)
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid Begin data: %v", e), ud.Data)
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
			TraceRxMessage(msg, e)
		}
		if e == nil {
			t.CdPA = ud.CgPA
			t.rxStack <- msg
			t.deregister()
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid End data: %v", e), ud.Data)
		}

	case 0x65: // Continue
		if msg, e := unmarshalTcContinue(v); e != nil {
			if TraceRxMessage != nil {
				TraceRxMessage(msg, e)
			}
			sendAbort(ud.CgPA, msg.otid, TcBadlyFormattedTransactionPortion)

			if RxFailureNotify != nil {
				RxFailureNotify(fmt.Errorf("invalid Continue data: %v", e), ud.Data)
			}
		} else if t := GetTransaction(msg.dtid); t == nil {
			if TraceRxMessage != nil {
				TraceRxMessage(msg, fmt.Errorf("no active TC"))
			}
			sendAbort(ud.CgPA, msg.otid, TcUnrecognizedTransactionID)
		} else if len(t.rxStack) == cap(t.rxStack) {
			if TraceRxMessage != nil {
				TraceRxMessage(msg, fmt.Errorf("unexpected response"))
			}
			sendAbort(ud.CgPA, msg.otid, TcResourceLimitation)
			t.deregister()
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
			TraceRxMessage(msg, e)
		}
		if e == nil {
			t.CdPA = ud.CgPA
			t.rxStack <- msg
			t.deregister()
		}
		if e != nil && RxFailureNotify != nil {
			RxFailureNotify(fmt.Errorf("invalid Abort data: %v", e), ud.Data)
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
	}
	if TraceTxMessage != nil {
		TraceTxMessage(msg, e)
	}
	if EndPoint != nil {
		ud := xua.UnitData{
			ReturnOnError: false,
			CdPA:          cdpa,
			CgPA:          cgpa,
			Data:          msg.marshalTc()}
		e = EndPoint.Write(ud)
	}
	return
}
