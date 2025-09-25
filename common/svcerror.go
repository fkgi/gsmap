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
AbsentSubscriberSM error operation.

	absentSubscriberSM  ERROR ::= {
		PARAMETER
			AbsentSubscriberSM-Param -- optional
		CODE local:6 }

	AbsentSubscriberSM-Param ::= SEQUENCE {
		absentSubscriberDiagnosticSM AbsentSubscriberDiagnosticSM OPTIONAL,
		-- AbsentSubscriberDiagnosticSM can be either for non-GPRS or for GPRS
		extensionContainer           ExtensionContainer           OPTIONAL,
		...,
		additionalAbsentSubscriberDiagnosticSM [0] AbsentSubscriberDiagnosticSM OPTIONAL,
		-- if received, additionalAbsentSubscriberDiagnosticSM
		-- is for GPRS and absentSubscriberDiagnosticSM is for non-GPRS
		-- ^^^^^^ R99 ^^^^^^
		imsi [1] IMSI OPTIONAL,
		-- when sent from HLR to IP-SM-GW, IMSI shall be present if UNRI is not set
		-- to indicate that the absent condition is met for CS and PS but not for IMS.
		requestedRetransmissionTime [2] Time OPTIONAL,
		userIdentifierAlert         [3] IMSI OPTIONAL }
