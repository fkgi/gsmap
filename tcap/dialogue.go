package tcap

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
)

var DialogueHandler = func(q AARQ) Dialogue {
	return &AARE{
		Context:   q.Context,
		Result:    Accept,
		ResultSrc: SrcUsrNull}
}

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

/*
AARE is response dialogue.

	AARE-apdu ::= [APPLICATION 1] IMPLICIT SEQUENCE {
		protocol-version         [0]  IMPLICIT BIT STRING { version1 (0) } DEFAULT { version1 },
		application-context-name [1]  OBJECT IDENTIFIER,
		result                   [2]  Associate-result,
		result-source-diagnostic [3]  Associate-source-diagnostic,
		user-information         [30] IMPLICIT SEQUENCE OF EXTERNAL OPTIONAL }
*/
type AARE struct {
	Context   gsmap.AppContext `json:"application-context-name"`
	Result    Result           `json:"result"`
	ResultSrc ResultSrc        `json:"result-source-diagnostic"`
	// info UserInformation
}

func (d AARE) String() string {
	buf := new(strings.Builder)
	fmt.Fprint(buf, "AARE")
	fmt.Fprintf(buf, "\n%sprotocol-version:         version1", gsmap.LogPrefix)
	fmt.Fprintf(buf, "\n%sapplication-context-name: %x", gsmap.LogPrefix, d.Context)
	fmt.Fprintf(buf, "\n%sresult:                   %s", gsmap.LogPrefix, d.Result)
	fmt.Fprintf(buf, "\n%sresult-source-diagnostic:", gsmap.LogPrefix)
	switch d.ResultSrc {
	case SrcUsrNull:
		fmt.Fprintf(buf, "\n%s| dialogue-service-user: null", gsmap.LogPrefix)
	case SrcUsrNoReason:
		fmt.Fprintf(buf, "\n%s| dialogue-service-user: no-reason-given", gsmap.LogPrefix)
	case SrcUsrACNameNotSupported:
		fmt.Fprintf(buf, "\n%s| dialogue-service-user: application-context-name-not-supported", gsmap.LogPrefix)
	case SrcPrvNull:
		fmt.Fprintf(buf, "\n%s| dialogue-service-provider: null", gsmap.LogPrefix)
	case SrcPrvNoReason:
		fmt.Fprintf(buf, "\n%s| dialogue-service-provider: no-reason-given", gsmap.LogPrefix)
	case SrcPrvNoCommonDiagPortion:
		fmt.Fprintf(buf, "\n%s| dialogue-service-provider: no-common-dialogue-portion", gsmap.LogPrefix)
	default:
		fmt.Fprintf(buf, "\n%s| dialogue-service-???: unknown", gsmap.LogPrefix)
	}
	return buf.String()
}

func (d AARE) MarshalJSON() ([]byte, error) {
	j := map[string]any{}
	j["AARE"] = struct {
		Ver string `json:"protocol-version"`
		AARE
	}{
		Ver:  "version1",
		AARE: d}
	return json.Marshal(j)
}

/*
	Associate-result ::= INTEGER {
		accepted         (0),
		reject-permanent (1) }
*/
type Result byte

const (
	Accept          Result = 0x00
	RejectPermanent Result = 0x01
)

func (r Result) String() string {
	switch r {
	case Accept:
		return "accepted"
	case RejectPermanent:
		return "reject-permanent"
	}
	return ""
}

func (r *Result) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "accepted":
		*r = Accept
	case "reject-permanent":
		*r = RejectPermanent
	default:
		*r = 0
	}
	return
}

