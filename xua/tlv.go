package xua

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

/*
func writeInfo(w io.Writer, info string) {
	d := []byte(info)
	if len(d) > 248 {
		d = d[:248]
	}
	binary.Write(w, binary.BigEndian, uint16(0x0004))
	binary.Write(w, binary.BigEndian, uint16(4+len(d)))
	w.Write(d)
	if len(d)%4 != 0 {
		w.Write(make([]byte, 4-len(d)%4))
	}
}

func readInfo(r io.ReadSeeker, l uint16) (v string, e error) {
	d := make([]byte, l)
	_, e = r.Read(d)
	if e == nil && l%4 != 0 {
		_, e = r.Seek(int64(4-l%4), io.SeekCurrent)
	}
	v = string(d)
	return
}

func writeRoutingContext(w io.Writer, cx []uint32) {
	binary.Write(w, binary.BigEndian, uint16(0x0006))
	binary.Write(w, binary.BigEndian, uint16(4+len(cx)*4))
	for _, c := range cx {
		binary.Write(w, binary.BigEndian, c)
	}
}

func readRoutingContext(r io.ReadSeeker, l uint16) (v []uint32, e error) {
	if l%4 != 0 {
		e = errors.New("invalid length of parameter")
	} else {
		v = make([]uint32, l/4)
		for i := range v {
			if e = binary.Read(r, binary.BigEndian, &(v[i])); e != nil {
				return
			}
		}
	}
	return
}
*/

type PointCode struct {
	mask byte
	pc   uint32
}

func (p PointCode) String() string {
	return fmt.Sprintf("%d/%d", p.pc, p.mask)
}

func writeAPC(w io.Writer, v []PointCode) {
	binary.Write(w, binary.BigEndian, uint16(0x0012))
	binary.Write(w, binary.BigEndian, uint16(4+4*len(v)))
	for _, a := range v {
		w.Write([]byte{a.mask, byte(a.pc >> 16), byte(a.pc >> 8), byte(a.pc)})
	}
}

func readAPC(r io.ReadSeeker, l uint16) (v []PointCode, e error) {
	if l%4 != 0 {
		e = errors.New("invalid length of parameter")
	} else {
		v = make([]PointCode, l/4)
		for i := range v {
			if e = binary.Read(r, binary.BigEndian, &(v[i].pc)); e != nil {
				break
			}
			v[i].mask = byte(v[i].pc >> 24)
			v[i].pc = v[i].pc & 0x0fff
		}
	}
	return
}

func writeUint32(w io.Writer, t uint16, v uint32) {
	binary.Write(w, binary.BigEndian, t)
	binary.Write(w, binary.BigEndian, uint16(8))
	binary.Write(w, binary.BigEndian, v)
}

func readUint32(r io.ReadSeeker, l uint16) (v uint32, e error) {
	if l != 4 {
		e = errors.New("invalid length of parameter")
	} else {
		e = binary.Read(r, binary.BigEndian, &v)
	}
	return
}

func writeUint8(w io.Writer, t uint16, v uint8) {
	binary.Write(w, binary.BigEndian, t)
	binary.Write(w, binary.BigEndian, uint16(8))
	binary.Write(w, binary.BigEndian, uint32(v))
}

func readUint8(r io.ReadSeeker, l uint16) (v uint8, e error) {
	if l != 4 {
		e = errors.New("invalid length of parameter")
	} else if _, e = r.Seek(3, io.SeekCurrent); e == nil {
		e = binary.Read(r, binary.BigEndian, &v)
	}
	return
}

func writeData(w io.Writer, d []byte) {
	binary.Write(w, binary.BigEndian, uint16(0x010B))
	binary.Write(w, binary.BigEndian, uint16(4+len(d)))
	w.Write(d)
	if len(d)%4 != 0 {
		w.Write(make([]byte, 4-len(d)%4))
	}
}

func readData(r io.ReadSeeker, l uint16) (d []byte, e error) {
	d = make([]byte, l)
	_, e = r.Read(d)
	return
}
