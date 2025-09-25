package gsmap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

/*
	Universal       byte = 0x00
	Application     byte = 0x40
	ContextSpecific byte = 0x80
	Private         byte = 0xc0

	Primitive   byte = 0x00
	Constructed byte = 0x20

	Boolean     byte = 0x01
	Integer     byte = 0x02
	BitString   byte = 0x03
	OctetString byte = 0x04
	Null        byte = 0x05
	OID         byte = 0x06
	External    byte = 0x08
	Enum        byte = 0x0a
	UTF8String  byte = 0x0c
	Sequence    byte = 0x10
	Set         byte = 0x11
*/

func WriteTLV(w *bytes.Buffer, t byte, v []byte) []byte {
	w.WriteByte(t)
	if l := len(v); l < 128 {
		w.Write([]byte{byte(l)})
	} else if l <= 0xff {
		w.Write([]byte{0x81, byte(l)})
	} else if l <= 0xffff {
		w.Write([]byte{0x82,
			byte((l & 0xff00) >> 8),
			byte(l & 0x00ff)})
	} else if l <= 0xffffff {
		w.Write([]byte{0x83,
			byte((l & 0xff0000) >> 16),
			byte((l & 0x00ff00) >> 8),
			byte(l & 0x0000ff)})
	} else if l <= 0xffffffff {
		w.Write([]byte{0x84,
			byte((l & 0xff000000) >> 24),
			byte((l & 0x00ff0000) >> 16),
			byte((l & 0x0000ff00) >> 8),
			byte(l & 0x000000ff)})
	} else if l <= 0xffffffffff {
		w.Write([]byte{0x85,
			byte((l & 0xff00000000) >> 32),
			byte((l & 0x00ff000000) >> 24),
			byte((l & 0x0000ff0000) >> 16),
			byte((l & 0x000000ff00) >> 8),
			byte(l & 0x00000000ff)})
	} else {
		panic("failed to packing ASN.1 TLV: too large data")
	}
	if len(v) != 0 {
		w.Write(v)
	}
	return w.Bytes()
}

func ReadTLV(r *bytes.Buffer, tag byte) (t byte, v []byte, e error) {
	if t, e = r.ReadByte(); e != nil {
		return
	}
	if t&0x1f == 0x1f {
		e = unexpectedTLV("invalid lengh of tag", 2)
		return
	}
	if tag != 0x00 && tag != t {
		e = unexpectedTLV(
			fmt.Sprintf("expected tags are [%#x] but %#x", tag, t), 2)
		return
	}

	var b byte
	if b, e = r.ReadByte(); e == io.EOF {
		e = unexpectedTLV("no lengh info", 2)
		return
	} else if e != nil {
		return
	}
	l := int(b)
	if b&0x80 == 0x80 {
		buf := make([]byte, b&0x7f)
		if l, e = r.Read(buf); e == io.EOF {
			e = unexpectedTLV("invalid lengh info", 2)
			return
		} else if e != nil {
			return
		} else if l != len(buf) {
			e = unexpectedTLV("invalid lengh info", 2)
			return
		}
		l = 0
		for _, b := range buf {
			l = (l << 8) | int(b)
		}
	}

	v = make([]byte, l)
	if l, e = r.Read(v); e == io.EOF {
		e = unexpectedTLV("invalid value", 2)
	} else if e == nil && l != len(v) {
		e = unexpectedTLV("invalid value", 2)
	}
	return
}

func UnmarshalIntenger(data []byte) int {
	i, _ := binary.Varint(data)
	return int(i)
}

/*
func ReadLongTLV(r *bytes.Buffer, tag []byte) (t []byte, v []byte, e error) {
	var b byte
	if b, e = r.ReadByte(); e != nil {
		return
	}
	t = []byte{b}
	if b&0x1f == 0x1f {
		for {
			if b, e = r.ReadByte(); e != nil {
				return
			}
			t = append(t, b)
			if b&0x80 != 0x80 {
				break
			}
		}
	}
	if tag != nil {
		if len(tag) != len(t) {
			e = InvalidTagError{Expected: tag, Actual: t[0]}
			return
		}
		for i := range tag {
			if tag[i] != t[i] {
				e = InvalidTagError{Expected: tag, Actual: t[0]}
				return
			}
		}
	}

	if b, e = r.ReadByte(); e == io.EOF {
		e = InvalidStructureError("no lengh info")
		return
	} else if e != nil {
		return
	}
	l := int(b)
	if b&0x80 == 0x80 {
		buf := make([]byte, b&0x7f)
		if l, e = r.Read(buf); e == io.EOF {
			e = InvalidStructureError("invalid lengh info")
			return
		} else if e != nil {
			return
		} else if l != len(buf) {
			e = InvalidStructureError("invalid lengh info")
			return
		}
		l = 0
		for _, b := range buf {
			l = (l << 8) | int(b)
		}
	}

	v = make([]byte, l)
	if l, e = r.Read(v); e == io.EOF {
		e = InvalidStructureError("invalid value")
	} else if e == nil && l != len(v) {
		e = InvalidStructureError("invalid value")
	}

	return
}

func MarshalBitString(bits ...bool) []byte {
	p := 8 - (len(bits) % 8)
	if p == 8 {
		p = 0
	}

	ret := make([]byte, (len(bits)+p)/8+1)
	ret[0] = byte(p)
	for i, b := range bits {
		if !b {
			continue
		}
		switch i % 8 {
		case 0:
			ret[1+(i/8)] |= 0x80
		case 1:
			ret[1+(i/8)] |= 0x40
		case 2:
			ret[1+(i/8)] |= 0x20
		case 3:
			ret[1+(i/8)] |= 0x10
		case 4:
			ret[1+(i/8)] |= 0x08
		case 5:
			ret[1+(i/8)] |= 0x04
		case 6:
			ret[1+(i/8)] |= 0x02
		case 7:
			ret[1+(i/8)] |= 0x01
		}
	}
	return ret
}

func UnmarshalBitString(data []byte) []bool {
	if len(data) < 3 || data[0] > 7 {
		return nil
	}
	ret := make([]bool, 8*len(data[1:])-int(data[0]))
	for o, d := range data[1:] {
		o = o * 8
		if d&0x80 == 0x80 {
			ret[o] = true
		}
		o++
		if d&0x40 == 0x40 && len(ret) > o {
			ret[o] = true
		}
		o++
		if d&0x20 == 0x20 && len(ret) > o {
			ret[o] = true
		}
		o++
		if d&0x10 == 0x10 && len(ret) > o {
			ret[o] = true
		}
		o++
		if d&0x08 == 0x08 && len(ret) > o {
			ret[o] = true
		}
		o++
		if d&0x04 == 0x04 && len(ret) > o {
			ret[o] = true
		}
		o++
		if d&0x02 == 0x02 && len(ret) > o {
			ret[o] = true
		}
		o++
		if d&0x01 == 0x01 && len(ret) > o {
			ret[o] = true
		}
	}
	return ret
}
*/
