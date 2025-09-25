package xua

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"
)

var (
	// tr    = time.Second * 2 // Pending Recovery timer
	tack = time.Second * 2 // Wait Response timer
	// tbeat = time.Second * 30 // Heartbeat interval

	ReturnOnError = false
	LocalAddr     SCCPAddr // local GT address
	LocalPC       uint32   // local Point Code
	SLSMask       uint32   = 0x0000000f
	Network       uint8    = 0
)

/*
Message of xUA

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Version    |   Reserved    | Message Class | Message Type  |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                        Message Length                         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                         Message Data                          |
*/
type message interface {
	// handleMsg handles this message
	handleMessage(*ASP)
}

type txMessage interface {
	message

	// handleResult handles result of this message
	handleResult(message)

	// marshal returns Message Class, Message Type and binary Message Data
	marshal() (uint8, uint8, []byte)
}

type rxMessage interface {
	message

	// unmarshal decodes specified Tag/length TLV value from reader
	unmarshal(uint16, uint16, io.ReadSeeker) error
}

type ASP struct {
	sock    int
	msgQ    chan message
	ctrlMsg txMessage
	ctxs    []uint32

	PointCode uint32

	state    Status
	sequence chan uint32

	TxTransfer uint64
	RxTransfer uint64
	TxResponse uint64
	RxResponse uint64
}

func (c ASP) State() Status {
	return c.state
}

func (c *ASP) LocalAddr() net.Addr {
	ptr, n, e := sctpGetladdrs(c.sock)
	if e != nil {
		return nil
	}
	return resolveFromRawAddr(ptr, n)
}

func (c *ASP) RemoteAddr() net.Addr {
	ptr, n, e := sctpGetpaddrs(c.sock)
	if e != nil {
		return nil
	}
	return resolveFromRawAddr(ptr, n)
}

func (c *ASP) DialAndServe(la, pa *SCTPAddr, ctx uint32) (e error) {
	c.msgQ = make(chan message, 1024)
	c.ctrlMsg = nil
	c.ctxs = []uint32{ctx}
	c.state = Down
	c.sequence = make(chan uint32, 1)
	c.sequence <- 0

	// dial SCTP
	if la == nil || pa == nil || len(la.IP) == 0 || len(pa.IP) == 0 {
		e = fmt.Errorf("nil address")
	} else if la.IP[0].To4() != nil && pa.IP[0].To4() != nil {
		c.sock, e = sockOpenV4()
	} else if la.IP[0].To16() != nil && pa.IP[0].To16() != nil {
		c.sock, e = sockOpenV6()
	} else {
		e = fmt.Errorf("invalid address")
	}
	if e != nil {
		e = &net.OpError{Op: "dial", Net: "sctp", Source: la, Addr: pa, Err: e}
		return
	}
	if e = sctpBindx(c.sock, la.rawBytes()); e != nil {
		_ = sockClose(c.sock)
		e = &net.OpError{Op: "dial", Net: "sctp", Source: la, Addr: pa, Err: e}
		return
	}
	if e = sctpConnectx(c.sock, pa.rawBytes()); e != nil {
		_ = sockClose(c.sock)
		e = &net.OpError{Op: "dial", Net: "sctp", Source: la, Addr: pa, Err: e}
		return
	}

	go func() { // event procedure
		for m, ok := <-c.msgQ; ok; m, ok = <-c.msgQ {
			m.handleMessage(c)
		}
	}()
	go func() {
		r := make(chan error, 1)
		c.msgQ <- &ASPUP{result: r}
		e = <-r
		if AspUpNotify != nil {
			AspUpNotify(la, pa, e)
		}
		if e != nil {
			_ = sockClose(c.sock)
			return
		}

		r = make(chan error, 1)
		c.msgQ <- &ASPAC{mode: Loadshare, ctx: c.ctxs, result: r}
		e = <-r
		if AsUpNotify != nil {
			AsUpNotify(c.ctxs, e)
		}
		if e != nil {
			_ = sockClose(c.sock)
		}
	}()

	for { // rx data procedure
		buf := make([]byte, 1500)

		n, e := sctpRecvmsg(c.sock, buf)
		if eno, ok := e.(*syscall.Errno); ok && eno.Temporary() {
			continue
		}
		if e != nil {
			break
		}

		m, e := readHandler(buf[:n])
		if e != nil {
			_ = sockClose(c.sock)
			break
		}

		if msg, ok := m.(*RxCLDT); ok && msg.protocolClass == 0 {
			c.RxTransfer++
			if PayloadHandler != nil {
				sharedQ <- msg.userData
			} else if msg.returnOnError {
				c.msgQ <- &TxCLDR{
					ctx:   msg.ctx,
					cause: SubsystemFailure,
					userData: userData{
						cgpa: LocalAddr,
						cdpa: msg.cgpa,
						data: msg.data}}
			}
		} else if msg, ok := m.(*RxDATA); ok && msg.cause == Success && msg.protocolClass == 0 {
			c.RxTransfer++
			if PayloadHandler != nil {
				sharedQ <- msg.userData
			} else if msg.returnOnError {
				c.msgQ <- &TxDATA{
					ctx:   msg.ctx,
					cause: SubsystemFailure,
					userData: userData{
						cgpa: LocalAddr,
						cdpa: msg.cgpa,
						data: msg.data}}
			}
		} else {
			c.msgQ <- m
		}
	}
	c.state = Down
	close(c.msgQ)

	return
}

