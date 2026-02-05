package ipsp

import (
	"bytes"
	"fmt"
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
	code ErrCode
	ctx  uint32
	apc  []PointCode
	na   *uint32
	// info []byte
}

func (m *ERR) handleMessage(c *ASP) {
	if StateNotify != nil {
		ErrorNotify(c.id, m.code)
	}
	c.handleCtrlAns(m)
}

func (m *ERR) handleResult(msg message) {}

func (m *ERR) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)

	// Error Code
	writeUint32(buf, 0x000c, uint32(m.code))

	// Routing Context (Optional)
	if m.ctx != 0 {
		writeUint32(buf, 0x0006, m.ctx)
	}

	// Affected Point Code (Optional)
	if len(m.apc) != 0 {
		writeAPC(buf, m.apc)
	}

	// Network Appearance (Optional)
	if m.na != nil {
		writeUint32(buf, 0x010d, *m.na)
	}

	// Diagnostic Info (Optional)
	// if len(m.info) != 0 {
	//	writeInfo(buf, m.info)
	// }

	return 0x00, 0x00, buf.Bytes()
}

func (m *ERR) unmarshal(t, l uint16, r io.ReadSeeker) (e error) {
	switch t {
	case 0x000C: // Error Code
		var c uint32
		c, e = readUint32(r, l)
		m.code = ErrCode(c)
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

type ErrCode uint32

const (
	InvalidVersion                 ErrCode = 0x01
	UnsupportedMessageClass        ErrCode = 0x03
	UnsupportedMessageType         ErrCode = 0x04
	UnsupportedTrafficHandlingMode ErrCode = 0x05
	UnexpectedMessage              ErrCode = 0x06
	ProtocolError                  ErrCode = 0x07
	InvalidStreamID                ErrCode = 0x09
	Refused                        ErrCode = 0x0d
	ASPIDRequired                  ErrCode = 0x0e
	InvalidASPID                   ErrCode = 0x0f
	InvalidParameterValue          ErrCode = 0x11
	ParameterFieldError            ErrCode = 0x12
	UnexpectedParameter            ErrCode = 0x13
	DestinationStatusUnknown       ErrCode = 0x14
	InvalidNetworkAppearance       ErrCode = 0x15
	MissingParameter               ErrCode = 0x16
	InvalidRoutingContext          ErrCode = 0x19
	NoConfiguredASforASP           ErrCode = 0x1a
	SubsystemStatusUnknown         ErrCode = 0x1b
	InvalidLoadsharingLabel        ErrCode = 0x1c
)

func (c ErrCode) String() string {
	switch c {
	case InvalidVersion:
		return "invalid_version(0x01)"
	case UnsupportedMessageClass:
		return "unsupported_message_class(0x03)"
	case UnsupportedMessageType:
		return "unsupported_message_type(0x04)"
	case UnsupportedTrafficHandlingMode:
		return "unsupported_traffic_handling_mode(0x05)"
	case UnexpectedMessage:
		return "unexpected_message(0x06)"
	case ProtocolError:
		return "protocol_error(0x07)"
	case InvalidStreamID:
		return "invalid_stream_ID(0x09)"
	case Refused:
		return "refused(0x0d)"
	case ASPIDRequired:
		return "ASP_ID_required(0x0e)"
	case InvalidASPID:
		return "invalid_ASP_ID(0x0f)"
	case InvalidParameterValue:
		return "invalid_parameter_value(0x11)"
	case ParameterFieldError:
		return "parameter_field_error(0x12)"
	case UnexpectedParameter:
		return "unexpected_parameter(0x13)"
	case DestinationStatusUnknown:
		return "destination_status_unknown(0x14)"
	case InvalidNetworkAppearance:
		return "invalid_network_appearance(0x15)"
	case MissingParameter:
		return "missing_parameter(0x16)"
	case InvalidRoutingContext:
		return "invalid_routing_context(0x19)"
	case NoConfiguredASforASP:
		return "no_configured_AS_for_ASP(0x1a)"
	case SubsystemStatusUnknown:
		return "subsystem_status_unknown(0x1b)"
	case InvalidLoadsharingLabel:
		return "invalid_loadsharing_label(0x1c)"
	default:
		return fmt.Sprintf("unknown_error(%x)", uint32(c))
	}
}

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

	result chan error
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
	if m.result != nil {
		m.result <- a.sendAnswer(m)
	} else {
		if a.state == m.status {
			return
		}

		if StateNotify != nil {
			StateNotify(a.id, m.status)
		}
		switch m.status {
		case Down, Inactive, Active, Pending:
			a.state = m.status
			a.statNotif <- m.status
		}
	}
}

func (m *NTFY) handleResult(msg message) {}

func (m *NTFY) marshal() (uint8, uint8, []byte) {
	buf := new(bytes.Buffer)
	// Status
	writeUint32(buf, 0x000D, uint32(m.status))

	// ASP Identifier (Optional)
	// if m.id != nil {
	//	writeUint32(buf, 0x0011, *m.id)
	// }

	// Routing Context (Optional)
	if m.ctx != 0 {
		writeUint32(buf, 0x0006, m.ctx)
	}

	// Info String (Optional)
	// if len(m.info) != 0 {
	//	writeInfo(buf, m.info)
	// }
	return 0x01, 0x01, buf.Bytes()
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
