package tcap

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
)

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
