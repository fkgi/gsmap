package tcap

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
Message

	TCMessage ::= CHOICE {
		unidirectional [APPLICATION 1] IMPLICIT Unidirectional,
		begin          [APPLICATION 2] IMPLICIT Begin,
		end            [APPLICATION 4] IMPLICIT End,
		continue       [APPLICATION 5] IMPLICIT Continue,
		abort          [APPLICATION 7] IMPLICIT Abort }

	DialoguePortion ::= [APPLICATION 11] EXTERNAL
	ComponentPortion ::= [APPLICATION 12] IMPLICIT SEQUENCE SIZE (1..MAX) OF Component
*/
type Message interface {
	marshalTc() []byte
	Components() []gsmap.Component
	fmt.Stringer
}

func marshalTid(tag byte, tid uint32) []byte {
	return gsmap.WriteTLV(new(bytes.Buffer), tag, []byte{
		byte(0xff & (tid >> 24)),
		byte(0xff & (tid >> 16)),
		byte(0xff & (tid >> 8)),
		byte(0xff & (tid))})
}

func unmarshalTid(buf *bytes.Buffer, tag byte) (tid uint32, e error) {
	_, v, e := gsmap.ReadTLV(buf, tag)
	if e == nil {
		for _, b := range v {
			tid = (tid << 8) | uint32(b)
		}
	}
	return
}

func marshalDialogueAndComponents(buf *bytes.Buffer, d Dialogue, c []gsmap.Component) {
	// dialoguePortion, application(40) + constructed(20) + 11(0b)
	if d != nil {
		gsmap.WriteTLV(buf, 0x6b, marshalDialogue(d))
	}
	// components, application(40) + constructed(20) + 12(0c)
	if len(c) != 0 {
		gsmap.WriteTLV(buf, 0x6c, marshalComponents(c))
	}
}

func unmarshalDialogueAndComponents(buf *bytes.Buffer) (d Dialogue, c []gsmap.Component, e error) {
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		e = nil
		return
	} else if e != nil {
		return
	}

	// dialoguePortion, application(40) + constructed(20) + 11(0b)
	if t == 0x6b {
		if d, e = unmarshalDialogue(v); e != nil {
			return
		}
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			e = nil
			return
		} else if e != nil {
			return
		}
	}

	// components, application(40) + constructed(20) + 12(0c)
	if t == 0x6c {
		if c, e = unmarshalComponents(v); e != nil {
			return
		}
	} else {
		e = gsmap.UnexpectedTag([]byte{0x6c}, t)
	}
	return
}

/*
Unidirectional is not supported

	Unidirectional ::= SEQUENCE{
		dialoguePortion DialoguePortion OPTIONAL,
		components      ComponentPortion }
*/
type Unidirectional struct {
	dialogue  Dialogue
	component []gsmap.Component
}

func (m Unidirectional) String() string {
	buf := new(strings.Builder)
	fmt.Fprint(buf, "UNIDIRECTIONAL")
	if m.dialogue != nil {
		fmt.Fprint(buf, "\n | dialoguePortion:", m.dialogue)
	}
	for i, c := range m.component {
		fmt.Fprintf(buf, "\n | component[%d]: %s", i, c)
	}
	return buf.String()
}

func (m *Unidirectional) marshalTc() []byte {
	buf := new(bytes.Buffer)

	marshalDialogueAndComponents(buf, m.dialogue, m.component)

	// Unidirectional, application(40) + constructed(20) + 1(01)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x61, buf.Bytes())
}

func unmarshalUnidirectional(data []byte) (m *Unidirectional, e error) {
	m = &Unidirectional{}
	buf := bytes.NewBuffer(data)
	m.dialogue, m.component, e = unmarshalDialogueAndComponents(buf)
	return
}

func (m Unidirectional) Components() []gsmap.Component {
	return m.component
}
