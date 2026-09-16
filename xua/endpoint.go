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
)

type SignalingEndpoint struct {
	sock     int
	asps     chan map[int]*ASP
	block    chan any
	sharedQ  chan UnitData
	sequence chan uint8
	eventQ   chan message
	state    Status

	PayloadHandler func(UnitData)
	NetIndicator   uint8
	NetAppearance  uint32
	Context        uint32
	LocalPointCode uint32
	GwPointCode    uint32
}

func NewSignalingEndpoint(a *SCTPAddr) (se *SignalingEndpoint, e error) {
	se = &SignalingEndpoint{
		asps:     make(chan map[int]*ASP, 1),
		block:    make(chan any),
		sharedQ:  make(chan UnitData, maxWorkers),
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
	for range minWorkers {
		go func() {
			for req, ok := <-se.sharedQ; ok; req, ok = <-se.sharedQ {
				if se.PayloadHandler != nil {
					se.PayloadHandler(req)
				} else if req.ReturnOnError {
					se.Write(UnitData{
						Cause: SubsystemFailure,
						CgPA:  req.CdPA, CdPA: req.CgPA,
						Data: req.Data})
				}
			}
		}()
	}
	go func() {
		activeWorkers := make(chan int, 1)
		activeWorkers <- 0
		act := true
		for act {
			acl := len(se.sharedQ)
			if acl < minWorkers/2 {
				time.Sleep(time.Millisecond * 10)
				continue
			}
			a := <-activeWorkers
			if a+acl > maxWorkers {
				acl = maxWorkers - a
			}
			a += acl
			activeWorkers <- a

			for range acl {
				go func() {
					for c := 0; c < 500; c++ {
						if len(se.sharedQ) < minWorkers/2 {
							time.Sleep(time.Millisecond * time.Duration(8+rand.IntN(4)))
						} else if req, ok := <-se.sharedQ; !ok {
							act = false
							break
						} else if se.PayloadHandler != nil {
							se.PayloadHandler(req)
							c = 0
						} else if req.ReturnOnError {
							se.Write(UnitData{
								Cause: SubsystemFailure,
								CgPA:  req.CdPA, CdPA: req.CgPA,
								Data: req.Data})
							c = 0
						}
					}
					activeWorkers <- (<-activeWorkers - 1)
				}()
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	go func() {
		for m, ok := <-se.eventQ; ok; m, ok = <-se.eventQ {
			se.handleEvent(m)
		}
	}()
	return
}

func (se *SignalingEndpoint) handleEvent(m message) {
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

/*
func (se *SignalingEndpoint) Listen() (e error) {
	if e = sockListen(se.sock); e != nil {
		sockClose(se.sock)
	}
	return
}
*/

func (se *SignalingEndpoint) ConnectTo(a *SCTPAddr) {
	for {
		if s, e := sctpConnectx(se.sock, a.rawBytes()); e == nil {
			if TraceEvent != nil {
				TraceEvent("init", "down", "dialSuccess", nil)
			}

			asps := <-se.asps
			asps[s] = &ASP{
				sock:    s,
				ctx:     se.Context,
				handler: se.PayloadHandler,
				eventQ:  se.eventQ}
			se.asps <- asps

			asps[s].connectAndServe(se.sharedQ)

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

func (se *SignalingEndpoint) Close() {
	close(se.block)

	asps := <-se.asps
	for _, c := range asps {
		if c.state != nil {
			r := make(chan error)
			c.msgQ <- &ASPDN{result: r}
			<-r
		}
		// sockClose(v.sock)
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

	close(se.sharedQ)
	sockClose(se.sock)
}

func (se *SignalingEndpoint) Write(ud UnitData) error {
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
