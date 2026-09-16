package xua

import (
	"bytes"
	"fmt"
	"io"
)

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

/*
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
type UnitData struct {
	ReturnOnError bool
	ProtocolClass uint8
	Cause         Cause
	CgPA          SCCPAddr
	CdPA          SCCPAddr
	Data          []byte
}

func (u *UnitData) marshal() []byte {
	ud := new(bytes.Buffer)
	if u.Cause == Success {
		ud.WriteByte(0x09)
		if u.ReturnOnError {
			ud.WriteByte((u.ProtocolClass & 0x0f) | 0x80)
		} else {
			ud.WriteByte(u.ProtocolClass & 0x0f)
		}
	} else {
		ud.WriteByte(0x0a)
		ud.WriteByte(byte(u.Cause & 0x00ff))
	}

	ud.WriteByte(3)
	cdpa := u.CdPA.marshalSCCP()
	ud.WriteByte(byte(3 + len(cdpa)))
	cgpa := u.CgPA.marshalSCCP()
	ud.WriteByte(byte(3 + len(cdpa) + len(cgpa)))

	ud.WriteByte(byte(len(cdpa)))
	ud.Write(cdpa)
	ud.WriteByte(byte(len(cgpa)))
	ud.Write(cgpa)
	ud.WriteByte(byte(len(u.Data)))
	ud.Write(u.Data)

	return ud.Bytes()
}

func (u *UnitData) unmarshal(d []byte) error {
	buf := bytes.NewReader(d)
	if t, e := buf.ReadByte(); e != nil {
		return e
	} else {
		switch t {
		case 0x09:
			if t, e = buf.ReadByte(); e != nil {
				return e
			}
			u.ReturnOnError = t&0x80 == 0x80
			u.ProtocolClass = t & 0x0f
		case 0x0a:
			if t, e = buf.ReadByte(); e != nil {
				return e
			}
			u.Cause = Cause(t) | 0x0100
		default:
			return fmt.Errorf("unknown SCCP message type(%x)", t)
		}
	}

	pos := make([]byte, 3)
	_, e := buf.Read(pos)
	if e != nil {
		return e
	}
	d = make([]byte, buf.Len())
	buf.Read(d)
	buf.Reset(d)

	if _, e := buf.Seek(int64(pos[0]-3), io.SeekStart); e != nil {
		return e
	}
	if u.CdPA, e = readSCCPAddr(buf); e != nil {
		return e
	}
	if _, e := buf.Seek(int64(pos[1]-2), io.SeekStart); e != nil {
		return e
	}
	if u.CgPA, e = readSCCPAddr(buf); e != nil {
		return e
	}
	if _, e := buf.Seek(int64(pos[2]-1), io.SeekStart); e != nil {
		return e
	}
	if t, e := buf.ReadByte(); e != nil {
		return e
	} else {
		u.Data = make([]byte, t)
		_, e = buf.Read(u.Data)
		return e
	}
}
