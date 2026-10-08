package tcap

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
)

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
