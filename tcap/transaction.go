package tcap

import (
	"errors"
	"io"
	"math/rand"
	"time"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/xua"
)

var activeTC = make(chan map[uint32]*Transaction, 1)

func init() {
	activeTC <- map[uint32]*Transaction{}
}

func register(t *Transaction) {
	tcs := <-activeTC
	for ok := true; ok; _, ok = tcs[t.otid] {
		t.otid = rand.Uint32()
	}
	tcs[t.otid] = t
	activeTC <- tcs
}

func deregister(t *Transaction) {
	tcs := <-activeTC
	delete(tcs, t.otid)
	activeTC <- tcs
}

func GetTransaction(id uint32) (t *Transaction) {
	tcs := <-activeTC
	t = tcs[id]
	activeTC <- tcs
	return
}

type Transaction struct {
	otid    uint32
	dtid    uint32
	rxStack chan (Message)
	ctx     gsmap.AppContext

	CdPA         xua.SCCPAddr
	CgPA         xua.SCCPAddr
	LastInvokeID int8
}

func (t *Transaction) GetContext() gsmap.AppContext {
	return t.ctx
}

func (t *Transaction) send(m Message) Message {
	if send(t.CdPA, t.CgPA, m) != nil {
		return &TcAbort{dtid: t.otid, pCause: TcNoDestination}
	}

	timer := time.AfterFunc(Tw, func() {
		t.rxStack <- &TcAbort{dtid: t.otid, pCause: TcTimeout}
	})
	m = <-t.rxStack
	timer.Stop()
	return m
}

func (t *Transaction) GetIdentity() uint32 {
	return t.otid
}

func (t *Transaction) verifyDalogue(d Dialogue) (e error) {
	switch d := d.(type) {
	case *AARE:
		if d.Result == RejectPermanent && d.ResultSrc == SrcUsrACNameNotSupported {
			e = FallbackError{Context: d.Context}
		} else if d.Result != Accept {
			e = errors.New("dialogue rejected")
		} else if d.Context != t.ctx {
			e = errors.New("context missmatch")
		}
	case nil:
		if t.ctx&0x000000000000000f != 0x0000000000000001 {
			e = errors.New("unexpected dialogue")
		}
	default:
		e = errors.New("unexpected dialogue")
	}
	return
}

// End transaction with TC-End.
func (t *Transaction) End(c ...gsmap.Component) {
	if GetTransaction(t.otid) == nil {
		return
	}
	send(t.CdPA, t.CgPA, &TcEnd{dtid: t.dtid, component: c})
	deregister(t)
}

// Continue transaction with TC-Continue.
// Result error is io.EOF if response is TC-End.
func (t *Transaction) Continue(c ...gsmap.Component) ([]gsmap.Component, error) {
	if GetTransaction(t.otid) == nil {
		return nil, &TcAbort{dtid: t.otid, pCause: TcUnrecognizedMessageType}
	}

	msg := t.send(&TcContinue{otid: t.otid, dtid: t.dtid, component: c})
	switch m := msg.(type) {
	case *TcContinue:
		return m.component, nil
	case *TcEnd:
		return m.component, io.EOF
	case *TcAbort:
		return nil, m
	default:
		panic("unexpected response")
	}
}

// pReject transaction with TC-Abort with provider cause.
func (t *Transaction) pReject(c Cause) {
	send(t.CdPA, t.CgPA, &TcAbort{dtid: t.dtid, pCause: c})
	deregister(t)
}

// Reject transaction with TC-Abort with user cause.
func (t *Transaction) Reject() {
	send(t.CdPA, t.CgPA, &TcAbort{dtid: t.dtid, uCause: &ABRT{Source: SvcUser}})
	deregister(t)
}

// Discard transaction without any message.
func (t *Transaction) Discard() {
	if TraceTxMessage != nil {
		TraceTxMessage(&TcAbort{dtid: t.dtid, pCause: TcDiscard}, nil)
	}
	deregister(t)
}

