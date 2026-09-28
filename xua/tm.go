package xua

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

/*
TM: RTransfer Messages
Message class = 0x01
*/

/*
DATA is Payload Data message. (Message type = 0x01)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|        Tag = 0x0200           |          Length = 8           |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                       Network Appearance                      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|        Tag = 0x0006           |          Length = 8           |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                        Routing Context                        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|        Tag = 0x0210           |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	\                                                               \
	/                      * Protocol Data                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|        Tag = 0x0013           |          Length = 8           |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                        Correlation Id                         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Protocol Data

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                     Originating Point Code                    |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                     Destination Point Code                    |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|       SI      |       NI      |      MP       |      SLS      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	\                                                               \
	/                     User Protocol Data                        /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type DATA struct {
	na  uint32
	ctx uint32

	opc uint32
	dpc uint32
	// si uint8 = 0x03
	ni uint8
	// mp uint8 = 0x00
	sls uint8

	data UnitData // SCCP data

	// correlation *uint32
	result chan error
}

func (m *DATA) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		e = c.send(m, uint16(m.sls)+1)
		m.result <- e
	} else {
		// handle Rx
		if m.data.Cause == Success {
			// handle SCCP request
			if PayloadHandler != nil {
				PayloadHandler(m.data)
			} else if m.data.ReturnOnError {
				c.msgQ <- &DATA{
					na:  m.na,
					ctx: m.ctx,
					opc: m.dpc,
					dpc: m.opc,
					ni:  m.ni,
					sls: m.sls,
					data: UnitData{
						Cause: SubsystemFailure,
						CgPA:  m.data.CdPA, CdPA: m.data.CgPA,
						Data: m.data.Data}}
			}
		} else {
			// handle SCCP answer
			if TxFailureNotify != nil {
				TxFailureNotify(
					fmt.Errorf("error response(cause=%x) from peer", m.data.Cause),
					m.data.Data)
			}
		}
	}
	return
}

func (m *DATA) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.na != 0 { // Network Appearance (Optional)
		writeUint32(buf, 0x0200, m.na)
	}
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Protocol Data
	d := m.data.marshal()
	l := len(d)
	binary.Write(buf, binary.BigEndian, uint16(0x0210))
	binary.Write(buf, binary.BigEndian, uint16(16+l))
	binary.Write(buf, binary.BigEndian, m.opc)
	binary.Write(buf, binary.BigEndian, m.dpc)
	buf.WriteByte(0x03)
	buf.WriteByte(m.ni)
	buf.WriteByte(0x00)
	buf.WriteByte(m.sls)
	buf.Write(d)
	if l%4 != 0 {
		buf.Write(make([]byte, 4-l%4))
	}
	// if m.correlation != nil { // Correlation ID (Optional)
	//	writeUint32(buf, 0x0013, *m.correlation)
	// }
	return 0x01, 0x01, buf.Bytes()
}

func (m *DATA) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0200: // Network Appearance (Optional)
		m.na, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	case 0x0210: // Protocol Data
		d := make([]byte, 12)
		if _, e = io.ReadFull(r, d); e != nil {
			return
		}
		m.opc = uint32(d[0])<<24 | uint32(d[1])<<16 | uint32(d[2])<<8 | uint32(d[3])
		m.dpc = uint32(d[4])<<24 | uint32(d[5])<<16 | uint32(d[6])<<8 | uint32(d[7])
		m.ni = d[9]
		m.sls = d[11]

		d = make([]byte, l-12)
		if _, e = io.ReadFull(r, d); e == nil {
			e = m.data.unmarshal(d)
		}
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *DATA) name() string {
	if m.result != nil {
		return "txDATA"
	} else {
		return "rxDATA"
	}
}

func (*DATA) state() string { return "" }
