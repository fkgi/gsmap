package ifc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
reportSM-DeliveryStatus  OPERATION ::= {	--Timer s
	ARGUMENT
		ReportSM-DeliveryStatusArg
	RESULT
		ReportSM-DeliveryStatusRes	-- optional
	ERRORS {
		dataMissing           |
		unexpectedDataValue   |
		unknownSubscriber     |
		messageWaitingListFull}
	CODE	local:47 }

dataMissing must not be used in version 1.
result must be absent in version 1.
*/

func init() {
	a := ReportSmDeliveryStatusArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := ReportSmDeliveryStatusRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
ReportSmDeliveryStatusArg operation arg.

	ReportSM-DeliveryStatusArg ::= SEQUENCE {
		msisdn                           ISDN-AddressString,
		serviceCentreAddress             AddressString,
		sm-DeliveryOutcome               SM-DeliveryOutcome,
		absentSubscriberDiagnosticSM [0] AbsentSubscriberDiagnosticSM OPTIONAL,
		extensionContainer           [1] ExtensionContainer           OPTIONAL,
		...,
		gprsSupportIndicator                   [2]  NULL                         OPTIONAL,
		deliveryOutcomeIndicator               [3]  NULL                         OPTIONAL,
		additionalSM-DeliveryOutcome           [4]  SM-DeliveryOutcome           OPTIONAL,
		additionalAbsentSubscriberDiagnosticSM [5]  AbsentSubscriberDiagnosticSM OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		ip-sm-gw-Indicator                     [6]  NULL                         OPTIONAL,
		ip-sm-gw-sm-deliveryOutcome	           [7]  SM-DeliveryOutcome           OPTIONAL,
		ip-sm-gw-absentSubscriberDiagnosticSM  [8]  AbsentSubscriberDiagnosticSM OPTIONAL,
		imsi                                   [9]  IMSI                         OPTIONAL,
		singleAttemptDelivery                  [10] NULL                         OPTIONAL,
		correlationID                          [11] CorrelationID                OPTIONAL,
		smsf-3gpp-deliveryOutcomeIndicator     [12] NULL                         OPTIONAL,
		smsf-3gpp-deliveryOutcome              [13] SM-DeliveryOutcome           OPTIONAL,
		smsf-3gpp-absentSubscriberDiagSM       [14] AbsentSubscriberDiagnosticSM OPTIONAL,
		smsf-non-3gpp-deliveryOutcomeIndicator [15] NULL                         OPTIONAL,
		smsf-non-3gpp-deliveryOutcome          [16] SM-DeliveryOutcome           OPTIONAL,
		smsf-non-3gpp-absentSubscriberDiagSM   [17] AbsentSubscriberDiagnosticSM OPTIONAL }

gprsSupportIndicator is set only if the SMS-GMSC supports handling of two delivery outcomes.
DeliveryOutcomeIndicator is set when the SM-DeliveryOutcome is for GPRS.
If received, additionalSM-DeliveryOutcome is for GPRS.
If DeliveryOutcomeIndicator is set, then AdditionalSM-DeliveryOutcome shall be absent.
If received additionalAbsentSubscriberDiagnosticSM is for GPRS.
If DeliveryOutcomeIndicator is set, then AdditionalAbsentSubscriberDiagnosticSM shall be absent.
*/
type ReportSmDeliveryStatusArg struct {
	InvokeID int8 `json:"id"`

	MSISDN     gsmap.AddressString `json:"msisdn"`
	CenterAddr gsmap.AddressString `json:"serviceCentreAddress"`
	Outcome    Outcome             `json:"sm-DeliveryOutcome"`
	AbsentDiag gsmap.AbsentDiag    `json:"absentSubscriberDiagnosticSM,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	SupportGPRS       bool             `json:"gprsSupportIndicator,omitempty"`
	OutcomeIsGPRS     bool             `json:"deliveryOutcomeIndicator,omitempty"`
	AdditionalOutcome Outcome          `json:"additionalSM-DeliveryOutcome,omitempty"`
	AdditionalDiag    gsmap.AbsentDiag `json:"additionalAbsentSubscriberDiagnosticSM,omitempty"`
}

func (rsds ReportSmDeliveryStatusArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", rsds.Name(), rsds.InvokeID)
	fmt.Fprintf(buf, "\n%smsisdn:              %s", gsmap.LogPrefix, rsds.MSISDN)
	fmt.Fprintf(buf, "\n%sserviceCentreAddress:%s", gsmap.LogPrefix, rsds.CenterAddr)
	fmt.Fprintf(buf, "\n%ssm-DeliveryOutcome:  %s", gsmap.LogPrefix, rsds.Outcome)
	if rsds.AbsentDiag != 0 {
		fmt.Fprintf(buf, "\n%sabsentSubscriberDiagnosticSM:%s",
			gsmap.LogPrefix, rsds.AbsentDiag)
	}
	// Extension
	if rsds.SupportGPRS {
		fmt.Fprintf(buf, "\n%sgprsSupportIndicator:", gsmap.LogPrefix)
	}
	if rsds.OutcomeIsGPRS {
		fmt.Fprintf(buf, "\n%sdeliveryOutcomeIndicator:", gsmap.LogPrefix)
	}
	if rsds.AdditionalOutcome != 0 {
		fmt.Fprintf(buf, "\n%sadditionalSM-DeliveryOutcome:  %s",
			gsmap.LogPrefix, rsds.AdditionalOutcome)
	}
	if rsds.AdditionalDiag != 0 {
		fmt.Fprintf(buf, "\n%sadditionalAbsentSubscriberDiagnosticSM:%s",
			gsmap.LogPrefix, rsds.AdditionalDiag)
	}
	return buf.String()
}

func (rsds ReportSmDeliveryStatusArg) GetInvokeID() int8           { return rsds.InvokeID }
func (ReportSmDeliveryStatusArg) GetLinkedID() *int8               { return nil }
func (ReportSmDeliveryStatusArg) Code() byte                       { return 47 }
func (ReportSmDeliveryStatusArg) Name() string                     { return "ReportSM-DeliveryStatus-Arg" }
func (ReportSmDeliveryStatusArg) DefaultContext() gsmap.AppContext { return ShortMsgGateway1 }

func (ReportSmDeliveryStatusArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		ReportSmDeliveryStatusArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.ReportSmDeliveryStatusArg, e
	}
	c := tmp.ReportSmDeliveryStatusArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (rsds ReportSmDeliveryStatusArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, rsds.MSISDN.Bytes())

	// serviceCentreAddress, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, rsds.CenterAddr.Bytes())

	// sm-DeliveryOutcome, universal(00) + primitive(00) + enum(0a)
	if tmp := rsds.Outcome.marshal(); tmp != nil {
		gsmap.WriteTLV(buf, 0x0a, tmp)
	} else {
		gsmap.WriteTLV(buf, 0x0a, []byte{0x00})
	}

	// absentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 0(00)
	if rsds.AbsentDiag != 0 && rsds.AbsentDiag <= gsmap.TempUnavailable {
		gsmap.WriteTLV(buf, 0x80, []byte{rsds.AbsentDiag.ToByte()})
	}

	// extensionContainer, context_specific(80) + constructed(20) + 1(01)

	// gprsSupportIndicator, context_specific(80) + primitive(00) + 2(02)
	if rsds.SupportGPRS {
		gsmap.WriteTLV(buf, 0x82, nil)
	}

	// deliveryOutcomeIndicator, context_specific(80) + primitive(00) + 3(03)
	if rsds.OutcomeIsGPRS {
		gsmap.WriteTLV(buf, 0x83, nil)
	}

	// additionalSM-DeliveryOutcome, context_specific(80) + primitive(00) + 4(04)
	if tmp := rsds.AdditionalOutcome.marshal(); tmp != nil {
		gsmap.WriteTLV(buf, 0x84, tmp)
	} else {
		gsmap.WriteTLV(buf, 0x84, []byte{0x00})
	}

	// additionalAbsentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 5(05)
	if rsds.AdditionalDiag != 0 && rsds.AdditionalDiag <= gsmap.TempUnavailable {
		gsmap.WriteTLV(buf, 0x85, []byte{rsds.AdditionalDiag.ToByte()})
	}

	// ReportSM-DeliveryStatusArg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (ReportSmDeliveryStatusArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// ReportSmDeliveryStatus-Arg, universal(00) + constructed(20) + sequence(10)
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}
	rsds := ReportSmDeliveryStatusArg{InvokeID: id}

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if rsds.MSISDN, e = gsmap.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// serviceCentreAddress, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if rsds.CenterAddr, e = gsmap.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// sm-DeliveryOutcome, universal(00) + primitive(00) + enum(0a)
	if _, v, e := gsmap.ReadTLV(buf, 0x0a); e != nil {
		return nil, e
	} else if e = rsds.Outcome.unmarshal(v); e != nil {
		return nil, e
	}

	// absentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 0(00)
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return rsds, nil
	} else if e != nil {
		return nil, e
	} else if t == 0x80 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedTLV("invalid parameter value")
		}
		rsds.AbsentDiag.FromByte(v[0])

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rsds, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, context_specific(80) + constructed(20) + 1(01)
	if t == 0x81 {
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rsds, nil
		} else if e != nil {
			return nil, e
		}
	}

	// gprsSupportIndicator, context_specific(80) + primitive(00) + 2(02)
	if t == 0x82 {
		rsds.SupportGPRS = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rsds, nil
		} else if e != nil {
			return nil, e
		}
	}

	// deliveryOutcomeIndicator, context_specific(80) + primitive(00) + 3(03)
	if t == 0x83 {
		rsds.OutcomeIsGPRS = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rsds, nil
		} else if e != nil {
			return nil, e
		}
	}

	// additionalSM-DeliveryOutcome, context_specific(80) + primitive(00) + 4(04)
	if t == 0x84 {
		if e = rsds.AdditionalOutcome.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rsds, nil
		} else if e != nil {
			return nil, e
		}
	}

	// additionalAbsentSubscriberDiagnosticSM, context_specific(80) + primitive(00) + 5(05)
	if t == 0x85 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedTLV("invalid parameter value")
		}
		rsds.AdditionalDiag.FromByte(v[0])

		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return rsds, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return rsds, nil
}

/*
ReportSmDeliveryStatusRes operation res.

	ReportSM-DeliveryStatusRes ::= SEQUENCE {
		storedMSISDN       ISDN-AddressString OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		...}
*/
type ReportSmDeliveryStatusRes struct {
	InvokeID int8 `json:"id"`

	MSISDN gsmap.AddressString `json:"storedMSISDN,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (rsds ReportSmDeliveryStatusRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", rsds.Name(), rsds.InvokeID)
	if !rsds.MSISDN.IsEmpty() {
		fmt.Fprintf(buf, "\n%smsisdn: %s", gsmap.LogPrefix, rsds.MSISDN)
	}
	return buf.String()
}

