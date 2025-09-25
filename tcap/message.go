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

/*
TcBegin message.

	Begin ::= SEQUENCE {
		otid            OrigTransactionID,
		dialoguePortion DialoguePortion    OPTIONAL,
		components      ComponentPortion   OPTIONAL }

	OrigTransactionID ::= [APPLICATION 8] IMPLICIT OCTET STRING (SIZE (1..4) )
*/
type TcBegin struct {
	otid      uint32
	dialogue  Dialogue
	component []gsmap.Component
}

func (m TcBegin) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "TC-BEGIN (otid=%x)", m.otid)
	if m.dialogue != nil {
		fmt.Fprint(buf, "\n | dialoguePortion:", m.dialogue)
	}
	for i, c := range m.component {
		fmt.Fprintf(buf, "\n | component[%d]: %s", i, c)
	}
	return buf.String()
}

func (m *TcBegin) marshalTc() []byte {
	buf := new(bytes.Buffer)

	// otid, application(40) + primitive(00) + 8(08)
	buf.Write(marshalTid(0x48, m.otid))

	marshalDialogueAndComponents(buf, m.dialogue, m.component)

	// TcBegin, application(40) + constructed(20) + 2(02)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x62, buf.Bytes())
}

func unmarshalTcBegin(data []byte) (m *TcBegin, e error) {
	m = &TcBegin{}
	buf := bytes.NewBuffer(data)

	// otid, application(40) + primitive(00) + 8(08)
	if m.otid, e = unmarshalTid(buf, 0x48); e != nil {
		return
	}

	m.dialogue, m.component, e = unmarshalDialogueAndComponents(buf)
	return
}

func (m TcBegin) Components() []gsmap.Component {
	return m.component
}

/*
TcEnd message.

	End ::= SEQUENCE {
		dtid            DestTransactionID,
		dialoguePortion DialoguePortion    OPTIONAL,
		components      ComponentPortion   OPTIONAL }

	DestTransactionID ::= [APPLICATION 9] IMPLICIT OCTET STRING (SIZE (1..4) )
*/
type TcEnd struct {
	dtid      uint32
	dialogue  Dialogue
	component []gsmap.Component
}

func (m TcEnd) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "TC-END (dtid=%x)", m.dtid)
	if m.dialogue != nil {
		fmt.Fprint(buf, "\n | dialoguePortion:", m.dialogue)
	}
	for i, c := range m.component {
		fmt.Fprintf(buf, "\n | component[%d]: %s", i, c)
	}
	return buf.String()
}

func (m *TcEnd) marshalTc() []byte {
	buf := new(bytes.Buffer)

	// dtid, application(40) + primitive(00) + 9(09)
	buf.Write(marshalTid(0x49, m.dtid))

	marshalDialogueAndComponents(buf, m.dialogue, m.component)

	// TcEnd, application(40) + constructed(20) + 4(04)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x64, buf.Bytes())
}

func unmarshalTcEnd(data []byte) (m *TcEnd, e error) {
	m = &TcEnd{}
	buf := bytes.NewBuffer(data)

	// dtid, application(40) + primitive(00) + 9(09)
	if m.dtid, e = unmarshalTid(buf, 0x49); e != nil {
		return
	}

	m.dialogue, m.component, e = unmarshalDialogueAndComponents(buf)
	return
}

func (m TcEnd) Components() []gsmap.Component {
	return m.component
}

/*
TcContinue

	Continue ::= SEQUENCE {
		otid            OrigTransactionID,
		dtid            DestTransactionID,
		dialoguePortion DialoguePortion    OPTIONAL,
		components      ComponentPortion   OPTIONAL }
*/
type TcContinue struct {
	otid      uint32
	dtid      uint32
	dialogue  Dialogue
	component []gsmap.Component
}

func (m TcContinue) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "TC-CONTINUE (otid=%x, dtid=%x)", m.otid, m.dtid)
	if m.dialogue != nil {
		fmt.Fprint(buf, "\n | dialoguePortion:", m.dialogue)
	}
	for i, c := range m.component {
		fmt.Fprintf(buf, "\n | component[%d]: %s", i, c)
	}
	return buf.String()
}

func (m *TcContinue) marshalTc() []byte {
	buf := new(bytes.Buffer)

	// otid, application(40) + primitive(00) + 8(08)
	buf.Write(marshalTid(0x48, m.otid))

	// dtid, application(40) + primitive(00) + 9(09)
	buf.Write(marshalTid(0x49, m.dtid))

	marshalDialogueAndComponents(buf, m.dialogue, m.component)

	// TcBegin, application(40) + constructed(20) + 5(05)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x65, buf.Bytes())
}

