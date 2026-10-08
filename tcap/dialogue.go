package tcap

import (
	"bytes"
	"errors"

	"github.com/fkgi/gsmap"
)

/*
Dialogue

	EXTERNAL ::= [UNIVERSAL 8] IMPLICIT SEQUENCE {
		oid        OBJECT IDENTIFIER,
		dialog [0] EXPLICIT DialoguePDU }
*/
type Dialogue interface {
	marshalDialogue() []byte
	unmarshalDialogue([]byte) error
}

func marshalDialogue(d Dialogue) []byte {
	buf := new(bytes.Buffer)

	// oid, universal(00) + primitive(00) + OID(06)
	// Dialogue-As-ID = 0x00 11 86 05 01 01 01
	gsmap.WriteTLV(buf, 0x06, []byte{0x00, 0x11, 0x86, 0x05, 0x01, 0x01, 0x01})

	// dialog, context_specific(80) + constructed(20) + 0(00)
	b := gsmap.WriteTLV(buf, 0xa0, d.marshalDialogue())

	// ExternalObject, universal(00) + constructed(20) + external(08)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x28, b)
}

func unmarshalDialogue(data []byte) (Dialogue, error) {
	buf := bytes.NewBuffer(data)

	// ExternalObject, universal(00) + constructed(20) + external(08)
	if _, v, e := gsmap.ReadTLV(buf, 0x28); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// oid, universal(00) + primitive(00) + OID(06)
	if _, v, e := gsmap.ReadTLV(buf, 0x06); e != nil {
		return nil, e
	} else if len(v) != 7 ||
		v[0] != 0x00 || v[1] != 0x11 || v[2] != 0x86 ||
		v[3] != 0x05 || v[4] != 0x01 || v[5] != 0x01 || v[6] != 0x01 {
		return nil, errors.New("unknown Object ID")
	}

	// dialog, context_specific(80) + constructed(20) + 0(00)
	var d Dialogue
	if _, v, e := gsmap.ReadTLV(buf, 0xa0); e != nil {
		return nil, e
	} else if t, v, e := gsmap.ReadTLV(bytes.NewBuffer(v), 0x00); e != nil {
		return nil, e
	} else {
		switch t {
		case 0x60: // AARQ, application(40) + constructed(20) + 0(00)
			d = &AARQ{}
		case 0x61: // AARE, application(40) + constructed(20) + 1(01)
			d = &AARE{}
		case 0x64: // ABRT, application(40) + constructed(20) + 4(04)
			d = &ABRT{}
		default:
			return nil, gsmap.UnexpectedTag([]byte{0x60, 0x61, 0x64}, t)
		}
		if e = d.unmarshalDialogue(v); e != nil {
			return nil, e
		}
	}

	return d, nil
}