func (r Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

/*
	Associate-source-diagnostic ::= CHOICE {
		dialogue-service-user     [1] INTEGER {
			null                                   (0),
			no-reason-given                        (1),
			application-context-name-not-supported (2) },
		dialogue-service-provider [2] INTEGER {
			null                       (0),
			no-reason-given            (1),
			no-common-dialogue-portion (2) } }
*/
type ResultSrc byte

const (
	SrcUsrNull                ResultSrc = 0x00
	SrcUsrNoReason            ResultSrc = 0x01
	SrcUsrACNameNotSupported  ResultSrc = 0x02
	SrcPrvNull                ResultSrc = 0x10
	SrcPrvNoReason            ResultSrc = 0x11
	SrcPrvNoCommonDiagPortion ResultSrc = 0x12
)

func (r *ResultSrc) UnmarshalJSON(b []byte) (e error) {
	var m map[string]string
	e = json.Unmarshal(b, &m)
	for k, v := range m {
		switch k {
		case "dialogue-service-user":
			switch v {
			case "null":
				*r = SrcUsrNull
			case "no-reason-given":
				*r = SrcUsrNoReason
			case "application-context-name-not-supported":
				*r = SrcUsrACNameNotSupported
			}
		case "dialogue-service-provider":
			switch v {
			case "null":
				*r = SrcPrvNull
			case "no-reason-given":
				*r = SrcPrvNoReason
			case "no-common-dialogue-portion":
				*r = SrcPrvNoCommonDiagPortion
			}
		}
		break
	}
	return
}
func (r ResultSrc) MarshalJSON() ([]byte, error) {
	m := map[string]string{}
	switch r {
	case SrcUsrNull:
		m["dialogue-service-user"] = "null"
	case SrcUsrNoReason:
		m["dialogue-service-user"] = "no-reason-given"
	case SrcUsrACNameNotSupported:
		m["dialogue-service-user"] = "application-context-name-not-supported"
	case SrcPrvNull:
		m["dialogue-service-provider"] = "null"
	case SrcPrvNoReason:
		m["dialogue-service-provider"] = "no-reason-given"
	case SrcPrvNoCommonDiagPortion:
		m["dialogue-service-provider"] = "no-common-dialogue-portion"
	}
	return json.Marshal(m)
}

func (d *AARE) marshalDialogue() []byte {
	buf := new(bytes.Buffer)

	// protocol-version, context_specific(80) + primitive(00) + 0(00)
	// value = v1 (0x07 80)
	gsmap.WriteTLV(buf, 0x80, []byte{0x07, 0x80})

	// application-context-name, context_specific(80) + constructed(20) + 1(01)
	// OBJECT IDENTIFIER, universal(00) + primitive(00) + OID(06)
	gsmap.WriteTLV(buf, 0xa1, gsmap.WriteTLV(new(bytes.Buffer), 0x06, d.Context.Marshal()))

	// result, context_specific(80) + constructed(20) + 2(02)
	// Associate-result, universal(00) + primitive(00) + integer(02)
	gsmap.WriteTLV(buf, 0xa2, gsmap.WriteTLV(new(bytes.Buffer), 0x02, []byte{byte(d.Result)}))

	// result-source-diagnostic, context_specific(80) + constructed(20) + 3(03)
	// Associate-source-diagnostic, context_specific(80) + constructed(20) + 1/2(01/02)
	// dialogue-service-***, universal(00) + primitive(00) + integer(02)
	b := gsmap.WriteTLV(buf, 0xa3,
		gsmap.WriteTLV(new(bytes.Buffer), ((byte(d.ResultSrc)>>4)+1)|0xa0,
			gsmap.WriteTLV(new(bytes.Buffer), 0x02, []byte{byte(d.ResultSrc) & 0x0f})))

	// user-information, context_specific(80) + constructed(20) + 30(1e)

	// AARE-apdu, application(40) + constructed(20) + 1(01)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x61, b)
}

func (d *AARE) unmarshalDialogue(b []byte) error {
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

	// result, context_specific(80) + constructed(20) + 2(02)
	// Associate-result, universal(00) + primitive(00) + integer(02)
	if _, v, e := gsmap.ReadTLV(buf, 0xa2); e != nil {
		return e
	} else if _, v, e = gsmap.ReadTLV(bytes.NewBuffer(v), 0x02); e != nil {
		return e
	} else if len(v) != 1 || v[0] > 1 {
		e = gsmap.UnexpectedTLV("invalid parameter value")
	} else {
		d.Result = Result(v[0])
	}

	// result-source-diagnostic, context_specific(80) + constructed(20) + 3(03)
	// Associate-source-diagnostic, context_specific(80) + constructed(20) + 1/2(01/02)
	// dialogue-service-***, universal(00) + primitive(00) + integer(02)
	if _, v, e := gsmap.ReadTLV(buf, 0xa3); e != nil {
		return e
	} else if t, v, e := gsmap.ReadTLV(bytes.NewBuffer(v), 0x00); e != nil {
		return e
	} else if t != 0xa1 && t != 0xa2 {
		return gsmap.UnexpectedTag([]byte{0xa1, 0xa2}, t)
	} else {
		d.ResultSrc = ResultSrc((0x0f&t)-1) << 4

		if _, v, e = gsmap.ReadTLV(bytes.NewBuffer(v), 0x02); e != nil {
			return e
		} else if len(v) != 1 || v[0] > 2 {
			e = gsmap.UnexpectedTLV("invalid parameter value")
		} else {
			d.ResultSrc = d.ResultSrc | ResultSrc(v[0])
		}
	}

	// user-information, context_specific(80) + constructed(20) + 30(1e)

	return nil
}

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
		e = gsmap.UnexpectedTLV("invalid parameter value")
	} else {
		d.Source = Source(v[0])
	}

	return nil
}
