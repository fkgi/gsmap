package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
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
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err UnknownSubscriber) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Diag != 0 {
		fmt.Fprintf(buf, "\n%sunknownSubscriberDiagnostic: %s", gsmap.LogPrefix, err.Diag)
	}
	return buf.String()
}

func (err UnknownSubscriber) GetInvokeID() int8 { return err.InvokeID }
func (UnknownSubscriber) Code() byte            { return 1 }
func (UnknownSubscriber) Name() string          { return "UnknownSubscriber" }

func (UnknownSubscriber) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
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
		gsmap.WriteTLV(buf, 0x0a, tmp)
	}

	if buf.Len() != 0 {
		// UnknownSubscriberParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (UnknownSubscriber) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// UnknownSubscriberParam, universal(00) + constructed(20) + sequence(10)
	err := UnknownSubscriber{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
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

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
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
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*d = ImsiUnknown
	case 1:
		*d = GprsEpsSubscriptionUnknown
	case 2:
		*d = NpdbMismatch
	default:
		return gsmap.UnexpectedEnumValue(b)
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

/*
UnidentifiedSubscriber error operation.
UnidentifiedSubParam must not be used in version <3.

	unidentifiedSubscriber  ERROR ::= {
		PARAMETER
			UnidentifiedSubParam -- optional
		CODE local:5 }
	UnidentifiedSubParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		... }
*/
type UnidentifiedSubscriber struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := UnidentifiedSubscriber{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err UnidentifiedSubscriber) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err UnidentifiedSubscriber) GetInvokeID() int8 { return err.InvokeID }
func (UnidentifiedSubscriber) Code() byte            { return 5 }
func (UnidentifiedSubscriber) Name() string          { return "UnidentifiedSubscriber" }

func (UnidentifiedSubscriber) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		UnidentifiedSubscriber
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.UnidentifiedSubscriber, e
	}
	c := tmp.UnidentifiedSubscriber
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (UnidentifiedSubscriber) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// UnidentifiedSubParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (UnidentifiedSubscriber) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// UnidentifiedSubParam, universal(00) + constructed(20) + sequence(10)
	err := UnidentifiedSubscriber{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t, v, e := gsmap.ReadTLV(buf, 0x00); e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	} else if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}
	return err, nil
}

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
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err RoamingNotAllowed) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Cause != 0 {
		fmt.Fprintf(buf, "\n%sroamingNotAllowedCause: %s", gsmap.LogPrefix, err.Cause)
	}
	if err.AdditionalCause != 0 {
		fmt.Fprintf(buf, "\n%sadditionalRoamingNotAllowedCause: %s", gsmap.LogPrefix, err.AdditionalCause)
	}
	return buf.String()
}

func (err RoamingNotAllowed) GetInvokeID() int8 { return err.InvokeID }
func (RoamingNotAllowed) Code() byte            { return 8 }
func (RoamingNotAllowed) Name() string          { return "RoamingNotAllowed" }

func (RoamingNotAllowed) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
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
		gsmap.WriteTLV(buf, 0x0a, tmp)
	} else {
		gsmap.WriteTLV(buf, 0x0a, []byte{0x00})
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// additionalRoamingNotAllowedCause,
	// context_specific(80) + primitive(00) + 0(00)
	if tmp := err.AdditionalCause.marshal(); tmp != nil {
		gsmap.WriteTLV(buf, 0x80, tmp)
	}

	if buf.Len() != 0 {
		// RoamingNotAllowedParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (RoamingNotAllowed) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// RoamingNotAllowedParam, universal(00) + constructed(20) + sequence(10)
	err := RoamingNotAllowed{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// roamingNotAllowedCause, universal(00) + primitive(00) + enum(0a)
	if _, v, e := gsmap.ReadTLV(buf, 0x0a); e != nil {
		return nil, e
	} else if e = err.Cause.unmarshal(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
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

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
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
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = PlmnRoamingNotAllowed
	case 3:
		*c = OperatorDeterminedBarring
	default:
		return gsmap.UnexpectedEnumValue(b)
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
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = SupportedRatTypesNotAllowed
	default:
		return gsmap.UnexpectedEnumValue(b)
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

/*
IllegalSubscriber error operation.
IllegalSubscriberParam must not be used in version <3.

	illegalSubscriber  ERROR ::= {
		PARAMETER
			IllegalSubscriberParam -- optional
		CODE local:9 }
	IllegalSubscriberParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		... }
*/
type IllegalSubscriber struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := IllegalSubscriber{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err IllegalSubscriber) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err IllegalSubscriber) GetInvokeID() int8 { return err.InvokeID }
func (IllegalSubscriber) Code() byte            { return 9 }
func (IllegalSubscriber) Name() string          { return "IllegalSubscriber" }

func (IllegalSubscriber) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		IllegalSubscriber
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.IllegalSubscriber, e
	}
	c := tmp.IllegalSubscriber
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (IllegalSubscriber) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// IllegalSubscriberParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (IllegalSubscriber) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// IllegalSubscriberParam, universal(00) + constructed(20) + sequence(10)
	err := IllegalSubscriber{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t, v, e := gsmap.ReadTLV(buf, 0x00); e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	} else if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}
	return err, nil
}

/*
IllegalEquipment error operation.
IllegalEquipment must not be used in version 1.
IllegalEquipmentParam must not be used in version <3

	illegalEquipment  ERROR ::= {
		PARAMETER
			IllegalEquipmentParam -- optional
		CODE local:12 }
	IllegalEquipmentParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		... }
*/
type IllegalEquipment struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := IllegalEquipment{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err IllegalEquipment) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err IllegalEquipment) GetInvokeID() int8 { return err.InvokeID }
func (IllegalEquipment) Code() byte            { return 12 }
func (IllegalEquipment) Name() string          { return "IllegalEquipment" }

func (IllegalEquipment) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		IllegalEquipment
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.IllegalEquipment, e
	}
	c := tmp.IllegalEquipment
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (IllegalEquipment) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// IllegalEquipmentParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (IllegalEquipment) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// IllegalEquipmentParam, universal(00) + constructed(20) + sequence(10)
	err := IllegalEquipment{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t, v, e := gsmap.ReadTLV(buf, 0x00); e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	} else if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}
	return err, nil
}

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
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err AbsentSubscriber) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Reason != 0 {
		fmt.Fprintf(buf, "\n%sabsentSubscriberReason: %s", gsmap.LogPrefix, err.Reason)
	}
	return buf.String()
}

func (err AbsentSubscriber) GetInvokeID() int8 { return err.InvokeID }
func (AbsentSubscriber) Code() byte            { return 27 }
func (AbsentSubscriber) Name() string          { return "AbsentSubscriber" }

func (AbsentSubscriber) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
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
		gsmap.WriteTLV(buf, 0x80, tmp)
	}

	if buf.Len() != 0 {
		// AbsentSubscriberParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (AbsentSubscriber) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// AbsentSubscriberParam, universal(00) + constructed(20) + sequence(10)
	err := AbsentSubscriber{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
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

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
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
		return gsmap.UnexpectedEnumValue(b)
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
		return gsmap.UnexpectedEnumValue(b)
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
