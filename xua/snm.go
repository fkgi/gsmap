package xua

import (
	"bytes"
	"encoding/binary"
	"io"
)

/*
SNM/SSNM: Signalling Network Management Messages
Message class = 0x02
*/

/*
DUNA is Destination Unavailable message. (Message type = 0x01)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0006          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0012          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Mask       |                 Affected PC 1                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Affected Point Code                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8003          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                 Reserved                      |   SSN value   |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0112          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                    Reserved                   |      SMI      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0004          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type DUNA struct {
	ctx []uint32
	apc []PointCode
	ssn uint8
	smi uint8
	// info    string
}

func (m *DUNA) handleMessage(*ASP) {
	if DunaNotify != nil {
		DunaNotify(m.apc)
	}
}

func (m *DUNA) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readRoutingContext(r, l)
	case 0x0012: // Affeccted Point Code
		m.apc, e = readAPC(r, l)
	case 0x8003: // SSN (Optional)
		m.ssn, e = readUint8(r, l)
	case 0x0112: // SMI (Optional)
		m.smi, e = readUint8(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

/*
DAVA is Destination Available message. (Message type = 0x02)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0006          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0012          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Mask       |                 Affected PC 1                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Affected Point Code                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8003          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                 Reserved                      |   SSN value   |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0112          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                    Reserved                   |      SMI      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0004          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type DAVA struct {
	ctx []uint32
	apc []PointCode
	ssn uint8
	smi uint8
	// info    string
}

func (m *DAVA) handleMessage(*ASP) {
	if DavaNotify != nil {
		DavaNotify(m.apc)
	}
}

func (m *DAVA) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readRoutingContext(r, l)
	case 0x0012: // Affected Point Code
		m.apc, e = readAPC(r, l)
	case 0x8003: // SSN (Optional)
		m.ssn, e = readUint8(r, l)
	case 0x0112: // SMI (Optional)
		m.smi, e = readUint8(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

/*
DAUD is Destination State Audit message. (Message type = 0x03)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0006          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0012          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Mask       |                 Affected PC 1                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Affected Point Code                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8003          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                 Reserved                      |   SSN value   |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x010c          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|             Cause             |            User               |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0004          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type DAUD struct {
	ctx   []uint32
	apc   []PointCode
	ssn   uint8
	cause uint16
	user  uint16
	// info    string
}

func (m *DAUD) handleMessage(*ASP) {
	if DaudNotify != nil {
		DaudNotify(m.apc)
	}
}
func (m *DAUD) handleResult(message) {}

func (m *DAUD) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Routing Context (Optional)
	if len(m.ctx) != 0 {
		writeRoutingContext(buf, m.ctx)
	}

	// Affected PC
	writeAPC(buf, m.apc)

	// SSN (Optional)
	if m.ssn != 0 {
		writeUint32(buf, 0x8003, uint32(m.ssn))
	}

	// User/Cause (Optional)
	if m.user == 3 {
		binary.Write(buf, binary.BigEndian, uint16(0x010C))
		binary.Write(buf, binary.BigEndian, uint16(8))
		binary.Write(buf, binary.BigEndian, m.cause)
		binary.Write(buf, binary.BigEndian, m.user)
	}

	// Info String (Optional)
	// if len(m.info) != 0 {
	// 	writeInfo(buf, m.info)
	// }
	return 0x02, 0x03, buf.Bytes()
}

/*
SCON is  Signalling Congestion message. (Message type = 0x04)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0006          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0012          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Mask       |                 Affected PC 1                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Affected Point Code                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8003          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                 Reserved                      |   SSN value   |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0118          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                     * Congestion Level                        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0112          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                    Reserved                   |      SMI      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0004          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type SCON struct {
	ctx        []uint32
	apc        []PointCode
	ssn        uint8
	congestion uint32
	smi        uint8
	// info       string
}

func (m *SCON) handleMessage(*ASP) {
	if SconNotify != nil {
		SconNotify(m.apc, m.congestion)
	}
}

func (m *SCON) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readRoutingContext(r, l)
	case 0x0012: // Affected Point Code
		m.apc, e = readAPC(r, l)
	case 0x8003: // SSN (Optional)
		m.ssn, e = readUint8(r, l)
	case 0x0118: // Congestion Level
		m.congestion, e = readUint32(r, l)
	case 0x0112: // SMI (Optional)
		m.smi, e = readUint8(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

/*
DUPU is Destination User Part Unavailable. (Message type = 0x05)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0006          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0012          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Mask       |                 Affected PC 1                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Affected Point Code                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x010c          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|           * Cause             |          * User               |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0004          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	\                                                               \
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type DUPU struct {
	ctx   []uint32
	apc   []PointCode
	cause uint16
	user  uint16
	// info    string
}

func (m *DUPU) handleMessage(*ASP) {
	if DupuNotify != nil {
		DupuNotify(m.apc, m.cause)
	}
}

func (m *DUPU) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readRoutingContext(r, l)
	case 0x0012: // Affected Point Code
		m.apc, e = readAPC(r, l)
	case 0x010C: // Cause/User
		if e = binary.Read(r, binary.BigEndian, &m.cause); e == nil {
			e = binary.Read(r, binary.BigEndian, &m.user)
		}
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

/*
DRST is Destination Restricted message. (Message type = 0x06)

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0006          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0012          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Mask       |                 Affected PC 1                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                      * Affected Point Code                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8003          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                 Reserved                      |   SSN value   |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0112          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                    Reserved                   |      SMI      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0004          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                          Info String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type DRST struct {
	ctx []uint32
	apc []PointCode
	ssn uint8
	smi uint8
	// info    string
}

func (m *DRST) handleMessage(*ASP) {
	if DrstNotify != nil {
		DrstNotify(m.apc)
	}
}

func (m *DRST) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readRoutingContext(r, l)
	case 0x0012: // Affected Point Code
		m.apc, e = readAPC(r, l)
	case 0x8003: // SSN (Optional)
		m.ssn, e = readUint8(r, l)
	case 0x0112: // SMI (Optional)
		m.smi, e = readUint8(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}
