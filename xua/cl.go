package xua

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

/*
CL: SCCP Connectionless (CL) Messages
Message class = 0x07
*/

/*
CLDT is Connectionless Data Transfer message. (Message type = 0x01)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0006         |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                     * Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0115          |             Length = 8        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|              Reserved                         | *Protocol Cl. |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0102          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Source Address                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0103          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                   * Destination Address                       /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0116          |             Length = 8        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                      * Sequence  Control                      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0101          |             Length = 8        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|              Reserved                         | SS7 Hop Count |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0113          |             Length = 8        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                Reserved                       |   Importance  |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0114          |             Length = 8        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|              Reserved                         |  Msg Priority |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0013          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                         Correlation ID                        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0117          |            Length = 32        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	| first/remain  |             Segmentation Reference            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x010B          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                           * Data                              /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type CLDT struct {
	ctx           []uint32
	returnOnError bool
	protocolClass uint8
	//cgpa          SCCPAddr
	//cdpa          SCCPAddr
	sequenceCtrl uint32

	// hopCount   uint8
	// importance *uint8
	// priority   *uint8
	// correlation *uint32

	// first      bool
	// remain     uint8
	// segmentRef *uint32

	//data []byte
	userData
}

func (m CLDT) CgPA() SCCPAddr {
	return m.cgpa
}
func (m CLDT) CdPA() SCCPAddr {
	return m.cdpa
}
func (m CLDT) Data() []byte {
	return m.data
}

type TxCLDT CLDT
type RxCLDT CLDT

func (m *TxCLDT) handleMessage(c *ASP) {
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

	i, e := sctpSend(c.sock, buf.Bytes(), uint16(m.sequenceCtrl&SLSMask)+1, false)
	if TxFailureNotify != nil {
		if e != nil {
			TxFailureNotify(e, buf.Bytes())
		} else if i != len(buf.Bytes()) {
			TxFailureNotify(fmt.Errorf("failed to send complete data"), buf.Bytes())
		}
	}
}

func (*TxCLDT) handleResult(message) {}

func (m *TxCLDT) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Routing Context
	writeRoutingContext(buf, m.ctx)

	// Protocol Class
	if m.returnOnError {
		writeUint8(buf, 0x0115, m.protocolClass|0x80)
	} else {
		writeUint8(buf, 0x0115, m.protocolClass)
	}

	// Source Address
	d := m.cgpa.marshalSUA()
	binary.Write(buf, binary.BigEndian, uint16(0x0102))
	binary.Write(buf, binary.BigEndian, uint16(4+len(d)))
	buf.Write(d)

	// Destination Address
	d = m.cdpa.marshalSUA()
	binary.Write(buf, binary.BigEndian, uint16(0x0103))
	binary.Write(buf, binary.BigEndian, uint16(4+len(d)))
	buf.Write(d)

	// Sequence Control
	writeUint32(buf, 0x0116, m.sequenceCtrl)

	// SS7 Hop Count (Optional)
	// if m.hopCount != 0 {
	// 	writeUint8(buf, 0x0101, m.hopCount)
	// }

	// Importance (Optional)
	// if m.importance != nil {
	// 	writeUint8(buf, 0x0113, *m.importance)
	// }

	// Message Priority (Optional)
	// if m.priority != nil {
	//	writeUint8(buf, 0x0114, *m.priority)
	// }

	// Correlation ID (Optional)
	// if m.correlation != nil {
	//	writeUint32(buf, 0x0013, *m.correlation)
	// }

	// Segmentation (Optional)
	// if m.segmentRef != nil {
	//	var v uint32
	//	if m.first {
	//		v |= 0x80000000
	//	}
	//	v |= uint32(m.remain) << 32
	//	v |= *m.segmentRef
	//	writeUint32(buf, 0x0117, v)
	// }

	// Data
	writeData(buf, m.data)

	return 0x07, 0x01, buf.Bytes()
}

func (m *RxCLDT) handleMessage(c *ASP) {
	c.RxTransfer++

	if PayloadHandler != nil {
		PayloadHandler(m.cgpa, m.cdpa, m.data)
	} else if m.returnOnError {
		c.msgQ <- &TxCLDR{
			ctx:   m.ctx,
			cause: SubsystemFailure,
			userData: userData{
				cgpa: LocalAddr,
				cdpa: m.cgpa,
				data: m.data}}
	}
}

func (m *RxCLDT) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context
		m.ctx, e = readRoutingContext(r, l)
	case 0x0115: // Protocol Class
		m.protocolClass, e = readUint8(r, l)
		m.returnOnError = m.protocolClass&0x80 == 0x80
		m.protocolClass = m.protocolClass & 0x7F
	case 0x0102: // Source Address
		m.cgpa, e = readSUAAddr(r, l)
	case 0x0103: // Destination Address
		m.cdpa, e = readSUAAddr(r, l)
	case 0x0116: // Sequence Control
		m.sequenceCtrl, e = readUint32(r, l)
	// case 0x0101: // SS7 Hop Count (Optional)
	//	m.hopCount, e = readUint8(r, l)
	// case 0x0113: // Importance (Optional)
	//	var tmp uint8
	//	if tmp, e = readUint8(r, l); e == nil {
	//		m.importance = &tmp
	//	}
	// case 0x0114: // Message Priority (Optional)
	//	var tmp uint8
	//	if tmp, e = readUint8(r, l); e == nil {
	//		m.priority = &tmp
	//	}
	case 0x010B: // Data
		m.data, e = readData(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

/*
CLDR is Connectionless Data Response message. (Message type = 0x02)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0006         |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                     * Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0106          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                         * SCCP Cause                          |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0102          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Source Address                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0103          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                   * Destination Address                       /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0101          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|              Reserved                         | SS7 Hop Count |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0113          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                Reserved                       |   Importance  |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0114          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|              Reserved                         |  Msg Priority |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0013          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                         Correlation ID                        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0117          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	| first/remain  |             Segmentation Reference            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x010b          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-
	/                             Data                              /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type CLDR struct {
	ctx   []uint32
	cause Cause
	//cgpa  SCCPAddr
	//cdpa  SCCPAddr

	// hopCount   uint8
	// importance *uint8
	// priority   *uint8
	// correlation *uint32

	// first      bool
	// remain     uint8
	// segmentRef *uint32

	//data []byte
	userData
}
type RxCLDR CLDR
type TxCLDR CLDR

type Cause uint32

const (
	Success                               Cause = 0x0000
	NoTranslationForAnAddressOfSuchNature Cause = 0x0100
	NoTranslationForThisSpecificAddress   Cause = 0x0101
	SubsystemCongestion                   Cause = 0x0102
	SubsystemFailure                      Cause = 0x0103
	UnequippedUser                        Cause = 0x0104
	MtpFailure                            Cause = 0x0105
	NetworkCongestion                     Cause = 0x0106
	Unqualified                           Cause = 0x0107
	ErrorInMessageTransport               Cause = 0x0108
	ErrorInLocalProcessing                Cause = 0x0109
	DestinationCannotPerformReassembly    Cause = 0x010a
	SccpFailure                           Cause = 0x010b
	HopCounterViolation                   Cause = 0x010c
	SegmentationNotSupported              Cause = 0x010d
	SegmentationFailure                   Cause = 0x010e
)

func (m *TxCLDR) handleMessage(c *ASP) {
	c.TxResponse++

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

	i, e := sctpSend(c.sock, buf.Bytes(), 0, false)
	if TxFailureNotify != nil {
		if e != nil {
			TxFailureNotify(e, buf.Bytes())
		} else if i != len(buf.Bytes()) {
			TxFailureNotify(fmt.Errorf("failed to send complete data"), buf.Bytes())
		}
	}
}

func (*TxCLDR) handleResult(message) {}

func (m *TxCLDR) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Routing Context
	writeRoutingContext(buf, m.ctx)

	// SCCP Cause
	writeUint32(buf, 0x0106, uint32(m.cause))

	// Source Address
	d := m.cgpa.marshalSUA()
	binary.Write(buf, binary.BigEndian, uint16(0x0102))
	binary.Write(buf, binary.BigEndian, uint16(4+len(d)))
	buf.Write(d)

	// Destination Address
	d = m.cdpa.marshalSUA()
	binary.Write(buf, binary.BigEndian, uint16(0x0103))
	binary.Write(buf, binary.BigEndian, uint16(4+len(d)))
	buf.Write(d)

	// SS7 Hop Count (Optional)
	// if m.hopCount != 0 {
	// 	writeUint8(buf, 0x0101, m.hopCount)
	// }

	// Importance (Optional)
	// if m.importance != nil {
	// 	writeUint8(buf, 0x0113, *m.importance)
	// }

	// Message Priority (Optional)
	// if m.priority != nil {
	//	writeUint8(buf, 0x0114, *m.priority)
	// }

	// Correlation ID (Optional)
	// if m.correlation != nil {
	//	writeUint32(buf, 0x0013, *m.correlation)
	// }

	// Segmentation (Optional)
	// if m.segmentRef != nil {
	//	var v uint32
	//	if m.first {
	//		v |= 0x80000000
	//	}
	//	v |= uint32(m.remain) << 32
	//	v |= *m.segmentRef
	//	writeUint32(buf, 0x0117, v)
	// }

	// Data
	if len(m.data) != 0 {
		writeData(buf, m.data)
	}
	return 0x07, 0x02, buf.Bytes()
}

func (m *RxCLDR) handleMessage(c *ASP) {
	if TxFailureNotify != nil {
		TxFailureNotify(
			fmt.Errorf("error response(cause=%x) from peer", m.cause), m.data)
	}
	c.RxResponse++
}

func (m *RxCLDR) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context
		m.ctx, e = readRoutingContext(r, l)
	case 0x0106: // Sequence Control
		var tmp uint32
		tmp, e = readUint32(r, l)
		m.cause = Cause(tmp)
	case 0x0102: // Source Address
		m.cgpa, e = readSUAAddr(r, l)
	case 0x0103: // Destination Address
		m.cdpa, e = readSUAAddr(r, l)
	// case 0x0101: // SS7 Hop Count (Optional)
	//	m.hopCount, e = readUint8(r, l)
	// case 0x0113: // Importance (Optional)
	//	var tmp uint8
	//	if tmp, e = readUint8(r, l); e == nil {
	//		m.importance = &tmp
	//	}
	// case 0x0114: // Message Priority (Optional)
	//	var tmp uint8
	//	if tmp, e = readUint8(r, l); e == nil {
	//		m.priority = &tmp
	//	}
	case 0x010B: // Data
		m.data, e = readData(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}
