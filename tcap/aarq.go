package tcap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
AARQ is request dialogue.

	AARQ-apdu ::= [APPLICATION 0] IMPLICIT SEQUENCE {
		protocol-version         [0]  IMPLICIT BIT STRING { version1 (0) } DEFAULT { version1 },
		application-context-name [1]  OBJECT IDENTIFIER,
		user-information         [30] IMPLICIT SEQUENCE OF EXTERNAL OPTIONAL }
*/
type AARQ struct {
	Context gsmap.AppContext `json:"application-context-name"`
	// info UserInformation
}

func (d AARQ) String() string {
	buf := new(strings.Builder)
	fmt.Fprint(buf, "AARQ")
	fmt.Fprintf(buf, "\n%sprotocol-version:         version1", gsmap.LogPrefix)
	fmt.Fprintf(buf, "\n%sapplication-context-name: %x", gsmap.LogPrefix, d.Context)
	return buf.String()
}

func (d AARQ) MarshalJSON() ([]byte, error) {
	j := map[string]any{}
	j["AARQ"] = struct {
		Ver string `json:"protocol-version"`
		AARQ
	}{
		Ver:  "version1",
		AARQ: d}
	return json.Marshal(j)
}

func (d *AARQ) marshalDialogue() []byte {
	buf := new(bytes.Buffer)

	// protocol-version, context_specific(80) + primitive(00) + 0(00)
	// value = v1 (0x07 80)
	gsmap.WriteTLV(buf, 0x80, []byte{0x07, 0x80})

	// application-context-name, context_specific(80) + constructed(20) + 1(01)
	// OBJECT IDENTIFIER, universal(00) + primitive(00) + OID(06)
	b := gsmap.WriteTLV(buf, 0xa1,
		gsmap.WriteTLV(new(bytes.Buffer), 0x06, d.Context.Marshal()))

	// user-information, context_specific(80) + constructed(20) + 30(1e)

	// AARQ-apdu, application(40) + constructed(20) + 0(00)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x60, b)
}

func (d *AARQ) unmarshalDialogue(b []byte) error {
	buf := bytes.NewBuffer(b)

	// protocol-version, context_specific(80) + primitive(00) + 0(00)
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e != nil {
		return e
	} else if t == 0x80 {
		if len(v) != 2 || v[0] != 0x07 || v[1] != 0x80 {
			return errors.New("unknown version")
		}
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e != nil {
			return e
		}
	}

	// application-context-name, context_specific(80) + constructed(20) + 1(00)
	// OBJECT IDENTIFIER, universal(00) + primitive(00) + OID(06)
	if t == 0xa1 {
		if _, v, e = gsmap.ReadTLV(bytes.NewBuffer(v), 0x06); e != nil {
			return e
		}
		d.Context.Unmarshal(v)
	} else {
		return gsmap.UnexpectedTag([]byte{0xa1}, t)
	}

	// user-information, context_specific(80) + constructed(20) + 30(1e)

	return nil
}
