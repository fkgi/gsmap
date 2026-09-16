package xua

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

/*
ASPTM: ASP Traffic Maintenance Messages
Message class = 0x04
*/

const (
	Override  uint32 = 1
	Loadshare uint32 = 2
	Broadcast uint32 = 3
)

/*
type Label struct {
	start uint8
	end   uint8
	value uint16
}
*/

/*
ASPAC is ASP Active message. (Message type = 0x01)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x000B        |           Length = 8          |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                       Traffic Mode Type                       |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0006        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0110         |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|     start     |      end      |        TID label value        |
	+-------------------------------+-------------------------------+
	|          Tag = 0x010F         |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|     start     |      end      |        DRN label value        |
	+-------------------------------+-------------------------------+
	|           Tag = 0x0004        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPAC struct {
	mode uint32
	ctx  uint32
	// tid  *Label
	// drn  *Label
	// info string
	retrans int
	result  chan error
}

func (m *ASPAC) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		switch r := c.state.(type) {
		case *inactive:
			c.state = m
			m.ctx = c.ctx
			if e = c.send(m, 0); e != nil {
				c.state = r
				m.result <- e
			} else {
				time.AfterFunc(TAck, func() {
					c.msgQ <- &timeout{srcMessage: m}
				})
			}
		case *down, *active, *closing, *ASPUP, *ASPAC, *ASPIA, *ASPDN:
			e = errors.New("unexpected request")
			m.result <- e
		default:
			e = errors.New("unknown state")
			m.result <- e
		}
	} else {
		// handle Rx
		switch r := c.state.(type) {
		case *inactive, *ASPAC:
			if c.ctx != 0 && m.ctx != c.ctx {
				c.state = &inactive{}
				e = c.send(&ERR{
					code: InvalidRoutingContext, ctx: m.ctx}, 0)
			} else if m.mode != 0 && m.mode != Loadshare {
				c.state = &inactive{}
				e = c.send(&ERR{
					code: UnsupportedTrafficHandlingMode, ctx: m.ctx}, 0)
			} else {
				c.state = &active{}
				ack := &ASPACAck{ctx: c.ctx}
				if m.mode == 0 {
					ack.mode = Loadshare
				}
				if e = c.send(ack, 0); e != nil {
					c.state = r
				}
			}
		case *down, *ASPUP:
			e = errors.New("unexpected request")
			e = c.send(&ERR{code: UnexpectedMessage, ctx: m.ctx}, 0)
		case *active, *closing, *ASPIA, *ASPDN:
			if c.ctx != 0 && m.ctx != c.ctx {
				e = c.send(&ERR{
					code: InvalidRoutingContext, ctx: m.ctx}, 0)
			} else if m.mode != 0 && m.mode != Loadshare {
				e = errors.New("invalid traffic mode")
			} else {
				ack := &ASPACAck{ctx: c.ctx}
				if m.mode == 0 {
					ack.mode = Loadshare
				}
				e = c.send(ack, 0)
			}
		default:
			e = errors.New("unknown state")
		}
	}
	return
}

func (m *ASPAC) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.mode != 0 { // Traffic Mode Type (Optional)
		writeUint32(buf, 0x000b, m.mode)
	}
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// if m.tid != nil { // TID Label (Optional)
	//	binary.Write(buf, binary.BigEndian, uint16(0x0110))
	//	binary.Write(buf, binary.BigEndian, uint16(8))
	//	buf.WriteByte(m.tid.start)
	//	buf.WriteByte(m.tid.end)
	//	binary.Write(buf, binary.BigEndian, m.tid.value)
	// }
	// if m.drn != nil { // DRN Label (Optional)
	// 	binary.Write(buf, binary.BigEndian, uint16(0x010F))
	// 	binary.Write(buf, binary.BigEndian, uint16(8))
	// 	buf.WriteByte(m.drn.start)
	// 	buf.WriteByte(m.drn.end)
	// 	binary.Write(buf, binary.BigEndian, m.drn.value)
	// }
	// if len(m.info) != 0 { // Info String (Optional)
	// 	writeInfo(buf, m.info)
	// }
	return 0x04, 0x01, buf.Bytes()
}

func (m *ASPAC) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x000B: // Traffic Mode Type (Optional)
		m.mode, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	// case 0x0110: // TID Label (Optional)
	// case 0x010F: // DRN Label (Optional)
	// case 0x0004: // Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *ASPAC) name() string {
	if m.result != nil {
		return "txASPAC"
	} else {
		return "rxASPAC"
	}
}

func (*ASPAC) state() string { return "wait_ASPACAck" }

/*
ASPIA is ASP Inactive message. (Message type = 0x02)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0006        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0004        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          INFO String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPIA struct {
	ctx uint32
	// info    string
	retrans int
	result  chan error
}

func (m *ASPIA) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		switch r := c.state.(type) {
		case *active:
			c.state = m
			m.ctx = c.ctx
			if e = c.send(m, 0); e != nil {
				c.state = r
				m.result <- e
			} else {
				time.AfterFunc(TAck, func() {
					c.msgQ <- &timeout{srcMessage: m}
				})
			}
		case *down, *inactive, *closing, *ASPUP, *ASPAC, *ASPIA, *ASPDN:
			e = errors.New("unexpected request")
			m.result <- e
		default:
			e = errors.New("unknown state")
			m.result <- e
		}
	} else {
		// handle Rx
		switch r := c.state.(type) {
		case *active, *ASPIA:
			if c.ctx != 0 && m.ctx != c.ctx {
				e = c.send(&ERR{
					code: InvalidRoutingContext, ctx: m.ctx}, 0)
			} else {
				c.state = &inactive{}
				if e = c.send(&ASPIAAck{ctx: c.ctx}, 0); e != nil {
					c.state = r
				}
			}
		case nil, *ASPUP:
			e = errors.New("unexpected request")
			e = c.send(&ERR{code: UnexpectedMessage, ctx: m.ctx}, 0)
		case *inactive, *ASPAC, *ASPDN:
			if c.ctx != 0 && m.ctx != c.ctx {
				e = errors.New("invalid context")
			} else {
				e = c.send(&ASPIAAck{ctx: c.ctx}, 0)
			}
		default:
			e = errors.New("unknown state")
		}
	}
	return
}

func (m *ASPIA) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// if len(m.info) != 0 { // Info String (Optional)
	// 	writeInfo(buf, m.info)
	// }
	return 0x04, 0x02, buf.Bytes()
}

func (m *ASPIA) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	// case 0x0004: // Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *ASPIA) name() string {
	if m.result != nil {
		return "txASPIA"
	} else {
		return "rxASPIA"
	}
}

func (*ASPIA) state() string { return "wait_ASPIAAck" }

/*
ASPACAck is ASP Active Ack message. (Message type = 0x03)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x000B        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                       Traffic Mode Type                       |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           Tag = 0x0006        |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                     * Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0004         |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPACAck struct {
	mode uint32
	ctx  uint32
	// info    string
}

func (m *ASPACAck) handle(c *ASP) (e error) {
	switch r := c.state.(type) {
	case *ASPAC:
		if c.ctx != 0 && m.ctx != c.ctx {
			e = fmt.Errorf("routing context missmatch")
		} else if m.mode != 0 && m.mode != r.mode {
			c.state = &inactive{}
			e = fmt.Errorf("traffic mode missmatch")
			r.result <- e
		} else {
			c.ctx = m.ctx
			c.state = &active{}
			r.result <- nil
		}
	case *down, *inactive, *active, *closing, *ASPUP, *ASPDN, *ASPIA:
		e = errors.New("unexpected message")
	default:
		e = errors.New("unknown state")
	}
	return
}

func (m *ASPACAck) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.mode != 0 { // Traffic Mode Type (Optional)
		writeUint32(buf, 0x000B, m.mode)
	}
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Info String (Optional)
	return 0x04, 0x03, buf.Bytes()
}

func (m *ASPACAck) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x000B: // Traffic Mode Type (Optional)
		m.mode, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *ASPACAck) name() string { return "rxASPACAck" }
func (*ASPACAck) state() string  { return "" }

/*
ASPIAAck is ASP Inactive Ack message. (Message type = 0x04)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0006         |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0004         |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ASPIAAck struct {
	ctx uint32
	// info    string
}

func (m *ASPIAAck) handle(c *ASP) (e error) {
	switch r := c.state.(type) {
	case *ASPIA:
		if c.ctx != 0 && m.ctx != c.ctx {
			e = fmt.Errorf("routing context missmatch")
		} else {
			c.state = &inactive{}
			r.result <- nil
		}
	case *down, *inactive, *active, *closing, *ASPUP, *ASPDN, *ASPAC:
		e = errors.New("unexpected message")
	default:
		e = errors.New("unknown state")
	}
	return
}

func (m *ASPIAAck) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Info String (Optional)
	return 0x04, 0x04, buf.Bytes()
}

func (m *ASPIAAck) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	// case 0x0004:	// Info String
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *ASPIAAck) name() string { return "rxASPIAAck" }
func (*ASPIAAck) state() string  { return "" }
