package xua

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	minWorkers = 128
	maxWorkers = 65535 - minWorkers
)

type userData struct {
	cgpa SCCPAddr
	cdpa SCCPAddr
	data []byte
}

var id = make(chan byte, 1)

func init() {
	id <- byte(time.Now().UnixMicro())
}
func nextID() byte {
	i := <-id + 1
	id <- i
	return i
}

type SignalingEndpoint struct {
	sock    int
	asps    chan map[int]*ASP
	block   chan any
	sharedQ chan userData

	PayloadHandler func(SCCPAddr, SCCPAddr, []byte)
	NetIndicator   uint8
	NetAppearance  *uint32
	Context        uint32
	PointCode      uint32
	SCCPAddr
	ReturnOnError bool
}

func NewSignalingEndpoint(a *SCTPAddr) (se *SignalingEndpoint, e error) {
	se = &SignalingEndpoint{}
	if a == nil || len(a.IP) == 0 {
		e = fmt.Errorf("nil address")
	} else if a.IP[0].To4() == nil {
		e = fmt.Errorf("invalid address")
	} else if se.sock, e = sockOpen(); e != nil {
	} else if e = sctpBindx(se.sock, a.rawBytes()); e != nil {
		sockClose(se.sock)
	} else {
		se.asps = make(chan map[int]*ASP, 1)
		se.asps <- map[int]*ASP{}
		se.block = make(chan any)
		se.sharedQ = make(chan userData, maxWorkers)
	}

	var activeWorkers = make(chan int, 1)
	activeWorkers <- 0

	worker := func() {
		for c := 0; c < 500; {
			if len(se.sharedQ) < minWorkers {
				time.Sleep(time.Millisecond * 10)
				c++
				continue
			}
			if req, ok := <-se.sharedQ; !ok {
				break
			} else if se.PayloadHandler == nil {
				se.selectASP().msgQ <- &TxDATA{
					ctx:      se.Context,
					cause:    SubsystemFailure,
					userData: userData{cgpa: se.SCCPAddr, cdpa: req.cgpa, data: req.data}}
			} else {
				se.PayloadHandler(req.cgpa, req.cdpa, req.data)
			}
			c = 0
		}
		activeWorkers <- (<-activeWorkers - 1)
	}

	for i := 0; i < minWorkers; i++ {
		go func() {
			for req, ok := <-se.sharedQ; ok; req, ok = <-se.sharedQ {
				a := <-activeWorkers
				activeWorkers <- a
				if len(se.sharedQ) > minWorkers && a < maxWorkers {
					activeWorkers <- (<-activeWorkers + 1)
					go worker()
				} else if se.PayloadHandler == nil {
					se.selectASP().msgQ <- &TxDATA{
						ctx:      se.Context,
						cause:    SubsystemFailure,
						userData: userData{cgpa: se.SCCPAddr, cdpa: req.cgpa, data: req.data}}
				} else {
					se.PayloadHandler(req.cgpa, req.cdpa, req.data)
				}
			}
		}()
	}
	return
}

/*
func (se *SignalingEndpoint) Listen() (e error) {
	if e = sockListen(se.sock); e != nil {
		sockClose(se.sock)
	}
	return
}
*/

func (se *SignalingEndpoint) ConnectTo(a *SCTPAddr) error {
	if a == nil || len(a.IP) == 0 {
		return fmt.Errorf("nil address")
	} else if a.IP[0].To4() == nil {
		return fmt.Errorf("invalid address")
	}

	go func() {
	svc:
		for {
			i := nextID()
			if SctpNotify != nil {
				SctpNotify(i, fmt.Sprintf("connecting to %s", a.String()))
			}
			if s, e := sctpConnectx(se.sock, a.rawBytes()); e == nil {
				asps := <-se.asps
				asps[s] = &ASP{id: i, sock: s, gt: se.SCCPAddr}
				se.asps <- asps

				asps[s].connectAndServe(se.Context, se.sharedQ)

				asps = <-se.asps
				delete(asps, s)
				se.asps <- asps
				sockClose(s)

				if SctpNotify != nil {
					SctpNotify(i, "closed")
				}
			} else if SctpNotify != nil {
				SctpNotify(i, "failed to connect: "+e.Error())
			}

			select {
			case <-se.block:
				break svc
			case <-time.After(time.Second * 30):
			}
		}
	}()
	return nil
}

func (se *SignalingEndpoint) Close() {
	close(se.block)
	asps := <-se.asps
	for _, v := range asps {
		if v.state == Active || v.state == Inactive {
			r := make(chan error, 1)
			v.msgQ <- &ASPDN{result: r}
			<-r
		}
		sockClose(v.sock)
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

func (se *SignalingEndpoint) selectASP() (c *ASP) {
	asps := <-se.asps
	list := make([]*ASP, 0, len(asps))
	for _, c = range asps {
		list = append(list, c)
	}
	se.asps <- asps
	for {
		c = list[rand.Intn(len(list))]
		if c.state == Active {
			break
		}
	}
	return
}

func (se *SignalingEndpoint) Write(dpc uint32, cdpa SCCPAddr, data []byte) {
	c := se.selectASP()
	seq := <-c.sequence
	c.sequence <- seq + 1

	c.msgQ <- &TxDATA{
		na:            se.NetAppearance,
		ctx:           se.Context,
		opc:           se.PointCode,
		dpc:           dpc,
		ni:            se.NetIndicator,
		sls:           uint8(seq & SLSMask),
		returnOnError: se.ReturnOnError,
		userData: userData{
			cgpa: se.SCCPAddr,
			cdpa: cdpa,
			data: data}}
}
