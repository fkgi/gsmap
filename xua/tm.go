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
	|                      * Routing Context                        |
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

Unitdata (UDT)

	Message type code     F 1 octet
	Protocol class        F 1 octet
	Called party address  V 3- octets
	Calling party address V 3- octets
	Data                  V 2- octets

Unitdata Service (UDTS)

	Message type code     F 1 octet
	Return cause          F 1 octet
	Called party address  V 3- octets
	Calling party address V 3- octets
	Data                  V 2– octets
*/
type DATA struct {
	// na  *uint32
	ctx uint32

	opc uint32
	dpc uint32
	// si uint8 = 0x03
	ni uint8
	// mp uint8 = 0x00
	sls uint8

	// SCCP data
	returnOnError bool
	protocolClass uint8
	cause         Cause
	//cgpa          SCCPAddr
	//cdpa          SCCPAddr
	//data          []byte

	userData

	// correlation *uint32
}

type TxDATA DATA
type RxDATA DATA

func (m *TxDATA) handleMessage(c *ASP) {
	c.TxTransfer++

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

	i, e := sctpSend(c.sock, buf.Bytes(), uint16(m.sls)+1, true)
	if TxFailureNotify != nil {
		if e != nil {
			TxFailureNotify(e, buf.Bytes())
		} else if i != len(buf.Bytes()) {
			TxFailureNotify(fmt.Errorf("failed to send complete data"), buf.Bytes())
		}
	}
}

func (*TxDATA) handleResult(message) {}

func (m *TxDATA) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Network Appearance (Optional)
	//if m.na != nil {
	//	writeUint32(buf, 0x0200, *m.na)
	//}

	// Routing Context
	writeUint32(buf, 0x0006, m.ctx)

	// Protocol Data
	ud := new(bytes.Buffer)
	if m.cause == Success {
		ud.WriteByte(0x09)
		if m.returnOnError {
			ud.WriteByte((m.protocolClass & 0x0f) | 0x80)
		} else {
			ud.WriteByte(m.protocolClass & 0x0f)
		}
	} else {
		ud.WriteByte(0x0a)
		ud.WriteByte(byte(m.cause & 0x00ff))
	}
	cdpa := m.cdpa.marshalSCCP()
	cgpa := m.cgpa.marshalSCCP()
	ud.WriteByte(3)
	ud.WriteByte(byte(3 + len(cdpa)))
	ud.WriteByte(byte(3 + len(cdpa) + len(cgpa)))
	ud.WriteByte(byte(len(cdpa)))
	ud.Write(cdpa)
	ud.WriteByte(byte(len(cgpa)))
	ud.Write(cgpa)
	ud.WriteByte(byte(len(m.data)))
	ud.Write(m.data)
	l := ud.Len()

	binary.Write(buf, binary.BigEndian, uint16(0x0210))
	binary.Write(buf, binary.BigEndian, uint16(16+l))
	binary.Write(buf, binary.BigEndian, m.opc)
	binary.Write(buf, binary.BigEndian, m.dpc)
	buf.WriteByte(0x03)
	buf.WriteByte(m.ni)
	buf.WriteByte(0x00)
	buf.WriteByte(m.sls)
	ud.WriteTo(buf)
	if l%4 != 0 {
		buf.Write(make([]byte, 4-l%4))
	}

	// Correlation ID (Optional)
	// if m.correlation != nil {
	//	writeUint32(buf, 0x0013, *m.correlation)
	// }

	return 0x01, 0x01, buf.Bytes()
}

func (m *RxDATA) handleMessage(c *ASP) {
	if m.cause == Success {
		c.RxTransfer++
		if PayloadHandler != nil {
			PayloadHandler(m.cgpa, m.cdpa, m.data)
		} else if m.returnOnError {
			c.msgQ <- &TxDATA{
				ctx:   m.ctx,
				cause: SubsystemFailure,
				userData: userData{
					cgpa: LocalAddr,
					cdpa: m.cgpa,
					data: m.data}}
		}
	} else {
		if TxFailureNotify != nil {
			TxFailureNotify(
				fmt.Errorf("error response(cause=%x) from peer", m.cause), m.data)
		}
		c.RxResponse++
	}
}

func (m *RxDATA) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	//case 0x0200: // Network Appearance (Optional)
	//	*m.na, e = readUint32(r, l)
	case 0x0006: // Routing Context
		m.ctx, e = readUint32(r, l)
	case 0x0210: // Protocol Data
		d := make([]byte, l)
		if _, e = r.Read(d); e != nil {
			return
		}

		buf := bytes.NewReader(d)
		binary.Read(buf, binary.BigEndian, &m.opc)
		binary.Read(buf, binary.BigEndian, &m.dpc)
		buf.ReadByte()
		if m.ni, e = buf.ReadByte(); e != nil {
			return
		}
		buf.ReadByte()
		if m.sls, e = buf.ReadByte(); e != nil {
			return
		}

		var t byte
		if t, e = buf.ReadByte(); e != nil {
			return
		}
		switch t {
		case 0x09:
			t, e = buf.ReadByte()
			m.returnOnError = t&0x80 == 0x80
			m.protocolClass = t & 0x0f
		case 0x0a:
			t, e = buf.ReadByte()
			m.cause = Cause(t) | 0x0100
		default:
			e = fmt.Errorf("unknown SCCP message type(%x)", t)
			return
		}
		if e != nil {
			return
		}
		buf.Seek(int64(3), io.SeekCurrent)
		if m.cdpa, e = readSCCPAddr(buf); e != nil {
			return
		}
		if m.cgpa, e = readSCCPAddr(buf); e != nil {
			return
		}
		if t, e = buf.ReadByte(); e != nil {
			return
		}
		m.data = make([]byte, t)
		if _, e = buf.Read(m.data); e != nil {
			return
		}
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}
