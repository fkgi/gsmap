package xua

import (
	"bytes"
	"fmt"
	"io"
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

	result chan error
}

func (m *ASPAC) handleMessage(c *ASP) {
	if e := handleCtrlReq(c, m); e != nil {
		m.result <- e
	}
}

func (m *ASPAC) handleResult(msg message) {
	switch res := msg.(type) {
	case *ERR:
		m.result <- fmt.Errorf("error with code %d", res.code)
	case *ASPACAck:
		if res.mode != m.mode {
			m.result <- fmt.Errorf("traffic mode missmatch")
		} else {
			m.result <- nil
		}
	default:
		m.result <- fmt.Errorf("unexpected result")
	}
}

func (m *ASPAC) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Traffic Mode Type (Optional)
	if m.mode != 0 {
		writeUint32(buf, 0x000b, m.mode)
	}

	// Routing Context (Optional)
	writeUint32(buf, 0x0006, m.ctx)

	// TID Label (Optional)
	// if m.tid != nil {
	//	binary.Write(buf, binary.BigEndian, uint16(0x0110))
	//	binary.Write(buf, binary.BigEndian, uint16(8))
	//	buf.WriteByte(m.tid.start)
	//	buf.WriteByte(m.tid.end)
	//	binary.Write(buf, binary.BigEndian, m.tid.value)
	// }

	// DRN Label (Optional)
	// if m.drn != nil {
	// 	binary.Write(buf, binary.BigEndian, uint16(0x010F))
	// 	binary.Write(buf, binary.BigEndian, uint16(8))
	// 	buf.WriteByte(m.drn.start)
	// 	buf.WriteByte(m.drn.end)
	// 	binary.Write(buf, binary.BigEndian, m.drn.value)
	// }

	// Info String (Optional)
	// if len(m.info) != 0 {
	// 	writeInfo(buf, m.info)
	// }
	return 0x04, 0x01, buf.Bytes()
}

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

	result chan error
}

func (m *ASPIA) handleMessage(c *ASP) {
	if e := handleCtrlReq(c, m); e != nil {
		m.result <- e
	}
}

func (m *ASPIA) handleResult(msg message) {
	switch res := msg.(type) {
	case *ERR:
		m.result <- fmt.Errorf("error with code %d", res.code)
	case *ASPIAAck:
		m.result <- nil
	default:
		m.result <- fmt.Errorf("unexpected result")
	}
}

func (m *ASPIA) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Routing Context (Optional)
	writeUint32(buf, 0x0006, m.ctx)

	// Info String (Optional)
	// if len(m.info) != 0 {
	// 	writeInfo(buf, m.info)
	// }
	return 0x04, 0x02, buf.Bytes()
}

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

func (m *ASPACAck) handleMessage(c *ASP) { handleCtrlAns(c, m) }

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

func (m *ASPIAAck) handleMessage(c *ASP) { handleCtrlAns(c, m) }

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