func unmarshalTcContinue(data []byte) (m *TcContinue, e error) {
	m = &TcContinue{}
	buf := bytes.NewBuffer(data)

	// otid, application(40) + primitive(00) + 8(08)
	if m.otid, e = unmarshalTid(buf, 0x48); e != nil {
		return
	}

	// dtid, application(40) + primitive(00) + 9(09)
	if m.dtid, e = unmarshalTid(buf, 0x49); e != nil {
		return
	}

	m.dialogue, m.component, e = unmarshalDialogueAndComponents(buf)
	return
}

func (m TcContinue) Components() []gsmap.Component {
	return m.component
}

/*
TcAbort

	Abort ::= SEQUENCE {
		dtid   DestTransactionID,
		reason CHOICE {
			p-abortCause P-AbortCause,
			u-abortCause DialoguePortion } OPTIONAL }

The u-abortCause may be generated by the component sublayer in which case it is an ABRT APDU,
or by the TC-User in which case it could be either an ABRT APDU or data in some user-defined abstract syntax.
*/
type TcAbort struct {
	dtid   uint32
	pCause Cause
	uCause Dialogue
}

func (m TcAbort) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "TC-ABORT (dtid=%x)", m.dtid)
	if m.uCause != nil {
		fmt.Fprint(buf, "\n | u-abortCause:", m.uCause)
	} else {
		fmt.Fprint(buf, "\n | ", m.pCause)
	}
	return buf.String()
}

func (m TcAbort) Error() string {
	if m.uCause != nil {
		return "Aborted by peer with u-abortCause"
	}
	if m.pCause < 0x10 {
		return fmt.Sprint("Aborted by peer with ", m.pCause)
	}
	return fmt.Sprint("Internal error with ", m.pCause)
}

/*
Cause

	P-AbortCause ::= [APPLICATION 10] IMPLICIT INTEGER {
		unrecognizedMessageType          (0),
		unrecognizedTransactionID        (1),
		badlyFormattedTransactionPortion (2),
		incorrectTransactionPortion      (3),
		resourceLimitation               (4) }
*/
type Cause byte

const (
	TcUnrecognizedMessageType          Cause = 0x00
	TcUnrecognizedTransactionID        Cause = 0x01
	TcBadlyFormattedTransactionPortion Cause = 0x02
	TcIncorrectTransactionPortion      Cause = 0x03
	TcResourceLimitation               Cause = 0x04

	TcTimeout       Cause = 0x10
	TcNoDestination Cause = 0x11
	TcDiscard       Cause = 0x12
)

func (c Cause) String() string {
	switch c {
	case TcUnrecognizedMessageType:
		return "p-abortCause: unrecognizedMessageType"
	case TcUnrecognizedTransactionID:
		return "p-abortCause: unrecognizedTransactionID"
	case TcBadlyFormattedTransactionPortion:
		return "p-abortCause: badlyFormattedTransactionPortion"
	case TcIncorrectTransactionPortion:
		return "p-abortCause: incorrectTransactionPortion"
	case TcResourceLimitation:
		return "p-abortCause: resourceLimitation"
	case TcTimeout:
		return "responseTimeout(internal)"
	case TcNoDestination:
		return "noDestinationFound(internal)"
	case TcDiscard:
		return "discard(internal)"
	default:
		return fmt.Sprintf("p-abortCause: unknown(%x)", byte(c))
	}
}

func (m *TcAbort) marshalTc() []byte {
	buf := new(bytes.Buffer)

	// dtid, application(40) + primitive(00) + 9(09)
	buf.Write(marshalTid(0x49, m.dtid))

	// p-AbortCause, application(40) + primitive(00) + 10(0a)
	if m.uCause == nil {
		gsmap.WriteTLV(buf, 0x4a, []byte{byte(m.pCause)})
	}

	// u-abortCause, application(40) + constructed(20) + 11(0b)
	if m.uCause != nil {
		gsmap.WriteTLV(buf, 0x6b, marshalDialogue(m.uCause))
	}

	// TcEnd, application(40) + constructed(20) + 4(07)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x67, buf.Bytes())
}

func unmarshalTcAbort(data []byte) (m *TcAbort, e error) {
	m = &TcAbort{}
	buf := bytes.NewBuffer(data)

	// dtid, application(40) + primitive(00) + 9(09)
	if m.dtid, e = unmarshalTid(buf, 0x49); e != nil {
		return
	}

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		e = nil
	} else if e != nil {
		return
	}
	switch t {
	case 0x6b: // u-abortCause, application(40) + constructed(20) + 11(0b)
		m.uCause, e = unmarshalDialogue(v)
	case 0x4a: // p-AbortCause, application(40) + primitive(00) + 10(0a)
		if len(v) != 1 || v[0] > 4 {
			e = gsmap.UnexpectedTLV("invalid parameter value")
		} else {
			m.pCause = Cause(v[0])
		}
	default:
		e = gsmap.UnexpectedTag([]byte{0x6b, 0x4a}, t)
	}
	return
}

func (m TcAbort) Components() []gsmap.Component {
	return nil
}

func (m TcAbort) Cause() (Cause, Dialogue) {
	return m.pCause, m.uCause
}
