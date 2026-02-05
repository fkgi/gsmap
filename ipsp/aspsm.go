package ipsp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	result chan error
}

func (m *ASPUP) handleMessage(c *ASP) {
	if m.result != nil {
		if e := c.handleCtrlReq(m); e != nil {
			m.result <- e
		}
	} else {
		c.sendAnswer(&ASPUPAck{})
	}
}

func (m *ASPUP) handleResult(msg message) {
	switch res := msg.(type) {
	case *ERR:
		m.result <- fmt.Errorf("error with code %s", res.code)
	case *ASPUPAck:
		m.result <- nil
	default:
		m.result <- fmt.Errorf("unexpected result")
	}
}

func (m *ASPUP) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// ASP Identifier (Optional)
	// if m.id != nil {
	//	writeUint32(buf, 0x0011, *m.id)
	// }

	// Info String (Optioal)
	// if len(m.info) != 0 {
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x01, buf.Bytes()
}

func (m *ASPUP) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	// case 0x0011: // ASP Identifier (Optional)
	// 	var id uint32
	// 	id, e = readUint32(r, l)
	// 	m.id = &id
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

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
	result chan error
}

func (m *ASPDN) handleMessage(c *ASP) {
	if m.result != nil {
		if e := c.handleCtrlReq(m); e != nil {
			m.result <- e
		}
	} else {
		c.sendAnswer(&ASPDNAck{})
	}
}

func (m *ASPDN) handleResult(msg message) {
	switch res := msg.(type) {
	case *ERR:
		m.result <- fmt.Errorf("error with code %s", res.code)
	case *ASPDNAck:
		m.result <- nil
	default:
		m.result <- fmt.Errorf("unexpected result")
	}
}

func (m *ASPDN) marshal() (uint8, uint8, []byte) {
	// Info String (Optioal)
	// if len(m.info) != 0 {
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x02, []byte{}
}

func (m *ASPDN) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

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
	data   []byte
	result chan error
}

func (m *BEAT) handleMessage(c *ASP) {
	if m.result != nil {
		if e := c.handleCtrlReq(m); e != nil {
			m.result <- e
		}
	} else {
		c.sendAnswer(&BEATAck{})
	}
}

func (m *BEAT) handleResult(msg message) {
	switch res := msg.(type) {
	case *ERR:
		m.result <- fmt.Errorf("error with code %s", res.code)
	case *BEATAck:
		m.result <- nil
	default:
		m.result <- fmt.Errorf("unexpected result")
	}
}

func (m *BEAT) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Heartbeat Data (Optional)
	if len(m.data) != 0 {
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

func (m *ASPUPAck) handleMessage(c *ASP) { c.handleCtrlAns(m) }
func (m *ASPUPAck) handleResult(message) {}

func (m *ASPUPAck) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	// Info String (Optioal)
	// if len(m.info) != 0 {
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x04, buf.Bytes()
}

func (m *ASPUPAck) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

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

func (m *ASPDNAck) handleMessage(c *ASP) { c.handleCtrlAns(m) }
func (m *ASPDNAck) handleResult(message) {}

func (m *ASPDNAck) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	// Info String (Optioal)
	// if len(m.info) != 0 {
	// 	writeInfo(buf, m.info)
	// }
	return 0x03, 0x05, buf.Bytes()
}

func (m *ASPDNAck) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	// case 0x0004:	// Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

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

func (m *BEATAck) handleMessage(c *ASP) { c.handleCtrlAns(m) }
func (m *BEATAck) handleResult(message) {}

func (m *BEATAck) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Heartbeat Data (Optional)
	if len(m.data) != 0 {
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
