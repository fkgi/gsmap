package xua

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/fkgi/teldata"
)

/*
SCCPAddr is address of SCCP

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|      Routing Indicator        |       Address Indicator       |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	/                       Address parameter(s)                    /
	\                                                               \
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Global Title

	0                 1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8001          |            Length             |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                Reserved                       |      GTI      |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|   No. Digits  | Trans. type   |    Num. Plan  | Nature of Add |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|2 addr.|1 addr.|4 addr.|3 addr.|6 addr.|5 addr.|8 addr.|7 addr.|
	|  sig. | sig.  |  sig. | sig.  |  sig. | sig.  |  sig. | sig.  |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|        .............          |filler |N addr.|   filler      |
	|                               |if req | sig.  |               |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Point Code

	0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8002          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                            Point Code                         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Subsystem Number

	0                   1                   2                   3
	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|         Tag = 0x8003          |            Length = 8         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                 Reserved                      |   SSN value   |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Translation type

	0              Unknown
	1 - 63         International services
	64 - 127       Spare
	128 - 254      National network specific
	255            Reserved
*/
type SCCPAddr struct {
	// RoutingIndicator
	// GlobalTitleIndicator

	GlobalTitle     teldata.GlobalTitle     `json:"gt,omitempty"`
	PointCode       uint32                  `json:"pc,omitempty"`
	SubsystemNumber teldata.SubsystemNumber `json:"ssn,omitempty"`
}

func (a *SCCPAddr) marshalSUA() []byte {
	buf := new(bytes.Buffer)

	var ai uint16
	if len(a.GlobalTitle.Digits) != 0 {
		binary.Write(buf, binary.BigEndian, uint16(0x8001))

		dig := a.GlobalTitle.Digits.Bytes()
		for i := range dig {
			if 0xf0&dig[i] == 0xf0 {
				dig[i] = 0x0f & dig[i]
			}
			if 0x0f&dig[i] == 0x0f {
				dig[i] = 0xf0 & dig[i]
			}
		}
		l := len(dig)
		if l%4 != 0 {
			dig = append(dig, make([]byte, 4-l%4)...)
			// l = len(dig)
		}
		l += 12
		binary.Write(buf, binary.BigEndian, uint16(l))

		var gti uint32 = 4 // TT, NPI and NAI
		if a.GlobalTitle.TranslationType == 0 &&
			a.GlobalTitle.NumberingPlan == teldata.UnknownNP {
			gti = 1 // NAI only
		} else if a.GlobalTitle.NatureOfAddress == teldata.UnknownNA &&
			a.GlobalTitle.NumberingPlan == teldata.UnknownNP {
			gti = 2 // TT only
		} else if a.GlobalTitle.NatureOfAddress == teldata.UnknownNA {
			gti = 3 // TT and NPI
		}
		binary.Write(buf, binary.BigEndian, gti)

		buf.WriteByte(byte(a.GlobalTitle.Digits.Length()))
		buf.WriteByte(a.GlobalTitle.TranslationType)

		buf.WriteByte(byte(a.GlobalTitle.NumberingPlan))

		switch a.GlobalTitle.NatureOfAddress {
		case teldata.International:
			buf.WriteByte(0x04)
		case teldata.NationalSignificant:
			buf.WriteByte(0x03)
		case teldata.NetworkSpecific:
			buf.WriteByte(0x02)
		case teldata.Subscriber:
			buf.WriteByte(0x01)
		default:
			buf.WriteByte(0x00)
		}

		buf.Write(dig)

		ai |= 0x04
	}
	if a.PointCode != 0 {
		writeUint32(buf, 0x8002, a.PointCode)
		ai |= 0x02
	}
	if a.SubsystemNumber != 0 {
		writeUint8(buf, 0x8003, a.SubsystemNumber.Uint())
		ai |= 0x01
	}

	var ri uint16 = 1 // Rout on GT
	if a.PointCode != 0 && a.SubsystemNumber != 0 {
		ri = 2 // Route on PC+SSN
	}

	d := buf.Bytes()
	buf = new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, ri)
	binary.Write(buf, binary.BigEndian, ai)
	buf.Write(d)

	return buf.Bytes()
}

