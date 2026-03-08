package ipsp

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
	tr   = time.Second * 30 // Pending Recovery timer
	tack = time.Second * 2  // Wait Response timer

	SLSMask uint8 = 0x0f
)

/*
message of xUA

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

	// handleResult handles result of this message
	handleResult(message)

	// marshal returns Message Class, Message Type and binary Message Data
	marshal() (uint8, uint8, []byte)

	// unmarshal decodes specified Tag/length TLV value from reader
	unmarshal(uint16, uint16, io.ReadSeeker) error
}

func getMessage(c, t byte) message {
	switch c {
	case 0x00:
		switch t {
		case 0x00:
			return new(ERR)
		case 0x01:
			return new(NTFY)
		}
	case 0x01:
		switch t {
		case 0x01:
			return new(DATA)
		}
	case 0x02:
		switch t {
		case 0x01:
			return new(DUNA)
		case 0x02:
			return new(DAVA)
		case 0x04:
			return new(SCON)
		case 0x05:
			return new(DUPU)
		case 0x06:
			return new(DRST)
		}
	case 0x03:
		switch t {
		case 0x01:
			return new(ASPUP)
		case 0x02:
			return new(ASPDN)
		case 0x03:
			return new(BEAT)
		case 0x04:
			return new(ASPUPAck)
		case 0x05:
			return new(ASPDNAck)
		case 0x06:
			return new(BEATAck)
		}
	case 0x04:
		switch t {
		case 0x01:
			return new(ASPAC)
		case 0x02:
			return new(ASPIA)
		case 0x03:
			return new(ASPACAck)
		case 0x04:
			return new(ASPIAAck)
		}
	}
	return nil
}

type ASP struct {
	sock int
	apc  uint32
	ctx  uint32

	msgQ    chan message
	ctrlMsg message

	handler func(SCCPAddr, SCCPAddr, []byte)

	stat Status

	TxTransfer uint64
	RxTransfer uint64
	TxResponse uint64
	RxResponse uint64
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

func (c *ASP) connectAndServe(sharedQ chan userData) {
	c.msgQ = make(chan message, 1024)
	c.ctrlMsg = nil
	c.stat = Down

	go func() {
		// ASP up
		r := make(chan error, 1)
		c.msgQ <- &ASPUP{result: r}
		if <-r != nil {
			sockClose(c.sock)
			return
		}
		c.stat = Inactive

		// ASP active
		r = make(chan error, 1)
		c.msgQ <- &ASPAC{mode: Loadshare, ctx: c.ctx, result: r}
		if <-r == nil {
			c.stat = Active
			return
		}

		// ASP down
		c.msgQ <- &ASPDN{sock: c.sock}
	}()

	go func() { // rx data procedure
		for {
			data, e := sctpRecvmsg(c.sock)
			if eno, ok := e.(*syscall.Errno); ok && eno.Temporary() {
				continue
			} else if e != nil {
				break
			} else if data[0] != 1 || len(data) < 8 {
				if RxFailureNotify != nil {
					RxFailureNotify(fmt.Errorf("invalid lengh of data"), data)
				}
				continue
			}

			m := getMessage(data[2], data[3])
			if m == nil {
				if RxFailureNotify != nil {
					RxFailureNotify(fmt.Errorf("unknown message: %x-%x", data[2], data[3]), data)
				}
				continue
			}

			for r := bytes.NewReader(data[8 : uint32(data[4])<<24|uint32(data[5])<<16|uint32(data[6])<<8|uint32(data[7])]); r.Len() > 4; {
				var t, l uint16
				binary.Read(r, binary.BigEndian, &t)
				binary.Read(r, binary.BigEndian, &l)
				l -= 4

				if e := m.unmarshal(t, l, r); e != nil {
					if RxFailureNotify != nil {
						RxFailureNotify(fmt.Errorf("invalid data for tag %x: %v", t, e), data)
					}
				}
				if l%4 != 0 {
					r.Seek(int64(4-l%4), io.SeekCurrent)
				}
			}

			if msg, ok := m.(*DATA); ok && msg.cause == 0 && msg.protocolClass == 0 {
				c.RxTransfer++
				sharedQ <- msg.userData
			} else {
				c.msgQ <- m
			}
		}

		c.ctrlMsg = nil
		sockClose(c.sock)
		close(c.msgQ)
	}()

	// event procedure
	for m, ok := <-c.msgQ; ok; m, ok = <-c.msgQ {
		m.handleMessage(c)
	}
	c.stat = Down
}

func (c *ASP) handleCtrlReq(m message) (e error) {
	if c.ctrlMsg != nil {
		e = errors.New("any other request is waiting answer")
	} else if e = c.send(m); e == nil {
		c.ctrlMsg = m
		time.AfterFunc(tack, func() {
			if c.ctrlMsg == m {
				c.msgQ <- &ERR{code: ProtocolError}
			}
		})
	}
	return
}

func (c *ASP) handleCtrlAns(m message) {
	if c.ctrlMsg != nil {
		c.ctrlMsg.handleResult(m)
		c.ctrlMsg = nil
	}
}

func (c *ASP) send(m message) error {
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

	_, e := sctpSend(c.sock, buf.Bytes(), 0)
	return e
}
