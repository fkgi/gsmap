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

type Transaction struct {
	otid    uint32
	dtid    uint32
	rxStack chan (Message)
	ctx     gsmap.AppContext

	CdPA         xua.SCCPAddr
	LastInvokeID int8
}

func (t *Transaction) GetContext() gsmap.AppContext {
	return t.ctx
}

func (t *Transaction) send(m Message) Message {
	if send(t.CdPA, m) != nil {
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

func (t *Transaction) register() {
	tcs := <-activeTC
	for ok := true; ok; _, ok = tcs[t.otid] {
		t.otid = rand.Uint32()
	}
	tcs[t.otid] = t
	activeTC <- tcs
}

func (t *Transaction) deregister() {
	tcs := <-activeTC
	delete(tcs, t.otid)
	activeTC <- tcs
}

func (t *Transaction) verifyDalogue(d Dialogue) error {
	switch d := d.(type) {
	case *AARE:
		if d.Result != Accept {
			return errors.New("dialogue rejected")
		} else if d.Context != t.ctx {
			return errors.New("context missmatch")
		}
	case nil:
		if t.ctx&0x000000000000000f != 0x0000000000000001 {
			return errors.New("unexpected dialogue")
		}
	default:
		return errors.New("unexpected dialogue")
	}
	return nil
}

func GetTransaction(id uint32) (t *Transaction) {
	tcs := <-activeTC
	t = tcs[id]
	activeTC <- tcs
	return
}

func (t *Transaction) End(c ...gsmap.Component) {
	send(t.CdPA, &TcEnd{dtid: t.dtid, component: c})
	t.deregister()
}

// Continue transaction.
// Result error is io.EOF if TC-End.
func (t *Transaction) Continue(c ...gsmap.Component) ([]gsmap.Component, error) {
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

func (t *Transaction) Reject() {
	send(t.CdPA, &TcAbort{dtid: t.dtid, uCause: &ABRT{Source: SvcUser}})
	t.deregister()
}

func (t *Transaction) Discard() {
	if TraceMessage != nil {
		TraceMessage(&TcAbort{dtid: t.dtid, pCause: TcDiscard}, Tx, nil)
	}
	t.deregister()
}

func DialTC(ctx gsmap.AppContext, cdpa xua.SCCPAddr, i ...gsmap.Component) (t *Transaction, c []gsmap.Component, e error) {
	t = &Transaction{
		CdPA:    cdpa,
		rxStack: make(chan Message, 1),
		ctx:     ctx}
	t.register()

	var d Dialogue
	if ctx&0x000000000000000f != 0x0000000000000001 {
		d = &AARQ{Context: ctx}
	}
	msg := t.send(&TcBegin{otid: t.otid, dialogue: d, component: i})

	switch m := msg.(type) {
	case *TcContinue:
		t.dtid = m.otid
		if e = t.verifyDalogue(m.dialogue); e == nil {
			c = m.component
		} else {
			sendAbort(t.CdPA, t.dtid, TcIncorrectTransactionPortion)
			t.deregister()
		}
	case *TcEnd:
		if e = t.verifyDalogue(m.dialogue); e == nil {
			c = m.component
			e = io.EOF
		}
	case *TcAbort:
		e = m
	default:
		e = errors.New("unexpected response")
	}

	return
}

func acceptTC(msg *TcBegin, cgpa xua.SCCPAddr) {
	t := &Transaction{
		dtid:    msg.otid,
		CdPA:    cgpa,
		rxStack: make(chan Message, 1)}

	var dres Dialogue
	if msg.dialogue == nil && len(msg.component) != 0 {
		if inv, ok := msg.component[0].(gsmap.Invoke); ok {
			t.ctx = inv.DefaultContext()
		}
	} else if dlg, ok := msg.dialogue.(*AARQ); ok {
		dres = DialogueHandler(*dlg)
		if re, ok := dres.(*ABRT); ok {
			send(cgpa, &TcAbort{dtid: t.dtid, uCause: re})
			return
		} else if re, ok := dres.(*AARE); !ok {
			send(cgpa, &TcAbort{dtid: t.dtid, pCause: TcUnrecognizedMessageType})
			return
		} else if re.Result != Accept {
			send(cgpa, &TcEnd{dtid: t.dtid, dialogue: dres})
			return
		} else {
			t.ctx = re.Context
		}
	}
	t.register()

	if cres, newctx, following := NewInvoke(t, msg.component); newctx != 0 {
		/*if newctx&0x000000000000000f == 0x0000000000000001 {
			send(cgpa, &TcAbort{
				dtid:   t.dtid,
				uCause: &ABRT{Source: SvcUser}})
		} else {*/
		send(cgpa, &TcAbort{
			dtid: t.dtid,
			uCause: &AARE{
				Context:   newctx,
				Result:    RejectPermanent,
				ResultSrc: SrcUsrACNameNotSupported}})
		// }
		t.deregister()
	} else if cres == nil {
		t.Discard()
	} else if len(cres) == 0 && len(msg.component) != 0 {
		t.Reject()
	} else if len(cres) == 0 && len(msg.component) == 0 && following == nil {
		t.Reject()
	} else if following == nil {
		send(cgpa, &TcEnd{
			dtid:      t.dtid,
			dialogue:  dres,
			component: cres})
		t.deregister()
	} else if send(cgpa, &TcContinue{
		otid:      t.otid,
		dtid:      t.dtid,
		dialogue:  dres,
		component: cres}) != nil {
		t.deregister()
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
			if m.pCause == TcTimeout {
				following(t, nil, errors.New("timeout"))
			} else {
				following(t, nil, errors.New(m.String()))
			}
		default:
			panic("unexpected response")
		}
	}
}
