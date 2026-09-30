package xua

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

const (
	minWorkers = 10
	maxWorkers = 20000 - minWorkers
	cooltime   = time.Second * 5
)

var (
	sharedQ       = make(chan UnitData, maxWorkers)
	activeWorkers = make(chan int, 1)

	PayloadHandler func(UnitData) = nil
)

func init() {
	activeWorkers <- 0
	for range minWorkers {
		go func() {
			for req, ok := <-sharedQ; ok; req, ok = <-sharedQ {
				PayloadHandler(req)
			}
		}()
	}
}

func workerCheck() {
	a := <-activeWorkers
	acl := len(sharedQ)
	if acl < (a+minWorkers)*4/5 {
		activeWorkers <- a
		return
	}

	if a+acl > maxWorkers {
		acl = maxWorkers - a
	}
	a += acl
	activeWorkers <- a

	for range acl {
		go func() {
			t := time.NewTimer(cooltime)
			for run := true; run; {
				select {
				case req, ok := <-sharedQ:
					if ok {
						PayloadHandler(req)
						t.Reset(cooltime)
					} else {
						run = false
					}
				case <-t.C:
					run = false
				}
			}
			t.Stop()
			activeWorkers <- (<-activeWorkers - 1)
		}()
	}
}

type SignalingPoint struct {
	sock     int
	asps     chan map[int]*ASP
	block    chan any
	sequence chan uint8
	eventQ   chan message
	state    Status

	NetIndicator   uint8
	NetAppearance  uint32
	Context        uint32
	LocalPointCode uint32
	GwPointCode    uint32
}

func NewSignalingEndPoint(a *SCTPAddr) (se *SignalingPoint, e error) {
	se = &SignalingPoint{
		asps:     make(chan map[int]*ASP, 1),
		block:    make(chan any),
		sequence: make(chan uint8, 1),
		eventQ:   make(chan message, 1024),
		state:    statusDwon}
	se.sequence <- 0

	if a == nil || len(a.IP) == 0 {
		e = fmt.Errorf("nil address")
	} else if a.IP[0].To4() == nil {
		e = fmt.Errorf("invalid address")
	} else if se.sock, e = sockSeqpacketOpen(); e != nil {
	} else if e = sctpBindx(se.sock, a.rawBytes()); e != nil {
		sockClose(se.sock)
	}
	if e != nil {
		return
	}

	se.asps <- map[int]*ASP{}
	go func() {
		for m, ok := <-se.eventQ; ok; m, ok = <-se.eventQ {
			se.handleEvent(m)
		}
	}()
	return
}

func NewSignalingTransferPoint(a *SCTPAddr) (se *SignalingPoint, e error) {
	se = &SignalingPoint{
		asps:     make(chan map[int]*ASP, 1),
		block:    make(chan any),
		sequence: make(chan uint8, 1),
		eventQ:   make(chan message, 1024),
		state:    statusDwon}
	se.sequence <- 0

	if a == nil || len(a.IP) == 0 {
		e = fmt.Errorf("nil address")
	} else if a.IP[0].To4() == nil {
		e = fmt.Errorf("invalid address")
	} else if se.sock, e = sockStreamOpen(); e != nil {
	} else if e = sctpBindx(se.sock, a.rawBytes()); e != nil {
		sockClose(se.sock)
	} else if e = sockListen(se.sock); e != nil {
		sockClose(se.sock)
	}
	if e != nil {
		return
	}

	se.asps <- map[int]*ASP{}
	go func() {
		for m, ok := <-se.eventQ; ok; m, ok = <-se.eventQ {
			se.handleEvent(m)
		}
	}()
	return
}

