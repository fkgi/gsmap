package xua

import (
	"io"
)

/*
MGMT: UA Management Messages
Message class = 0x00
*/

/*
ERR is Error message. (Message type = 0x00)
Direction is SGP -> ASP.

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x000C         |           Length = 8          |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                        * Error Code                           |
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
	/                        Affected Point Code                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x010D          |          Length = 8           |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                     Network Appearance                        |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x0007         |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                        Diagnostic Info                        /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
type ERR struct {
	code uint32
	ctx  uint32
	apc  []PointCode
	na   *uint32
	// info []byte
}

func (m *ERR) handleMessage(c *ASP) { handleCtrlAns(c, m) }

func (m *ERR) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x000C: // Error Code
		m.code, e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	case 0x0012: // Affected Point Code (Optional)
		m.apc, e = readAPC(r, l)
	case 0x010D: // Network Appearance (Optional)
		*m.na, e = readUint32(r, l)
	// case 0x0007:	// Diagnostic Info (Optional)
	//	m.info = make([]byte, l)
	//	_, e = r.Read(m.info)
	//	if e == nil && l%4 != 0 {
	//		_, e = r.Seek(int64(4-l%4), io.SeekCurrent)
	//	}
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

const (
	InvalidVersion                 uint32 = 0x01
	UnsupportedMessageClass        uint32 = 0x03
	UnsupportedMessageType         uint32 = 0x04
	UnsupportedTrafficHandlingMode uint32 = 0x05
	UnexpectedMessage              uint32 = 0x06
	ProtocolError                  uint32 = 0x07
	InvalidStreamID                uint32 = 0x09
	Refused                        uint32 = 0x0d
	ASPIDRequired                  uint32 = 0x0e
	InvalidASPID                   uint32 = 0x0f
	InvalidParameterValue          uint32 = 0x11
	ParameterFieldError            uint32 = 0x12
	UnexpectedParameter            uint32 = 0x13
	DestinationStatusUnknown       uint32 = 0x14
	InvalidNetworkAppearance       uint32 = 0x15
	MissingParameter               uint32 = 0x16
	InvalidRoutingContext          uint32 = 0x19
	NoConfiguredASforASP           uint32 = 0x1a
	SubsystemStatusUnknown         uint32 = 0x1b
	InvalidLoadsharingLabel        uint32 = 0x1c
)

/*
NTFY is Notify message. (Message type = 0x01)

	 0                     1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|          Tag = 0x000D         |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                         * Status                              |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|            Tag = 0x0011       |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                        ASP Identifier                         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-
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
type NTFY struct {
	status Status
	// id     *uint32
	ctx uint32
	// info    string
}

type Status uint32

const (
	Down     Status = 0x00010001
	Inactive Status = 0x00010002
	Active   Status = 0x00010003
	Pending  Status = 0x00010004

	InsufficientASPResourcesActive Status = 0x00020001
	AlternateASPActive             Status = 0x00020002
	ASPFialer                      Status = 0x00020003
)

func (s Status) String() string {
	switch s {
	case Down:
		return "down"
	case Inactive:
		return "inactive"
	case Active:
		return "active"
	case Pending:
		return "pending"
	}
	return ""
}

func (m *NTFY) handleMessage(a *ASP) {
	switch m.status {
	case Down, Inactive, Active, Pending:
		a.state = m.status
	}
}

func (m *NTFY) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	//	b []byte) (e error) {
	switch t {
	case 0x000D: // Status
		var tmp uint32
		tmp, e = readUint32(r, l)
		m.status = Status(tmp)
	// case 0x0011: // ASP Identifier (Optional)
	//	*(m.id), e = readUint32(r, l)
	case 0x0006: // Routing Context (Optional)
		m.ctx, e = readUint32(r, l)
	// case 0x0004: // Info String (Optional)
	// 	m.info, e = readInfo(r, l)
	default:
		_, e = r.Seek(int64(l), io.SeekCurrent)
	}
	return
}

// 0x02 TEI Status Request
// 0x03 TEI Status Confirm
// 0x04 TEI Status Indication