func (rsds ReportSmDeliveryStatusRes) MarshalJSON() ([]byte, error) {
	j := struct {
		InvokeID int8                 `json:"id"`
		MSISDN   *gsmap.AddressString `json:"storedMSISDN,omitempty"`
		// Extension
	}{
		InvokeID: rsds.InvokeID}
	if !rsds.MSISDN.IsEmpty() {
		j.MSISDN = &rsds.MSISDN
	}
	return json.Marshal(j)
}

func (rsds ReportSmDeliveryStatusRes) GetInvokeID() int8 { return rsds.InvokeID }
func (ReportSmDeliveryStatusRes) Code() byte             { return 47 }
func (ReportSmDeliveryStatusRes) Name() string           { return "ReportSM-DeliveryStatus-Res" }

func (ReportSmDeliveryStatusRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		ReportSmDeliveryStatusRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.ReportSmDeliveryStatusRes, e
	}
	c := tmp.ReportSmDeliveryStatusRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (rsds ReportSmDeliveryStatusRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// storedMSISDN, universal(00) + primitive(00) + octet_string(04)
	if !rsds.MSISDN.IsEmpty() {
		gsmap.WriteTLV(buf, 0x04, rsds.MSISDN.Bytes())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// ReportSM-DeliveryStatusRes, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (ReportSmDeliveryStatusRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// ReportSmDeliveryStatus-Res, universal(00) + constructed(20) + sequence(10)
	rsds := ReportSmDeliveryStatusRes{InvokeID: id}
	if buf.Len() == 0 {
		return rsds, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return rsds, nil
	} else if e != nil {
		return nil, e
	}

	// storedMSISDN, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		if rsds.MSISDN, e = gsmap.DecodeAddressString(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rsds, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}

	return rsds, nil
}