func (se *SignalingPoint) handleEvent(m message) {
	switch m := m.(type) {
	case *NTFY:
		se.state = m.status
		if AsStateNotify != nil {
			switch m.status {
			case statusInactive:
				AsStateNotify("inactive")
			case statusActive:
				AsStateNotify("active")
			case statusPending:
				AsStateNotify("pending")
			}
		}
	case *DUNA:
		if DunaNotify != nil {
			DunaNotify(m.apc)
		}
	case *DAVA:
		if DavaNotify != nil {
			DavaNotify(m.apc)
		}
	case *DAUD:
		if DaudNotify != nil {
			DaudNotify(m.apc)
		}
	case *SCON:
		if SconNotify != nil {
			SconNotify(m.apc, m.congestion)
		}
	case *DUPU:
		if DupuNotify != nil {
			DupuNotify(m.apc, m.cause)
		}
	case *DRST:
		if DrstNotify != nil {
			DrstNotify(m.apc)
		}
	}
}

func (se *SignalingPoint) ConnectTo(a *SCTPAddr) {
	for {
		if s, e := sctpConnectx(se.sock, a.rawBytes()); e == nil {
			if TraceEvent != nil {
				TraceEvent("init", "down", "dialSuccess", nil)
			}

			a := &ASP{
				sock:   s,
				ctx:    se.Context,
				eventQ: se.eventQ}
			asps := <-se.asps
			asps[s] = a
			se.asps <- asps

			a.connectAndServe(se)

			asps = <-se.asps
			delete(asps, s)
			se.asps <- asps
			sockClose(s)
		} else if TraceEvent != nil {
			TraceEvent("init", "init", "dialFailed", nil)
		}

		select {
		case <-se.block:
			return
		case <-time.After(time.Second * 30):
		}
	}
}

func (se *SignalingPoint) ListenAndServe() error {
	f := func(s int) {
		if TraceEvent != nil {
			TraceEvent("init", "down", "acceptNewCon", nil)
		}

		a := &ASP{
			sock:   s,
			eventQ: se.eventQ}
		asps := <-se.asps
		asps[s] = a
		se.asps <- asps

		a.acceptAndServe(se)

		asps = <-se.asps
		delete(asps, s)
		se.asps <- asps
		sockClose(s)
	}

	for {
		s, e := sctpAccept(se.sock)
		if e != nil {
			return e
		}
		f(s)
	}
}

func (se *SignalingPoint) ListASPs() []*ASP {
	res := []*ASP{}
	asps := <-se.asps
	se.asps <- asps
	for _, a := range asps {
		res = append(res, a)
	}
	return res
}

func (se *SignalingPoint) State() string {
	switch se.state {
	case statusInactive:
		return "inactive"
	case statusActive:
		return "active"
	case statusPending:
		return "pending"
	default:
		return "down"
	}
}

func (se *SignalingPoint) Close() {
	close(se.block)

	asps := <-se.asps
	for _, c := range asps {
		if c.state != nil {
			r := make(chan error)
			c.msgQ <- &ASPDN{result: r}
			<-r
		}
	}
	se.asps <- asps

	for {
		asps := <-se.asps
		l := len(asps)
		se.asps <- asps
		if l != 0 {
			time.Sleep(time.Millisecond * 100)
		} else {
			break
		}
	}
	sockClose(se.sock)
}

func (se *SignalingPoint) Write(ud UnitData) error {
	if se.state != statusActive {
		return errors.New("AS is not active")
	}

	asps := <-se.asps
	list := make([]*ASP, 0, len(asps))
	for _, c := range asps {
		if _, ok := c.state.(*active); ok {
			list = append(list, c)
		}
	}
	se.asps <- asps
	if len(list) == 0 {
		return errors.New("no available route")
	}

	c := list[rand.IntN(len(list))]
	res := make(chan error)
	seq := <-se.sequence
	se.sequence <- seq + 1
	c.msgQ <- &DATA{
		na:     se.NetAppearance,
		ctx:    c.ctx,
		opc:    se.LocalPointCode,
		dpc:    se.GwPointCode,
		ni:     se.NetIndicator,
		sls:    seq & SLSMask,
		data:   ud,
		result: res}
	return <-res
}
