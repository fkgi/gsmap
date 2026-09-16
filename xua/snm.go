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
	|         Tag = 0x0200          |          Length = 8           |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                      Network Appearance                       |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|        Tag = 0x0006           |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	\                                                               \
	/                       Routing Context                         /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x0012          |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|     Mask      |               * Affected PC 1                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	\                                                               \
	/                              ...                              /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|     Mask      |               * Affected PC n                 |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0004         |             Length            |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	\                                                               \
	/                          INFO String                          /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type DUNA struct {
	na  uint32
	ctx uint32
	apc []PointCode
	// info    string
	result chan error
}

func (m *DUNA) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		m.ctx = c.ctx
		e = c.send(m, 0)
		m.result <- e
	} else {
		// handle Rx
		if c.ctx != 0 && m.ctx != c.ctx {
			e = c.send(&ERR{code: InvalidRoutingContext, ctx: m.ctx}, 0)
		} else {
			c.eventQ <- m
		}
	}
	return
}

func (m *DUNA) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.na != 0 { // Network Appearance (Optional)
		writeUint32(buf, 0x0200, m.na)
	}
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Affeccted Point Code
	writeAPC(buf, m.apc)
	// Info String (Optional)
	return 0x02, 0x01, buf.Bytes()
}

func (m *DUNA) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0200: // Network Appearance (Optional)
		m.na, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	case 0x0012: // Affeccted Point Code
		m.apc, e = readAPC(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *DUNA) name() string {
	if m.result != nil {
		return "txDUNA"
	} else {
		return "rxDUNA"
	}
}

func (*DUNA) state() string { return "" }

/*
DAVA is Destination Available message. (Message type = 0x02)
DAVA Message parameters are the same as for the DUNA message.
*/
type DAVA struct {
	na  uint32
	ctx uint32
	apc []PointCode
	// info    string
	result chan error
}

func (m *DAVA) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		m.ctx = c.ctx
		e = c.send(m, 0)
		m.result <- e
	} else {
		// handle Rx
		if c.ctx != 0 && m.ctx != c.ctx {
			e = c.send(&ERR{code: InvalidRoutingContext, ctx: m.ctx}, 0)
		} else {
			c.eventQ <- m
		}
	}
	return
}

func (m *DAVA) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.na != 0 { // Network Appearance (Optional)
		writeUint32(buf, 0x0200, m.na)
	}
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Affeccted Point Code
	writeAPC(buf, m.apc)
	// Info String (Optional)
	return 0x02, 0x02, buf.Bytes()
}

func (m *DAVA) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0200: // Network Appearance (Optional)
		m.na, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	case 0x0012: // Affeccted Point Code
		m.apc, e = readAPC(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *DAVA) name() string {
	if m.result != nil {
		return "txDAVA"
	} else {
		return "rxDAVA"
	}
}

func (*DAVA) state() string { return "" }

/*
DAUD is Destination State Audit message. (Message type = 0x03)
DAUD Message parameters are the same as for the DUNA message.
*/
type DAUD struct {
	na  uint32
	ctx uint32
	apc []PointCode
	// info    string
	result chan error
}

func (m *DAUD) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		m.ctx = c.ctx
		e = c.send(m, 0)
		m.result <- e
	} else {
		// handle Rx
		if c.ctx != 0 && m.ctx != c.ctx {
			e = c.send(&ERR{code: InvalidRoutingContext, ctx: m.ctx}, 0)
		} else {
			c.eventQ <- m
		}
	}
	return
}

func (m *DAUD) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.na != 0 { // Network Appearance (Optional)
		writeUint32(buf, 0x0200, m.na)
	}
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Affeccted Point Code
	writeAPC(buf, m.apc)
	// Info String (Optional)
	return 0x02, 0x01, buf.Bytes()
}

func (m *DAUD) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0200: // Network Appearance (Optional)
		m.na, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	case 0x0012: // Affeccted Point Code
		m.apc, e = readAPC(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *DAUD) name() string {
	if m.result != nil {
		return "txDAUD"
	} else {
		return "rxDAUD"
	}
}

func (*DAUD) state() string { return "" }