*/
type AbsentSubscriberSM struct {
	InvokeID int8 `json:"id"`

	Diag AbsentDiag `json:"absentSubscriberDiagnosticSM,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	AdditionalDiag AbsentDiag `json:"additionalAbsentSubscriberDiagnosticSM,omitempty"`
}

func init() {
	c := AbsentSubscriberSM{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err AbsentSubscriberSM) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Diag != 0 && err.Diag <= TempUnavailable {
		fmt.Fprintf(buf, "\n%sabsentSubscriberDiagnosticSM:           %s",
			gsmap.LogPrefix, err.Diag)
	}
	if err.AdditionalDiag != 0 && err.AdditionalDiag <= TempUnavailable {
		fmt.Fprintf(buf, "\n%sadditionalAbsentSubscriberDiagnosticSM: %s",
			gsmap.LogPrefix, err.AdditionalDiag)
	}
	return buf.String()
}

func (err AbsentSubscriberSM) GetInvokeID() int8 { return err.InvokeID }
func (AbsentSubscriberSM) Code() byte            { return 6 }
func (AbsentSubscriberSM) Name() string          { return "AbsentSubscriberSM" }

func (AbsentSubscriberSM) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		AbsentSubscriberSM
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.AbsentSubscriberSM, e
	}
	c := tmp.AbsentSubscriberSM
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err AbsentSubscriberSM) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// absentSubscriberDiagnosticSM, universal(00) + primitive(00) + integer(02)
	if err.Diag != 0 {
		gsmap.WriteTLV(buf, 0x02, []byte{err.Diag.ToByte()})
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// additionalAbsentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 0(00)
	if err.AdditionalDiag != 0 {
		gsmap.WriteTLV(buf, 0x80, []byte{err.AdditionalDiag.ToByte()})
	}

	if buf.Len() != 0 {
		// AbsentSubscriberSM-Param, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (AbsentSubscriberSM) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// AbsentSubscriberSM-Param, universal(00) + constructed(20) + sequence(10)
	err := AbsentSubscriberSM{InvokeID: id}
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

	// absentSubscriberDiagnosticSM, universal(00) + primitive(00) + integer(02)
	if t == 0x02 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedTLV("invalid parameter value")
		}
		err.Diag.FromByte(v[0])

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
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

	// additionalAbsentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedTLV("invalid parameter value")
		}
		err.AdditionalDiag.FromByte(v[0])
		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return err, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return err, nil
}

/*
AbsentDiag

	AbsentSubscriberDiagnosticSM ::= INTEGER (0..255)
	-- AbsentSubscriberDiagnosticSM values are defined in 3GPP TS 23.040
*/
type AbsentDiag byte

const (
	_ AbsentDiag = iota
	NoPagingRespMSC
	IMSIDetached
	RoamingRestrict
	DeregisteredNonGPRS
	PurgedNonGPRS
	NoPagingRespSGSN
	GPRSDetached
	DeregisteredGPRS
	PurgedGPRS
	UnidentifiedSubsMSC
	UnidentifiedSubsSGSN
	DeregisteredIMS
	NoRespIPSMGW
	TempUnavailable
)

func (a AbsentDiag) String() string {
	switch a {
	case NoPagingRespMSC:
		return "no paging response via the MSC"
	case IMSIDetached:
		return "IMSI detached"
	case RoamingRestrict:
		return "roaming restriction"
	case DeregisteredNonGPRS:
		return "deregistered in the HLR for non GPRS"
	case PurgedNonGPRS:
		return "MS purged for non GPRS"
	case NoPagingRespSGSN:
		return "no paging response via the SGSN"
	case GPRSDetached:
		return "GPRS detached"
	case DeregisteredGPRS:
		return "deregistered in the HLR for GPRS"
	case PurgedGPRS:
		return "MS purged for GPRS"
	case UnidentifiedSubsMSC:
		return "Unidentified subscriber via the MSC"
	case UnidentifiedSubsSGSN:
		return "Unidentified subscriber via the SGSN"
	case DeregisteredIMS:
		return "deregistered in the HSS/HLR for IMS"
	case NoRespIPSMGW:
		return "no response via the IP-SM-GW"
	case TempUnavailable:
		return "the MS is temporarily unavailable"
	}
	return ""
}

func (a AbsentDiag) ToByte() byte {
	switch a {
	case NoPagingRespMSC:
		return 0
	case IMSIDetached:
		return 1
	case RoamingRestrict:
		return 2
	case DeregisteredNonGPRS:
		return 3
	case PurgedNonGPRS:
		return 4
	case NoPagingRespSGSN:
		return 5
	case GPRSDetached:
		return 6
	case DeregisteredGPRS:
		return 7
	case PurgedGPRS:
		return 8
	case UnidentifiedSubsMSC:
		return 9
	case UnidentifiedSubsSGSN:
		return 10
	case DeregisteredIMS:
		return 11
	case NoRespIPSMGW:
		return 12
	case TempUnavailable:
		return 13
	}
	return byte(a - 1)
}

func (a *AbsentDiag) FromByte(b byte) {
	switch b {
	case 0:
		*a = NoPagingRespMSC
	case 1:
		*a = IMSIDetached
	case 2:
		*a = RoamingRestrict
	case 3:
		*a = DeregisteredNonGPRS
	case 4:
		*a = PurgedNonGPRS
	case 5:
		*a = NoPagingRespSGSN
	case 6:
		*a = GPRSDetached
	case 7:
		*a = DeregisteredGPRS
	case 8:
		*a = PurgedGPRS
	case 9:
		*a = UnidentifiedSubsMSC
	case 10:
		*a = UnidentifiedSubsSGSN
	case 11:
		*a = DeregisteredIMS
	case 12:
		*a = NoRespIPSMGW
	case 13:
		*a = TempUnavailable
	}
	*a = AbsentDiag(b + 1)
}

func (d AbsentDiag) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.ToByte())
}

func (d *AbsentDiag) UnmarshalJSON(b []byte) (e error) {
	var a byte
	e = json.Unmarshal(b, &a)
	d.FromByte(a)
	return
}

/*
TeleserviceNotProvisioned error operation.
TeleservNotProvParam must not be used in version <3.

	teleserviceNotProvisioned  ERROR ::= {
		PARAMETER
			TeleservNotProvParam -- optional
		CODE	local:11 }

	TeleservNotProvParam ::= SEQUENCE {
		extensionContainer	ExtensionContainer	OPTIONAL,
		...}
*/
type TeleserviceNotProvisioned struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := TeleserviceNotProvisioned{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err TeleserviceNotProvisioned) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err TeleserviceNotProvisioned) GetInvokeID() int8 { return err.InvokeID }
func (TeleserviceNotProvisioned) Code() byte            { return 11 }
func (TeleserviceNotProvisioned) Name() string          { return "TeleserviceNotProvisioned" }

func (TeleserviceNotProvisioned) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		TeleserviceNotProvisioned
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.TeleserviceNotProvisioned, e
	}
	c := tmp.TeleserviceNotProvisioned
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (TeleserviceNotProvisioned) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// TeleservNotProvParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (TeleserviceNotProvisioned) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// TeleservNotProvParam, universal(00) + constructed(20) + sequence(10)
	err := TeleserviceNotProvisioned{InvokeID: id}
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
CallBarred error operation.
callBarringCause must not be used in version 3 and higher.
extensibleCallBarredParam must not be used in version <3.

	callBarred  ERROR ::= {
		PARAMETER
			CallBarredParam -- optional
		CODE	local:13 }

	CallBarredParam ::= CHOICE {
		callBarringCause          CallBarringCause,
		extensibleCallBarredParam ExtensibleCallBarredParam }
	ExtensibleCallBarredParam ::= SEQUENCE {
		callBarringCause	CallBarringCause	OPTIONAL,
		extensionContainer	ExtensionContainer	OPTIONAL,
		... ,
		unauthorisedMessageOriginator	[1] NULL	OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		anonymousCallRejection	[2] NULL	OPTIONAL }

unauthorisedMessageOriginator and anonymousCallRejection shall be mutually exclusive.
*/
type CallBarred struct {
	InvokeID int8 `json:"id"`

	NotExtensible bool             `json:"notExtensible,omitempty"`
	Cause         CallBarringCause `json:"callBarringCause,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	UnauthorisedMessageOriginator bool `json:"unauthorisedMessageOriginator,omitempty"`
	// AnonymousCallRejection bool
}

