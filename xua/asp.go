package xua

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"
)

var (
	// tr    = time.Second * 2 // Pending Recovery timer
	TAck       = time.Second * 2 // Wait Response timer
	MaxRetrans = 5
	// tbeat = time.Second * 30 // Heartbeat interval

	SLSMask uint8 = 0x0f
)

type ASP struct {
	sock   int
	ctx    uint32
	msgQ   chan message
	eventQ chan message
	state  message
}

func (c *ASP) State() string {
	switch c.state.(type) {
	case nil:
		return "down"
	case *inactive:
		return "inactive"
	case *active:
		return "active"
	case *ASPUP:
		return "waitingASPUP_ack"
	case *ASPAC:
		return "waitingASPAC_ack"
	case *ASPIA:
		return "waitingASPIA_ack"
	case *ASPDN:
		return "waitingASPDN_ack"
	default:
		return "unknown"
	}
}

func (c *ASP) LocalAddr() net.Addr {
	if ptr, n, e := sctpGetladdrs(c.sock); e != nil {
		return nil
	} else {
		return resolveFromRawAddr(ptr, n)
	}
}

func (c *ASP) RemoteAddr() net.Addr {
	if ptr, n, e := sctpGetpaddrs(c.sock); e != nil {
		return nil
	} else {
		return resolveFromRawAddr(ptr, n)
	}
}

func (c *ASP) connectAndServe(se *SignalingPoint) {
	c.msgQ = make(chan message, 1024)
	c.state = &down{}
	go c.recieve(se)
	go func() {
		// ASP up
		r := make(chan error, 1)
		c.msgQ <- &ASPUP{result: r}
		if <-r != nil {
			c.msgQ <- &down{}
		}
	}()

	// event procedure
	for m, ok := <-c.msgQ; ok; m, ok = <-c.msgQ {
		old := c.state
		e := m.handle(c)
		if TraceEvent != nil {
			TraceEvent(old.state(), c.state.state(), m.name(), e)
		}
		if _, ok := m.(*down); ok {
			break
		}
	}
}

func (c *ASP) acceptAndServe(se *SignalingPoint) {
	c.msgQ = make(chan message, 1024)
	c.state = &down{}
	go c.recieve(se)

	// event procedure
	for m, ok := <-c.msgQ; ok; m, ok = <-c.msgQ {
		old := c.state
		e := m.handle(c)
		if TraceEvent != nil {
			TraceEvent(old.state(), c.state.state(), m.name(), e)
		}
		if _, ok := m.(*down); ok {
			break
		}
	}
}

func (c *ASP) recieve(se *SignalingPoint) {
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

		if msg, ok := m.(*DATA); ok && msg.data.Cause == Success {
			if PayloadHandler == nil {
				if msg.data.ReturnOnError {
					seq := <-se.sequence
					se.sequence <- seq + 1
					c.msgQ <- &DATA{
						na:  se.NetAppearance,
						ctx: c.ctx,
						opc: se.LocalPointCode,
						dpc: se.GwPointCode,
						ni:  se.NetIndicator,
						sls: seq & SLSMask,
						data: UnitData{
							Cause: SubsystemFailure,
							CgPA:  msg.data.CdPA, CdPA: msg.data.CgPA,
							Data: msg.data.Data},
						result: make(chan error, 1)}
				}
				continue
			} else if msg.data.ProtocolClass == 0 {
				sharedQ <- msg.data
				workerCheck()
				continue
			}
		}
		c.msgQ <- m
	}
	if _, ok := c.state.(*down); !ok {
		c.msgQ <- &down{}
	}
}

func (c *ASP) send(m message, stream uint16) error {
	mc, mt, md := m.marshal()
	buf := new(bytes.Buffer)

	// version
	buf.WriteByte(1)
	// reserved
	buf.WriteByte(0)
	// Message Class
	buf.WriteByte(mc)
	// Message Type
	buf.WriteByte(mt)
	// Message Length
	binary.Write(buf, binary.BigEndian, uint32(len(md)+8))
	// Message Data
	buf.Write(md)

	_, e := sctpSend(c.sock, buf.Bytes(), stream)
	return e
}