func (c *ASP) Close() error {
	r := make(chan error, 1)
	c.msgQ <- &ASPDN{result: r}
	<-r

	return sockClose(c.sock)
}

func handleCtrlReq(c *ASP, m txMessage) (e error) {
	if c.ctrlMsg != nil {
		e = errors.New("any other request is waiting answer")
		return
	}

	cls, typ, b := m.marshal()
	buf := new(bytes.Buffer)

	// version
	buf.WriteByte(1)
	// reserved
	buf.WriteByte(0)
	// Message Class
	buf.WriteByte(cls)
	// Message Type
	buf.WriteByte(typ)
	// Message Length
	binary.Write(buf, binary.BigEndian, uint32(len(b)+8))
	// Message Data
	buf.Write(b)

	if _, e = sctpSend(c.sock, buf.Bytes(), 0, c.PointCode != 0); e != nil {
		return
	}

	c.ctrlMsg = m
	time.AfterFunc(tack, func() {
		if c.ctrlMsg == m {
			// Protocol Error
			c.msgQ <- &ERR{code: 0x07}
		}
	})
	return
}

func handleCtrlAns(c *ASP, m message) {
	if c.ctrlMsg != nil {
		c.ctrlMsg.handleResult(m)
		c.ctrlMsg = nil
	}
}

func readHandler(buf []byte) (m rxMessage, e error) {
	if buf[0] != 1 || len(buf) < 8 {
		e = errors.New("invalid version")
		return
	}

	r := bytes.NewReader(buf[4:])
	var l uint32
	if e = binary.Read(r, binary.BigEndian, &l); e != nil {
		return
	}

	switch buf[2] {
	case 0x00:
		switch buf[3] {
		case 0x00:
			m = new(ERR)
		case 0x01:
			m = new(NTFY)
		}
	case 0x01:
		switch buf[3] {
		case 0x01:
			m = new(RxDATA)
		}
	case 0x02:
		switch buf[3] {
		case 0x01:
			m = new(DUNA)
		case 0x02:
			m = new(DAVA)
		case 0x04:
			m = new(SCON)
		case 0x05:
			m = new(DUPU)
		case 0x06:
			m = new(DRST)
		}
	case 0x03:
		switch buf[3] {
		case 0x03:
			m = new(RxBEAT)
		case 0x04:
			m = new(ASPUPAck)
		case 0x05:
			m = new(ASPDNAck)
		case 0x06:
			m = new(RxBEATAck)
		}
	case 0x04:
		switch buf[3] {
		case 0x03:
			m = new(ASPACAck)
		case 0x04:
			m = new(ASPIAAck)
		}
	case 0x07:
		switch buf[3] {
		case 0x01:
			m = new(RxCLDT)
		case 0x02:
			m = new(RxCLDR)
		}
	}

	if m == nil {
		e = errors.New("invalid message")
		return
	}

	r = bytes.NewReader(buf[8:l])
	for r.Len() > 4 {
		var t, l uint16
		if e := binary.Read(r, binary.BigEndian, &t); e != nil {
			if RxFailureNotify != nil {
				RxFailureNotify(fmt.Errorf("invalid tag: %v", e), buf)
			}
			break
		}
		if e := binary.Read(r, binary.BigEndian, &l); e != nil {
			if RxFailureNotify != nil {
				RxFailureNotify(fmt.Errorf("invalid length: %v", e), buf)
			}
			break
		}
		l -= 4

		if e := m.unmarshal(t, l, r); e != nil {
			if RxFailureNotify != nil {
				RxFailureNotify(fmt.Errorf("invalid data: %v", e), buf)
			}
			break
		}

		if l%4 != 0 {
			r.Seek(int64(4-l%4), io.SeekCurrent)
		}
	}
	return
}

func (c *ASP) Write(cdpa SCCPAddr, data []byte) {
	seq := <-c.sequence
	c.sequence <- seq + 1

	if c.PointCode == 0 {
		c.msgQ <- &TxCLDT{
			ctx:           c.ctxs,
			returnOnError: ReturnOnError,
			sequenceCtrl:  seq,
			userData: userData{
				cgpa: LocalAddr,
				cdpa: cdpa,
				data: data}}
	} else {
		c.msgQ <- &TxDATA{
			ctx:           c.ctxs[0],
			opc:           LocalPC,
			dpc:           c.PointCode,
			ni:            Network,
			sls:           uint8(seq & SLSMask),
			returnOnError: ReturnOnError,
			userData: userData{
				cgpa: LocalAddr,
				cdpa: cdpa,
				data: data}}
	}
}
