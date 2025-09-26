package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
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
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err AbsentSubscriberSM) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Diag != 0 && err.Diag <= TempUnavailable {
		fmt.Fprintf(buf, "\n%sabsentSubscriberDiagnosticSM:           %s",
			LogPrefix, err.Diag)
	}
	if err.AdditionalDiag != 0 && err.AdditionalDiag <= TempUnavailable {
		fmt.Fprintf(buf, "\n%sadditionalAbsentSubscriberDiagnosticSM: %s",
			LogPrefix, err.AdditionalDiag)
	}
	return buf.String()
}

func (err AbsentSubscriberSM) GetInvokeID() int8 { return err.InvokeID }
func (AbsentSubscriberSM) Code() byte            { return 6 }
func (AbsentSubscriberSM) Name() string          { return "AbsentSubscriberSM" }

func (AbsentSubscriberSM) NewFromJSON(v []byte, id int8) (Component, error) {
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
		WriteTLV(buf, 0x02, []byte{err.Diag.ToByte()})
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// additionalAbsentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 0(00)
	if err.AdditionalDiag != 0 {
		WriteTLV(buf, 0x80, []byte{err.AdditionalDiag.ToByte()})
	}

	if buf.Len() != 0 {
		// AbsentSubscriberSM-Param, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (AbsentSubscriberSM) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// AbsentSubscriberSM-Param, universal(00) + constructed(20) + sequence(10)
	err := AbsentSubscriberSM{InvokeID: id}
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

	// absentSubscriberDiagnosticSM, universal(00) + primitive(00) + integer(02)
	if t == 0x02 {
		if len(v) != 1 {
			return nil, UnexpectedTLV("invalid parameter value")
		}
		err.Diag.FromByte(v[0])

		if t, v, e = ReadTLV(buf, 0x00); e == io.EOF {
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

		if t, v, e = ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// additionalAbsentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if len(v) != 1 {
			return nil, UnexpectedTLV("invalid parameter value")
		}
		err.AdditionalDiag.FromByte(v[0])
		/*
			if t, v, e = ReadTLV(buf, nil); e == io.EOF {
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
