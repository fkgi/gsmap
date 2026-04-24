package ipsp

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

type SignalingEndpoint struct {
	sock     int
	asps     chan map[int]*ASP
	block    chan any
	sharedQ  chan userData
	sequence chan uint8

	PayloadHandler func(SCCPAddr, SCCPAddr, []byte)

	NetIndicator  uint8
	NetAppearance *uint32
	PointCode     uint32
	SCCPAddr      SCCPAddr
	ReturnOnError bool
}

func NewSignalingEndpoint(a *SCTPAddr) (se *SignalingEndpoint, e error) {
	var sock int
	if a == nil || len(a.IP) == 0 {
		e = fmt.Errorf("nil address")
	} else if a.IP[0].To4() == nil {
		e = fmt.Errorf("invalid address")
	} else if sock, e = sockOpen(); e != nil {
	} else if e = sctpBindx(sock, a.rawBytes()); e != nil {
		sockClose(sock)
	} else if e = sockListen(sock); e != nil {
		sockClose(sock)
	}
	if e != nil {
		return
	}

	se = &SignalingEndpoint{
		sock:     sock,
		asps:     make(chan map[int]*ASP, 1),
		block:    make(chan any),
		sharedQ:  make(chan userData, maxWorkers),
		sequence: make(chan uint8, 1)}
	se.asps <- map[int]*ASP{}
	se.sequence <- 0

	activeWorkers := make(chan int, 1)
	activeWorkers <- 0

	worker := func() {
		for c := 0; c < 500; {
			if len(se.sharedQ) < minWorkers {
				time.Sleep(time.Millisecond * 10)
				c++
			} else if req, ok := <-se.sharedQ; !ok {
				break
			} else {
				se.PayloadHandler(req.cgpa, req.cdpa, req.data)
				c = 0
			}
		}
		activeWorkers <- (<-activeWorkers - 1)
	}

	for range minWorkers {
		go func() {
			for req, ok := <-se.sharedQ; ok; req, ok = <-se.sharedQ {
				if se.PayloadHandler == nil {
					c := se.selectASP()
					c.msgQ <- &DATA{
						ctx:      c.ctx,
						cause:    SubsystemFailure,
						userData: userData{cgpa: se.SCCPAddr, cdpa: req.cgpa, data: req.data},
						result:   make(chan error, 1)}
					continue
				}

				a := <-activeWorkers
				if len(se.sharedQ) > minWorkers && a < maxWorkers {
					a++
					go worker()
				}
				activeWorkers <- a
				se.PayloadHandler(req.cgpa, req.cdpa, req.data)
			}
		}()
	}

	go func() {
		for s, e := sctpAccept(sock); e == nil; s, e = sctpAccept(sock) {
			go func() {
				asps := <-se.asps
				asps[s] = &ASP{sock: s, handler: se.PayloadHandler}
				se.asps <- asps

				asps[s].listenAndServe(se.sharedQ)

				asps = <-se.asps
				delete(asps, s)
				se.asps <- asps
			}()
		}
	}()
	return
}

func (se *SignalingEndpoint) Close() {
	close(se.block)

	asps := <-se.asps
	for _, v := range asps {
		if v.stat == Active || v.stat == Inactive {
			v.msgQ <- &ASPDN{sock: v.sock}
		}
	}
	se.asps <- asps

	for {
		asps := <-se.asps
		l := len(asps)
		se.asps <- asps

		if l == 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	sockClose(se.sock)
	close(se.sharedQ)
}

func (se *SignalingEndpoint) selectASP() (c *ASP) {
	asps := <-se.asps
	list := make([]*ASP, 0, len(asps))
	for _, c = range asps {
		if c.stat == Active {
			list = append(list, c)
		}
	}
	se.asps <- asps
	c = list[rand.Intn(len(list))]
	return
}

func (se *SignalingEndpoint) Write(dpc uint32, cdpa SCCPAddr, data []byte) {
	c := se.selectASP()
	seq := <-se.sequence
	se.sequence <- seq + 1

	r := make(chan error)
	c.msgQ <- &DATA{
		na:            se.NetAppearance,
		ctx:           c.ctx,
		opc:           se.PointCode,
		dpc:           dpc,
		ni:            se.NetIndicator,
		sls:           seq & SLSMask,
		returnOnError: se.ReturnOnError,
		userData: userData{
			cgpa: se.SCCPAddr,
			cdpa: cdpa,
			data: data},
		result: r}
	if e := <-r; e != nil && TxFailureNotify != nil {
		TxFailureNotify(e, data)
	}
}