/*
SCON is  Signalling Congestion message. (Message type = 0x04)

	       0                   1                   2                   3
	       0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |         Tag = 0x0200          |           Length = 8          |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |                     * Network Appearance                      |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |        Tag = 0x0006           |             Length            |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      \                                                               \
	      /                     * Routing Context                         /
	      \                                                               \
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |         Tag = 0x0012          |             Length            |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |      Mask     |                 Affected PC 1                 |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      \                                                               \
	      /                              ...                              /
	      \                                                               \
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |      Mask     |                 Affected PC n                 |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |         Tag = 0x0206          |             Length = 8        |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |    reserved   |               * Concerned DPC                 |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |         Tag = 0x0205          |             Length = 8        |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |                   Reserved                    | *Cong.  Level  |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      |            Tag = 0x0004       |             Length            |
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	      \                                                               \
	      /                       * INFO String                           /
	      \                                                               \
	      +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

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
	na         uint32
	ctx        uint32
	apc        []PointCode
	concerned  PointCode
	congestion uint32
	// info       string
	result chan error
}

func (m *SCON) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		m.ctx = c.ctx
		e = c.send(m, 0)
		m.result <- e
	} else {
		// handle Rx
		if c.ctx != 0 && m.ctx != c.ctx {
			e = c.send(&ERR{code: InvalidRoutingContext, ctx: m.ctx}, 0)
		} else {
			c.eventQ <- m
		}
	}
	return
}

func (m *SCON) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.na != 0 { // Network Appearance (Optional)
		writeUint32(buf, 0x0200, m.na)
	}
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Affeccted Point Code
	writeAPC(buf, m.apc)
	// Concerned DPC (Option)
	// Congestion Level (Option)
	writeUint32(buf, 0x0118, m.congestion)
	// Info String (Optional)
	return 0x02, 0x04, buf.Bytes()
}

func (m *SCON) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0200: // Network Appearance (Optional)
		m.na, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	case 0x0012: // Affected Point Code
		m.apc, e = readAPC(r, l)
	case 0x0206: // oncerned DPC (Optional)
		//m.ssn, e = readUint8(r, l)
	case 0x0205: // Congestion Level
		m.congestion, e = readUint32(r, l)
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

func (m *SCON) name() string {
	if m.result != nil {
		return "txSCON"
	} else {
		return "rxSCON"
	}
}

func (*SCON) state() string { return "" }

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
	ctx   uint32
	apc   []PointCode
	cause uint16
	user  uint16
	// info    string
	result chan error
}

func (m *DUPU) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		m.ctx = c.ctx
		e = c.send(m, 0)
		m.result <- e
	} else {
		// handle Rx
		if c.ctx != 0 && m.ctx != c.ctx {
			e = c.send(&ERR{code: InvalidRoutingContext, ctx: m.ctx}, 0)
		} else {
			c.eventQ <- m
		}
	}
	return
}

func (m *DUPU) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Affeccted Point Code
	writeAPC(buf, m.apc)
	// Cause/User
	binary.Write(buf, binary.BigEndian, 0x0204)
	binary.Write(buf, binary.BigEndian, uint16(8))
	binary.Write(buf, binary.BigEndian, m.cause)
	binary.Write(buf, binary.BigEndian, m.user)
	// Info String (Optional)
	return 0x02, 0x05, buf.Bytes()
}

func (m *DUPU) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	case 0x0012: // Affected Point Code
		m.apc, e = readAPC(r, l)
	case 0x0204: // Cause/User
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

func (m *DUPU) name() string {
	if m.result != nil {
		return "txDUPU"
	} else {
		return "rxDUPU"
	}
}

func (*DUPU) state() string { return "" }

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
	ctx uint32
	apc []PointCode
	ssn uint8
	smi uint8
	// info    string
	result chan error
}

func (m *DRST) handle(c *ASP) (e error) {
	if m.result != nil {
		// handle Tx
		m.ctx = c.ctx
		e = c.send(m, 0)
		m.result <- e
	} else {
		// handle Rx
		if c.ctx != 0 && m.ctx != c.ctx {
			e = c.send(&ERR{code: InvalidRoutingContext, ctx: m.ctx}, 0)
		} else {
			c.eventQ <- m
		}
	}
	return
}

func (m *DRST) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	if m.ctx != 0 { // Routing Context (Optional)
		writeUint32(buf, 0x0006, m.ctx)
	}
	// Affeccted Point Code
	writeAPC(buf, m.apc)
	if m.ssn != 0 { // SSN (Optional)
		writeUint8(buf, 0x8003, m.ssn)
	}
	if m.smi != 0 { // SMI (Optional)
		writeUint8(buf, 0x0112, m.smi)
	}
	// Info String (Optional)
	return 0x02, 0x06, buf.Bytes()
}

func (m *DRST) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
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

func (m *DRST) name() string {
	if m.result != nil {
		return "txDRST"
	} else {
		return "rxDRST"
	}
}

func (*DRST) state() string { return "" }
