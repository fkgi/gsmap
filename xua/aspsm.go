package xua

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"
)

/*
ASPSM: ASP State Maintenance Messages
Message class = 0x03
*/

/*
ASPUP is ASP Up message. (Message type = 0x01)

	 0                     1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|            Tag = 0x0011       |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                        ASP Identifier                         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|            Tag = 0x0004       |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPUP struct {
	// id *uint32
	// info   string
	retrans int
	result  chan error
}

func (m *ASPUP) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		switch r := c.state.(type) {
		case *down:
			c.state = m
			if e = c.send(m, 0); e != nil {
				c.state = r
				m.result <- e
			} else {
				time.AfterFunc(TAck, func() {
					c.msgQ <- &timeout{srcMessage: m}
				})
			}
		case *inactive, *active, *closing, *ASPUP, *ASPAC, *ASPIA, *ASPDN:
			e = errors.New("unexpected request")
			m.result <- e
		default:
			e = errors.New("unknown state")
			m.result <- e
		}
	} else {
		// handle Rx
		switch r := c.state.(type) {
		case *down, *ASPUP:
			c.state = &inactive{}
			if e = c.send(&ASPUPAck{}, 0); e != nil {
				c.state = r
			} else {
				go func() {
					r := make(chan error)
					c.msgQ <- &NTFY{status: statusInactive, result: r}
					<-r
				}()
			}
		case *inactive, *active, *closing, *ASPAC, *ASPIA, *ASPDN:
			e = c.send(&ASPUPAck{}, 0)
		default:
			e = errors.New("unknown state")
		}
	}
	return
}

func (m *ASPUP) marshal() (uint8, uint8, []byte) {
	// buf := new(bytes.Buffer)
	// if m.id != nil { // ASP Identifier (Optional)
	//	writeUint32(buf, 0x0011, *m.id)
	// }
	// if len(m.info) != 0 { // Info String (Optioal)
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x01, []byte{}
}

func (m *ASPUP) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	// switch t {
	// case 0x0011: // ASP Identifier (Optional)
	// 	var id uint32
	// 	id, e = readUint32(r, l)
	// 	m.id = &id
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	// default:
	_, e = r.Seek(int64(l), io.SeekCurrent)
	//}
	return
}

func (m *ASPUP) name() string {
	if m.result != nil {
		return "txASPUP"
	} else {
		return "rxASPUP"
	}
}

func (*ASPUP) state() string { return "wait_ASPUPAck" }

/*
ASPDN is ASP Down message. (Message type = 0x02)

	 0                     1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0004        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPDN struct {
	// info   string
	retrans int
	result  chan error
}

func (m *ASPDN) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		switch r := c.state.(type) {
		case *inactive, *active:
			c.state = m
			if e = c.send(m, 0); e != nil {
				c.state = r
				m.result <- e
			} else {
				time.AfterFunc(TAck, func() {
					c.msgQ <- &timeout{srcMessage: m}
				})
			}
		case *down, *closing, *ASPUP, *ASPAC, *ASPIA, *ASPDN:
			e = errors.New("unexpected request")
			m.result <- e
		default:
			e = errors.New("unknown state")
			m.result <- e
		}
	} else {
		// handle Rx
		switch r := c.state.(type) {
		case *down, *closing, *ASPUP:
			e = c.send(&ASPDNAck{}, 0)
		case *inactive, *active, *ASPAC, *ASPIA, *ASPDN:
			c.state = &closing{}
			if e = c.send(&ASPDNAck{}, 0); e != nil {
				c.state = r
			} else {
				time.AfterFunc(time.Millisecond*100, func() {
					if c.state.name() != "down" {
						c.msgQ <- &down{}
					}
				})
			}
		default:
			e = errors.New("unknown state")
		}
	}
	return
}

func (m *ASPDN) marshal() (uint8, uint8, []byte) {
	// buf := new(bytes.Buffer)
	// if len(m.info) != 0 { // Info String (Optioal)
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x02, []byte{}
}

func (m *ASPDN) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	// switch t {
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	// default:
	_, e = r.Seek(int64(l), io.SeekCurrent)
	// }
	return
}

func (m *ASPDN) name() string {
	if m.result != nil {
		return "txASPDN"
	} else {
		return "rxASPDN"
	}
}

func (*ASPDN) state() string { return "wait_ASPDNAck" }

/*
BEAT is Heartbeat message. (Message type = 0x03)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0009        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Heartbeat Data                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type BEAT struct {
	data    []byte
	retrans int
	result  chan error
}

func (m *BEAT) handle(c *ASP) (e error) {
	switch c.state.(type) {
	case *active, *inactive:
		if m.result != nil {
			// handle Tx
			if e = c.send(m, 0); e != nil {
				m.result <- e
			} else {
				time.AfterFunc(TAck, func() {
					c.msgQ <- &timeout{srcMessage: m}
				})
			}
		} else {
			// handle Rx
			c.send(&BEATAck{data: m.data}, 0)
		}
	default:
		e = errors.New("unexpected message")
		if m.result != nil {
			m.result <- e
		} else {
			c.send(&ERR{code: UnexpectedMessage}, 0)
		}
	}
	return
}

func (m *BEAT) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if len(m.data) != 0 { // Heartbeat Data (Optional)
		binary.Write(buf, binary.BigEndian, uint16(0x0009))
		binary.Write(buf, binary.BigEndian, uint16(4+len(m.data)))
		buf.Write(m.data)
		if len(m.data)%4 != 0 {
			buf.Write(make([]byte, 4-len(m.data)%4))
		}
	}
	return 0x03, 0x03, buf.Bytes()
}

func (m *BEAT) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0009: // Heartbeat Data (Optional)
		m.data = make([]byte, l)
		_, e = r.Read(m.data)
		if e == nil && l%4 != 0 {
			_, e = r.Seek(int64(4-l%4), io.SeekCurrent)
		}
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *BEAT) name() string {
	if m.result != nil {
		return "txABEAT"
	} else {
		return "rxBEAT"
	}
}

func (*BEAT) state() string { return "wait_BEATAck" }

/*
ASPUPAck is ASP Up Ack message. (Message type = 0x04)

	 0                     1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|            Tag = 0x0004       |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPUPAck struct {
	// info   string
}

func (m *ASPUPAck) handle(c *ASP) (e error) {
	switch r := c.state.(type) {
	case *ASPUP:
		c.state = &inactive{}
		c.msgQ <- &ASPAC{mode: Loadshare, ctx: c.ctx, result: r.result}
	case *down, *inactive, *active, *closing, *ASPAC, *ASPDN, *ASPIA:
		e = errors.New("unexpected message")
	default:
		e = errors.New("unknown state")
	}
	return
}

func (m *ASPUPAck) marshal() (uint8, uint8, []byte) {
	// buf := new(bytes.Buffer)
	// if len(m.info) != 0 { // Info String (Optioal)
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x04, []byte{}
}

func (m *ASPUPAck) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	// switch t {
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	// default:
	_, e = r.Seek(int64(l), io.SeekCurrent)
	// }
	return
}

func (m *ASPUPAck) name() string { return "rxASPUPAck" }
func (*ASPUPAck) state() string  { return "" }

/*
ASPDNAck is ASP Down Ack message. (Message type = 0x05)

	 0                     1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0004        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPDNAck struct {
	// info   string
}

func (m *ASPDNAck) handle(c *ASP) (e error) {
	switch r := c.state.(type) {
	case *ASPDN:
		c.state = &closing{}
		c.msgQ <- &down{}
		r.result <- nil
	case *down, *inactive, *active, *closing, *ASPUP, *ASPAC, *ASPIA:
		e = errors.New("unexpected message")
	default:
		e = errors.New("unknown state")
	}
	return
}

func (m *ASPDNAck) marshal() (uint8, uint8, []byte) {
	// buf := new(bytes.Buffer)
	// if len(m.info) != 0 { // Info String (Optioal)
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x05, []byte{}
}

func (m *ASPDNAck) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	// switch t {
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	// default:
	_, e = r.Seek(int64(l), io.SeekCurrent)
	// }
	return
}

func (m *ASPDNAck) name() string { return "rxASPDNAck" }
func (*ASPDNAck) state() string  { return "" }

/*
BEATAck is Heartbeat Ack message. (Message type = 0x06)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0009        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Heartbeat Data                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type BEATAck struct {
	data []byte
}

func (m *BEATAck) handle(c *ASP) (e error) {
	switch c.state.(type) {
	case *active, *inactive:
		// return r.handleResult(nil)
	default:
		e = errors.New("unexpected message")
	}
	return
}

func (m *BEATAck) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if len(m.data) != 0 { // Heartbeat Data (Optional)
		binary.Write(buf, binary.BigEndian, uint16(0x0009))
		binary.Write(buf, binary.BigEndian, uint16(4+len(m.data)))
		buf.Write(m.data)
		if len(m.data)%4 != 0 {
			buf.Write(make([]byte, 4-len(m.data)%4))
		}
	}
	return 0x03, 0x06, buf.Bytes()
}

func (m *BEATAck) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0009: // Heartbeat Data (Optional)
		m.data = make([]byte, l)
		_, e = r.Read(m.data)
		if e == nil && l%4 != 0 {
			_, e = r.Seek(int64(4-l%4), io.SeekCurrent)
		}
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *BEATAck) name() string { return "rxBEATAck" }
func (*BEATAck) state() string  { return "" }
