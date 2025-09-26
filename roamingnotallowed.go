package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

/*
RoamingNotAllowed error operation.

	roamingNotAllowed  ERROR ::= {
		PARAMETER
			RoamingNotAllowedParam
		CODE	local:8 }
	RoamingNotAllowedParam ::= SEQUENCE {
		roamingNotAllowedCause RoamingNotAllowedCause,
		extensionContainer     ExtensionContainer     OPTIONAL,
		...,
		additionalRoamingNotAllowedCause [0] AdditionalRoamingNotAllowedCause OPTIONAL }
	--	if the additionalRoamingNotallowedCause is received by the MSC/VLR or SGSN then the
	--	roamingNotAllowedCause shall be discarded.
*/
type RoamingNotAllowed struct {
	InvokeID int8 `json:"id"`

	Cause RoamingNotAllowedCause `json:"roamingNotAllowedCause"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	AdditionalCause AdditionalRoamingNotAllowedCause `json:"additionalRoamingNotAllowedCause,omitempty"`
}

func init() {
	c := RoamingNotAllowed{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err RoamingNotAllowed) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Cause != 0 {
		fmt.Fprintf(buf, "\n%sroamingNotAllowedCause: %s", LogPrefix, err.Cause)
	}
	if err.AdditionalCause != 0 {
		fmt.Fprintf(buf, "\n%sadditionalRoamingNotAllowedCause: %s", LogPrefix, err.AdditionalCause)
	}
	return buf.String()
}

func (err RoamingNotAllowed) GetInvokeID() int8 { return err.InvokeID }
func (RoamingNotAllowed) Code() byte            { return 8 }
func (RoamingNotAllowed) Name() string          { return "RoamingNotAllowed" }

func (RoamingNotAllowed) NewFromJSON(v []byte, id int8) (Component, error) {
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

func (err RoamingNotAllowed) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// roamingNotAllowedCause, universal(00) + primitive(00) + enum(0a)
	if tmp := err.Cause.marshal(); tmp != nil {
		WriteTLV(buf, 0x0a, tmp)
	} else {
		WriteTLV(buf, 0x0a, []byte{0x00})
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// additionalRoamingNotAllowedCause,
	// context_specific(80) + primitive(00) + 0(00)
	if tmp := err.AdditionalCause.marshal(); tmp != nil {
		WriteTLV(buf, 0x80, tmp)
	}

	if buf.Len() != 0 {
		// RoamingNotAllowedParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (RoamingNotAllowed) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// RoamingNotAllowedParam, universal(00) + constructed(20) + sequence(10)
	err := RoamingNotAllowed{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// roamingNotAllowedCause, universal(00) + primitive(00) + enum(0a)
	if _, v, e := ReadTLV(buf, 0x0a); e != nil {
		return nil, e
	} else if e = err.Cause.unmarshal(v); e != nil {
		return nil, e
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

	// additionalRoamingNotAllowedCause,
	// context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if e = err.AdditionalCause.unmarshal(v); e != nil {
			return nil, e
		}
	}

	return err, nil
}

/*
RoamingNotAllowedCause

	RoamingNotAllowedCause ::= ENUMERATED {
		plmnRoamingNotAllowed     (0),
		operatorDeterminedBarring (3)}
*/
type RoamingNotAllowedCause byte

const (
	_ RoamingNotAllowedCause = iota
	PlmnRoamingNotAllowed
	OperatorDeterminedBarring
)

func (c RoamingNotAllowedCause) String() string {
	switch c {
	case PlmnRoamingNotAllowed:
		return "plmnRoamingNotAllowed"
	case OperatorDeterminedBarring:
		return "operatorDeterminedBarring"
	}
	return ""
}

func (c RoamingNotAllowedCause) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *RoamingNotAllowedCause) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "plmnRoamingNotAllowed":
		*c = PlmnRoamingNotAllowed
	case "operatorDeterminedBarring":
		*c = OperatorDeterminedBarring
	default:
		*c = 0
	}
	return
}

func (c *RoamingNotAllowedCause) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = PlmnRoamingNotAllowed
	case 3:
		*c = OperatorDeterminedBarring
	default:
		return UnexpectedEnumValue(b)
	}
	return nil
}

func (c RoamingNotAllowedCause) marshal() []byte {
	switch c {
	case PlmnRoamingNotAllowed:
		return []byte{0x00}
	case OperatorDeterminedBarring:
		return []byte{0x03}
	}
	return nil
}

/*
AdditionalRoamingNotAllowedCause

	AdditionalRoamingNotAllowedCause ::= ENUMERATED {
		supportedRAT-TypesNotAllowed (0),
		...}
*/
type AdditionalRoamingNotAllowedCause byte

const (
	_ AdditionalRoamingNotAllowedCause = iota
	SupportedRatTypesNotAllowed
)

func (c AdditionalRoamingNotAllowedCause) String() string {
	switch c {
	case SupportedRatTypesNotAllowed:
		return "supportedRAT-TypesNotAllowed"
	}
	return ""
}

func (c AdditionalRoamingNotAllowedCause) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *AdditionalRoamingNotAllowedCause) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "supportedRAT-TypesNotAllowed":
		*c = SupportedRatTypesNotAllowed
	default:
		*c = 0
	}
	return
}

func (c *AdditionalRoamingNotAllowedCause) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = SupportedRatTypesNotAllowed
	default:
		return UnexpectedEnumValue(b)
	}
	return nil
}

func (c AdditionalRoamingNotAllowedCause) marshal() []byte {
	switch c {
	case SupportedRatTypesNotAllowed:
		return []byte{0x00}
	}
	return nil
}