func init() {
	c := CallBarred{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err CallBarred) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Cause != 0 {
		fmt.Fprintf(buf, "\n%scallBarringCause: %s", gsmap.LogPrefix, err.Cause)
	}
	if err.UnauthorisedMessageOriginator {
		fmt.Fprintf(buf, "\n%sunauthorisedMessageOriginator:", gsmap.LogPrefix)
	}
	return buf.String()
}

func (err CallBarred) GetInvokeID() int8 { return err.InvokeID }
func (CallBarred) Code() byte            { return 13 }
func (CallBarred) Name() string          { return "CallBarred" }

func (CallBarred) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		CallBarred
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.CallBarred, e
	}
	c := tmp.CallBarred
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err CallBarred) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// callBarringCause, universal(00) + primitive(00) + enum(0a)
	if tmp := err.Cause.marshal(); tmp != nil {
		gsmap.WriteTLV(buf, 0x0a, tmp)
	}

	if err.NotExtensible {
		if buf.Len() != 0 {
			return buf.Bytes()
		} else {
			return nil
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// unauthorisedMessageOriginator, context_specific(80) + primitive(00) + 1(01)
	if err.UnauthorisedMessageOriginator {
		gsmap.WriteTLV(buf, 0x81, nil)
	}

	// anonymousCallRejection, context_specific(80) + primitive(00) + 2(02)
	// if err.AnonymousCallRejection {
	//	gsmap.WriteTLV(buf, []byte{0x82}, nil)
	// }

	if buf.Len() != 0 {
		// ExtensibleCallBarredParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (CallBarred) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	err := CallBarred{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	}

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e != nil {
		return nil, e
	}
	// callBarringCause, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Cause.unmarshal(v); e != nil {
			return nil, e
		}
		err.NotExtensible = true
		return err, nil
	}

	// ExtensibleCallBarredParam, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		buf = bytes.NewBuffer(v)
	} else {
		return nil, gsmap.UnexpectedTag([]byte{0x30}, t)
	}

	// OPTIONAL TLV
	t, v, e = gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	}

	// callBarringCause, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Cause.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, _, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// unauthorisedMessageOriginator, context_specific(80) + primitive(00) + 1(01)
	if t == 0x81 {
		err.UnauthorisedMessageOriginator = true
		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return err, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	// anonymousCallRejection, context_specific(80) + primitive(00) + 2(02)
	// if t[0] == 0x82 {
	//	err.AnonymousCallRejection = true
	//	if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
	//		return err, nil
	//	} else if e != nil {
	//		return nil, e
	//	}
	// }

	return err, nil
}

/*
CallBarringCause enum.

	CallBarringCause ::= ENUMERATED {
		barringServiceActive (0),
		operatorBarring      (1)}
*/
type CallBarringCause byte

const (
	_ CallBarringCause = iota
	BarringServiceActive
	OperatorBarring
)

func (c CallBarringCause) String() string {
	switch c {
	case BarringServiceActive:
		return "barringServiceActive"
	case OperatorBarring:
		return "operatorBarring"
	}
	return ""
}

func (c CallBarringCause) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *CallBarringCause) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "barringServiceActive":
		*c = BarringServiceActive
	case "operatorBarring":
		*c = OperatorBarring
	default:
		*c = 0
	}
	return
}

func (c *CallBarringCause) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = BarringServiceActive
	case 1:
		*c = OperatorBarring
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (c CallBarringCause) marshal() []byte {
	switch c {
	case BarringServiceActive:
		return []byte{0x00}
	case OperatorBarring:
		return []byte{0x01}
	}
	return nil
}

/*
FacilityNotSupported error operation.
FacilityNotSupParam must not be used in version <3.

	facilityNotSupported  ERROR ::= {
		PARAMETER
			FacilityNotSupParam -- optional
		CODE local:21 }

	FacilityNotSupParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		shapeOfLocationEstimateNotSupported          [0] NULL OPTIONAL,
		neededLcsCapabilityNotSupportedInServingNode [1] NULL OPTIONAL }
*/
type FacilityNotSupported struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := FacilityNotSupported{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err FacilityNotSupported) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err FacilityNotSupported) GetInvokeID() int8 { return err.InvokeID }
func (FacilityNotSupported) Code() byte            { return 21 }
func (FacilityNotSupported) Name() string          { return "FacilityNotSupported" }

func (FacilityNotSupported) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		FacilityNotSupported
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.FacilityNotSupported, e
	}
	c := tmp.FacilityNotSupported
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (FacilityNotSupported) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// FacilityNotSupParam, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (FacilityNotSupported) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// FacilityNotSupParam, universal(00) + constructed(20) + sequence(10)
	err := FacilityNotSupported{InvokeID: id}
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
		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return err, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}
	return err, nil
}
