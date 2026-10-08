package tcap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
ABRT is abort dialogue.

	ABRT-apdu ::= [APPLICATION 4] IMPLICIT SEQUENCE {
		abort-source     [0]  IMPLICIT ABRT-source,
		user-information [30] IMPLICIT SEQUENCE OF EXTERNAL OPTIONAL }
*/
type ABRT struct {
	Source Source `json:"abort-source"`
	// info UserInformation
}

func (d ABRT) String() string {
	buf := new(strings.Builder)
	fmt.Fprint(buf, "ABRT")
	fmt.Fprintf(buf, "\n%sabort-source: %s", gsmap.LogPrefix, d.Source)
	return buf.String()
}

/*
Source of ABRT.

	ABRT-source ::= INTEGER {
		dialogue-service-user     (0),
		dialogue-service-provider (1) }
*/
type Source byte

const (
	SvcUser     Source = 0x00
	SvcProvider Source = 0x01
)

func (s Source) String() string {
	switch s {
	case SvcUser:
		return "dialogue-service-user"
	case SvcProvider:
		return "dialogue-service-provider"
	}
	return fmt.Sprintf("unknown(%x)", byte(s))
}

func (s *Source) UnmarshalJSON(b []byte) (e error) {
	var t string
	e = json.Unmarshal(b, &t)
	switch t {
	case "dialogue-service-user":
		*s = SvcUser
	case "dialogue-service-provider":
		*s = SvcProvider
	default:
		*s = 0
	}
	return
}

func (s Source) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (d *ABRT) marshalDialogue() []byte {
	buf := new(bytes.Buffer)

	// abort-source, context_specific(80) + primitive(00) + 0(00)
	// universal Integer
	b := gsmap.WriteTLV(buf, 0x80, []byte{byte(d.Source)})

	// user-information, context_specific(80) + constructed(20) + 30(1e)

	// Dialogue, application(40) + constructed(20) + 4(04)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x64, b)
}

func (d *ABRT) unmarshalDialogue(b []byte) error {
	buf := bytes.NewBuffer(b)

	// abort-source, context_specific(80) + primitive(00) + 0(00)
	// universal Integer
	if _, v, e := gsmap.ReadTLV(buf, 0x80); e != nil {
		return e
	} else if len(v) != 1 || v[0] > 1 {
		return gsmap.UnexpectedTLV("invalid parameter value")
	} else {
		d.Source = Source(v[0])
	}

	return nil
}