func (a *SCCPAddr) marshalSCCP() []byte {
	buf := new(bytes.Buffer)

	var ai byte = 0x00
	buf.WriteByte(ai)
	if a.PointCode != 0 {
		binary.Write(buf, binary.LittleEndian, uint16(a.PointCode))
		ai |= 0x01
	}
	if a.SubsystemNumber != 0 {
		buf.WriteByte(a.SubsystemNumber.Uint())
		ai |= 0x02
	}
	if a.PointCode != 0 && a.SubsystemNumber != 0 {
		ai |= 0x40
	}

	if !a.GlobalTitle.IsEmpty() {
		isOdd := a.GlobalTitle.Digits[len(a.GlobalTitle.Digits)-1]&0xf0 == 0xf0
		var nai byte = 0x00
		switch a.GlobalTitle.NatureOfAddress {
		case teldata.International:
			nai = 0x04
		case teldata.NationalSignificant:
			nai = 0x03
		case teldata.NetworkSpecific:
			nai = 0x02
		case teldata.Subscriber:
			nai = 0x01
		}

		if a.GlobalTitle.TranslationType == 0 &&
			a.GlobalTitle.NumberingPlan == teldata.UnknownNP {
			ai |= 0x04 // NAI only --0001--
			if isOdd {
				buf.WriteByte(0x80 | nai)
			} else {
				buf.WriteByte(nai)
			}

		} else if a.GlobalTitle.NatureOfAddress == teldata.UnknownNA &&
			a.GlobalTitle.NumberingPlan == teldata.UnknownNP {
			ai |= 0x08 // TT only --0010--
			buf.WriteByte(a.GlobalTitle.TranslationType)

		} else if a.GlobalTitle.NatureOfAddress == teldata.UnknownNA {
			ai |= 0x0c // TT and NPI --0011--
			buf.WriteByte(a.GlobalTitle.TranslationType)
			if isOdd {
				buf.WriteByte(byte(a.GlobalTitle.NumberingPlan<<4 | 0x01))
			} else {
				buf.WriteByte(byte(a.GlobalTitle.NumberingPlan<<4 | 0x02))
			}

		} else {
			ai |= 0x10 // TT, NPI and NAI --0100--
			buf.WriteByte(a.GlobalTitle.TranslationType)
			if isOdd {
				buf.WriteByte(byte(a.GlobalTitle.NumberingPlan<<4 | 0x01))
			} else {
				buf.WriteByte(byte(a.GlobalTitle.NumberingPlan<<4 | 0x02))
			}
			buf.WriteByte(nai)
		}

		dig := a.GlobalTitle.Digits.Bytes()
		for i := range dig {
			if 0xf0&dig[i] == 0xf0 {
				dig[i] = 0x0f & dig[i]
			}
			if 0x0f&dig[i] == 0x0f {
				dig[i] = 0xf0 & dig[i]
			}
		}
		buf.Write(dig)
	}

	b := buf.Bytes()
	b[0] = ai
	return b
}

