package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

/*
UnknownSubscriber error operation.
UnknownSubscriberParam must not be used in version <3.

	unknownSubscriber  ERROR ::= {
		PARAMETER
			UnknownSubscriberParam -- optional
		CODE local:1 }
	UnknownSubscriberParam ::= SEQUENCE {
		extensionContainer          ExtensionContainer          OPTIONAL,
		...,
		unknownSubscriberDiagnostic UnknownSubscriberDiagnostic OPTIONAL}
*/
type UnknownSubscriber struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	Diag UnknownSubscriberDiagnostic `json:"unknownSubscriberDiagnostic,omitempty"`
}

func init() {
	c := UnknownSubscriber{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err UnknownSubscriber) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Diag != 0 {
		fmt.Fprintf(buf, "\n%sunknownSubscriberDiagnostic: %s", LogPrefix, err.Diag)
	}
	return buf.String()
}

func (err UnknownSubscriber) GetInvokeID() int8 { return err.InvokeID }
func (UnknownSubscriber) Code() byte            { return 1 }
func (UnknownSubscriber) Name() string          { return "UnknownSubscriber" }

func (UnknownSubscriber) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		UnknownSubscriber
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.UnknownSubscriber, e
	}
	c := tmp.UnknownSubscriber
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err UnknownSubscriber) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// unknownSubscriberDiagnostic, universal(00) + primitive(00) + enum(0a)
	if tmp := err.Diag.marshal(); tmp != nil {
		WriteTLV(buf, 0x0a, tmp)
	}

	if buf.Len() != 0 {
		// UnknownSubscriberParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (UnknownSubscriber) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// UnknownSubscriberParam, universal(00) + constructed(20) + sequence(10)
	err := UnknownSubscriber{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := ReadTLV(buf, 0x00)
	if e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, v, e = ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// unknownSubscriberDiagnostic, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Diag.unmarshal(v); e != nil {
			return nil, e
		}
	}

	return err, nil
}

/*
UnknownSubscriberDiagnostic.
if unknown values are received in UnknownSubscriberDiagnostic they shall be discarded.

	UnknownSubscriberDiagnostic ::= ENUMERATED {
		imsiUnknown                  (0),
		gprs-eps-SubscriptionUnknown (1),
		...,
		npdbMismatch                 (2) }
*/
type UnknownSubscriberDiagnostic byte

const (
	_ UnknownSubscriberDiagnostic = iota
	ImsiUnknown
	GprsEpsSubscriptionUnknown
	NpdbMismatch
)

func (d UnknownSubscriberDiagnostic) String() string {
	switch d {
	case ImsiUnknown:
		return "imsiUnknown"
	case GprsEpsSubscriptionUnknown:
		return "gprs-eps-SubscriptionUnknown"
	case NpdbMismatch:
		return "npdbMismatch"
	}
	return ""
}

func (d UnknownSubscriberDiagnostic) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *UnknownSubscriberDiagnostic) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "imsiUnknown":
		*d = ImsiUnknown
	case "gprs-eps-SubscriptionUnknown":
		*d = GprsEpsSubscriptionUnknown
	case "npdbMismatch":
		*d = NpdbMismatch
	default:
		*d = 0
	}
	return
}

func (d *UnknownSubscriberDiagnostic) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*d = ImsiUnknown
	case 1:
		*d = GprsEpsSubscriptionUnknown
	case 2:
		*d = NpdbMismatch
	default:
		return UnexpectedEnumValue(b)
	}
	return nil
}

func (d UnknownSubscriberDiagnostic) marshal() []byte {
	switch d {
	case ImsiUnknown:
		return []byte{0x00}
	case GprsEpsSubscriptionUnknown:
		return []byte{0x01}
	case NpdbMismatch:
		return []byte{0x02}
	}
	return nil
}
