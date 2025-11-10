package tcap

import (
	"errors"
	"io"
	"math/rand"
	"time"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/xua"
)

var (
	activeTC = map[uint32]*Transaction{}
	tcLock   = make(chan struct{}, 1)
	token    = struct{}{}
)

func init() {
	tcLock <- token
}

type Transaction struct {
	otid    uint32
	dtid    uint32
	rxStack chan (Message)
	ctx     gsmap.AppContext

	CdPA         xua.SCCPAddr
	LastInvokeID int8
}

type ComponentHandler func(*Transaction, []gsmap.Component, error)

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
	<-tcLock
	for ok := true; ok; _, ok = activeTC[t.otid] {
		t.otid = rand.Uint32()
	}
	activeTC[t.otid] = t
	tcLock <- token
}

func (t *Transaction) deregister() {
	<-tcLock
	delete(activeTC, t.otid)
	tcLock <- token
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
	<-tcLock
	t = activeTC[id]
	tcLock <- token
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
		if re, ok := dres.(*AARE); !ok || re.Result != Accept {
			send(cgpa, &TcEnd{
				dtid:     t.dtid,
				dialogue: dres})
			return
		} else {
			t.ctx = re.Context
		}
	}
	t.register()

	var reqCp []gsmap.Component
	if len(msg.component) == 0 {
		if send(cgpa, &TcContinue{
			otid:     t.otid,
			dtid:     t.dtid,
			dialogue: dres}) != nil {
			t.deregister()
			return
		}
		dres = nil

		timer := time.AfterFunc(Tw, func() {
			t.rxStack <- &TcAbort{dtid: t.otid, pCause: TcTimeout}
		})
		res := <-t.rxStack
		timer.Stop()

		switch res.(type) {
		case *TcContinue:
			reqCp = res.Components()
		case *TcEnd:
			// ToDo: End with first component
			NewInvoke(t, res.Components())
			t.deregister()
			return
		case *TcAbort:
			t.deregister()
			return
		default:
			panic("unexpected response")
		}
	} else {
		reqCp = msg.component
	}

	cres, following := NewInvoke(t, reqCp)
	if cres == nil {
		t.Discard()
		return
	} else if len(cres) == 0 {
		t.Reject()
		return
	}

	if following == nil {
		send(cgpa, &TcEnd{
			dtid:      t.dtid,
			dialogue:  dres,
			component: cres})
		t.deregister()
		return
	}

	if send(cgpa, &TcContinue{
		otid:      t.otid,
		dtid:      t.dtid,
		dialogue:  dres,
		component: cres}) != nil {
		t.deregister()
		return
	}

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