// DialTC begin local initiated new transaction.
func DialTC(ctx gsmap.AppContext, cdpa xua.SCCPAddr, cgpa xua.SCCPAddr, handshake bool, invokes ...gsmap.Component) (t *Transaction, c []gsmap.Component, e error) {
	t = &Transaction{
		CdPA:    cdpa,
		CgPA:    cgpa,
		rxStack: make(chan Message, 1),
		ctx:     ctx}
	register(t)

	if handshake {
		switch m := t.send(&TcBegin{otid: t.otid}).(type) {
		case *TcContinue:
			t.dtid = m.otid
		default:
			e = errors.New("unexpected response")
			deregister(t)
			return
		}
	}

	var d Dialogue
	if ctx&0x000000000000000f != 0x0000000000000001 {
		d = &AARQ{Context: ctx}
	}

	var msg Message
	if handshake {
		msg = &TcContinue{otid: t.otid, dtid: t.dtid, dialogue: d, component: invokes}
	} else {
		msg = &TcBegin{otid: t.otid, dialogue: d, component: invokes}
	}
	switch m := t.send(msg).(type) {
	case *TcContinue:
		t.dtid = m.otid
		if e = t.verifyDalogue(m.dialogue); e == nil {
			c = m.component
		} else {
			t.pReject(TcIncorrectTransactionPortion)
		}
		return
	case *TcEnd:
		if e = t.verifyDalogue(m.dialogue); e == nil {
			c = m.component
			e = io.EOF
		}
	case *TcAbort:
		e = m
		if d, ok := m.uCause.(*AARE); !ok {
		} else if d.Result == RejectPermanent && d.ResultSrc == SrcUsrACNameNotSupported {
			e = FallbackError{Context: d.Context}
		}
	default:
		e = errors.New("unexpected response")
	}
	deregister(t)
	return
}

// acceptTC begin peer initiated new transaction.
func acceptTC(msg *TcBegin, cgpa xua.SCCPAddr) {
	t := &Transaction{
		dtid:    msg.otid,
		CdPA:    cgpa,
		CgPA:    LocalGT,
		rxStack: make(chan Message, 1)}
	register(t)

	// handshake
	if msg.dialogue == nil && len(msg.component) == 0 {
		if send(t.CdPA, LocalGT, &TcContinue{otid: t.otid, dtid: t.dtid}) != nil {
			deregister(t)
		} else {
			timer := time.AfterFunc(Tw, func() {
				t.rxStack <- &TcAbort{dtid: t.otid, pCause: TcTimeout}
			})
			res := <-t.rxStack
			timer.Stop()

			switch m := res.(type) {
			case *TcContinue:
				msg.dialogue = m.dialogue
				msg.component = m.component
			case *TcEnd, *TcAbort:
				deregister(t)
				return
			default:
				panic("unexpected response")
			}
		}
	}

	if msg.dialogue == nil && len(msg.component) == 0 {
		// repetition of empty message
		t.pReject(TcIncorrectTransactionPortion)
		return
	} else if msg.dialogue == nil && len(msg.component) != 0 {
		// v1 begin
		if c, ok := msg.component[0].(gsmap.Invoke); ok {
			t.ctx = c.DefaultContext()
		} else {
			t.pReject(TcIncorrectTransactionPortion)
			return
		}
	} else if dlg, ok := msg.dialogue.(*AARQ); ok {
		// v2/v3 begin
		t.ctx = dlg.Context
	} else {
		// invalid dialogue
		t.pReject(TcIncorrectTransactionPortion)
		return
	}

	cres, dres, following := NewInvoke(t, msg.component)
	switch d := dres.(type) {
	case nil:
		dres = &AARE{Context: t.ctx, Result: Accept, ResultSrc: SrcUsrNull}
	case *ABRT:
		send(t.CdPA, t.CgPA, &TcAbort{dtid: t.dtid, uCause: dres})
		deregister(t)
		return
	case *AARE:
		if d.Result != Accept {
			send(t.CdPA, t.CgPA, &TcEnd{dtid: t.dtid, dialogue: dres})
			deregister(t)
			return
		}
	default:
		t.pReject(TcUnrecognizedMessageType)
		return
	}

	if cres == nil {
		t.Discard()
	} else if following == nil {
		send(t.CdPA, t.CgPA, &TcEnd{
			dtid:     t.dtid,
			dialogue: dres, component: cres})
		deregister(t)
	} else if send(t.CdPA, t.CgPA,
		&TcContinue{
			otid: t.otid, dtid: t.dtid,
			dialogue: dres, component: cres}) != nil {
		deregister(t)
	} else {
		timer := time.AfterFunc(Tw, func() {
			t.rxStack <- &TcAbort{dtid: t.otid, pCause: TcTimeout}
		})
		res := <-t.rxStack
		timer.Stop()

		switch m := res.(type) {
		case *TcContinue:
			following(t, m.component, nil)
		case *TcEnd:
			following(t, m.component, io.EOF)
		case *TcAbort:
			following(t, nil, m)
		default:
			panic("unexpected response")
		}
	}
}
