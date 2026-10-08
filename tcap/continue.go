package tcap

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
)

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
