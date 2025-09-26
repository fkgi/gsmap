package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

/*
AbsentSubscriber error operation.
absentSubscriberParam must not be used in version <3

	absentSubscriber  ERROR ::= {
		PARAMETER
			AbsentSubscriberParam -- optional
		CODE	local:27 }
	AbsentSubscriberParam ::= SEQUENCE {
		extensionContainer	ExtensionContainer	OPTIONAL,
		...,
		absentSubscriberReason	[0] AbsentSubscriberReason	OPTIONAL}
*/
type AbsentSubscriber struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	Reason absentSubscriberReason `json:"absentSubscriberReason,omitempty"`
}

func init() {
	c := AbsentSubscriber{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err AbsentSubscriber) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Reason != 0 {
		fmt.Fprintf(buf, "\n%sabsentSubscriberReason: %s", LogPrefix, err.Reason)
	}
	return buf.String()
}

func (err AbsentSubscriber) GetInvokeID() int8 { return err.InvokeID }
func (AbsentSubscriber) Code() byte            { return 27 }
func (AbsentSubscriber) Name() string          { return "AbsentSubscriber" }

func (AbsentSubscriber) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		AbsentSubscriber
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.AbsentSubscriber, e
	}
	c := tmp.AbsentSubscriber
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err AbsentSubscriber) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// absentSubscriberReason, context_specific(80) + primitive(00) + 0(00)
	if tmp := err.Reason.marshal(); tmp != nil {
		WriteTLV(buf, 0x80, tmp)
	}

	if buf.Len() != 0 {
		// AbsentSubscriberParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (AbsentSubscriber) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// AbsentSubscriberParam, universal(00) + constructed(20) + sequence(10)
	err := AbsentSubscriber{InvokeID: id}
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

	// absentSubscriberReason, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if e = err.Reason.unmarshal(v); e != nil {
			return nil, e
		}
	}

	return err, nil
}

/*
absentSubscriberReason enum.

	AbsentSubscriberReason ::= ENUMERATED {
		imsiDetach     (0),
		restrictedArea (1),
		noPageResponse (2),
		... ,
		purgedMS       (3),
		mtRoamingRetry (4),
		busySubscriber (5)}
*/
type absentSubscriberReason byte

const (
	_ absentSubscriberReason = iota
	ImsiDetach
	RestrictedArea
	NoPageResponse
	PurgedMS
	MtRoamingRetry
	BusySubscriber
)

func (d absentSubscriberReason) String() string {
	switch d {
	case ImsiDetach:
		return "imsiDetach"
	case RestrictedArea:
		return "restrictedArea"
	case NoPageResponse:
		return "noPageResponse"
	case PurgedMS:
		return "purgedMS"
	case MtRoamingRetry:
		return "mtRoamingRetry"
	case BusySubscriber:
		return "busySubscriber"
	}
	return ""
}

func (d absentSubscriberReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *absentSubscriberReason) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "imsiDetach":
		*d = ImsiDetach
	case "restrictedArea":
		*d = RestrictedArea
	case "noPageResponse":
		*d = NoPageResponse
	case "purgedMS":
		*d = PurgedMS
	case "mtRoamingRetry":
		*d = MtRoamingRetry
	case "busySubscriber":
		*d = BusySubscriber
	default:
		*d = 0
	}
	return
}

func (d *absentSubscriberReason) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*d = ImsiDetach
	case 1:
		*d = RestrictedArea
	case 2:
		*d = NoPageResponse
	case 3:
		*d = PurgedMS
	case 4:
		*d = MtRoamingRetry
	case 5:
		*d = BusySubscriber
	default:
		return UnexpectedEnumValue(b)
	}
	return nil
}

func (d absentSubscriberReason) marshal() []byte {
	switch d {
	case ImsiDetach:
		return []byte{0x00}
	case RestrictedArea:
		return []byte{0x01}
	case NoPageResponse:
		return []byte{0x02}
	case PurgedMS:
		return []byte{0x03}
	case MtRoamingRetry:
		return []byte{0x04}
	case BusySubscriber:
		return []byte{0x05}
	}
	return nil
}