func readSUAAddr(r io.ReadSeeker, l uint16) (a SCCPAddr, e error) {
	if l%4 != 0 {
		e = errors.New("invalid length of parameter")
		return
	}
	var ri, ai uint16
	if e = binary.Read(r, binary.BigEndian, &ri); e != nil {
		return
	}
	if e = binary.Read(r, binary.BigEndian, &ai); e != nil {
		return
	}

	buf := make([]byte, l-4)
	if _, e = r.Read(buf); e != nil {
		return
	}
	rr := bytes.NewReader(buf)
	for rr.Len() > 4 {
		var t, l uint16
		if e = binary.Read(rr, binary.BigEndian, &t); e != nil {
			break
		}
		if e = binary.Read(rr, binary.BigEndian, &l); e != nil {
			break
		}
		l -= 4

		switch t {
		case 0x8001: // GT
			if l < 8 {
				e = errors.New("invalid length of parameter")
				break
			}
			gthdr := make([]byte, 8)
			if _, e = rr.Read(gthdr); e != nil {
				break
			}
			// gti = gthdr[3]
			a.GlobalTitle.TranslationType = gthdr[5]
			a.GlobalTitle.NumberingPlan = teldata.NumberingPlan(gthdr[6])
			switch gthdr[7] {
			case 0x01:
				a.GlobalTitle.NatureOfAddress = teldata.Subscriber
			case 0x02:
				a.GlobalTitle.NatureOfAddress = teldata.NetworkSpecific
			case 0x03:
				a.GlobalTitle.NatureOfAddress = teldata.NationalSignificant
			case 0x04:
				a.GlobalTitle.NatureOfAddress = teldata.International
			default:
				a.GlobalTitle.NatureOfAddress = teldata.UnknownNA
			}

			buf = make([]byte, l-8)
			if _, e = rr.Read(buf); e != nil {
				break
			}
			l = uint16(gthdr[4])
			if l%2 != 0 {
				l = (l + (l % 2)) / 2
				buf[l-1] = buf[l-1] | 0xf0
			} else {
				l = l / 2
			}
			a.GlobalTitle.Digits = buf[:l]
		case 0x8002: // PC
			a.PointCode, e = readUint32(rr, l)
		case 0x8003: // SSN
			var tmp uint8
			if tmp, e = readUint8(rr, l); e == nil {
				a.SubsystemNumber = teldata.ParseSSN(tmp)
			}
		default:
			_, e = r.Seek(int64(l), io.SeekCurrent)
		}
		if e != nil {
			break
		}
	}

	return
}

func readSCCPAddr(buf *bytes.Reader) (a SCCPAddr, e error) {
	var t byte
	if t, e = buf.ReadByte(); e != nil {
		return
	}
	d := make([]byte, t)
	if _, e = buf.Read(d); e != nil {
		return
	}

	ai := d[0]
	buf = bytes.NewReader(d[1:])

	if ai&0x01 == 0x01 {
		var pc uint16
		if e = binary.Read(buf, binary.LittleEndian, &pc); e != nil {
			return
		}
		a.PointCode = uint32(pc)
	}
	if ai&0x02 == 0x02 {
		var sn byte
		if sn, e = buf.ReadByte(); e != nil {
			return
		}
		a.SubsystemNumber = teldata.SubsystemNumber(sn)
	}
	if ai&0x3c != 0 {
		isOdd := false
		var nai byte = 0x00
		switch ai & 0x3c {
		case 0x04:
			if nai, e = buf.ReadByte(); e == nil {
				isOdd = nai&0x80 == 0x80
				nai &= 0x7f
			}
		case 0x08:
			a.GlobalTitle.TranslationType, e = buf.ReadByte()
		case 0x0c:
			var t byte
			if a.GlobalTitle.TranslationType, e = buf.ReadByte(); e != nil {
			} else if t, e = buf.ReadByte(); e == nil {
				a.GlobalTitle.NumberingPlan = teldata.NumberingPlan(t >> 4)
				isOdd = t&0x01 == 0x01
			}
		case 0x10:
			var t byte
			if a.GlobalTitle.TranslationType, e = buf.ReadByte(); e != nil {
			} else if t, e = buf.ReadByte(); e == nil {
				a.GlobalTitle.NumberingPlan = teldata.NumberingPlan(t >> 4)
				isOdd = t&0x01 == 0x01
				nai, e = buf.ReadByte()
			}
		default:
			e = fmt.Errorf("invalid GT Indicator value %x", ai&0x3c)
		}
		if e != nil {
			return
		}
		if a.GlobalTitle.Digits, e = io.ReadAll(buf); e != nil {
			return
		}

		if isOdd {
			a.GlobalTitle.Digits[len(a.GlobalTitle.Digits)-1] = a.GlobalTitle.Digits[len(a.GlobalTitle.Digits)-1] | 0xf0
		}
		switch nai {
		case 0x01:
			a.GlobalTitle.NatureOfAddress = teldata.Subscriber
		case 0x02:
			a.GlobalTitle.NatureOfAddress = teldata.NetworkSpecific
		case 0x03:
			a.GlobalTitle.NatureOfAddress = teldata.NationalSignificant
		case 0x04:
			a.GlobalTitle.NatureOfAddress = teldata.International
		}
	}
	return
}
