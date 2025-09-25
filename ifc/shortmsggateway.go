package ifc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/common"
	"github.com/fkgi/teldata"
)

const (
	ShortMsgGateway1 gsmap.AppContext = 0x0004000001001401
	ShortMsgGateway2 gsmap.AppContext = 0x0004000001001402
	ShortMsgGateway3 gsmap.AppContext = 0x0004000001001403
)

/*
sendRoutingInfoForSM  OPERATION ::= {	--Timer m
	ARGUMENT
		RoutingInfoForSM-Arg
	RESULT
		RoutingInfoForSM-Res
	ERRORS {
		systemFailure             |
		dataMissing               |
		unexpectedDataValue       |
		facilityNotSupported      |
		unknownSubscriber         |
		teleserviceNotProvisioned |
		callBarred                |
		absentSubscriber          |
		absentSubscriberSM }
	CODE	local:45 }

absentSubscriber is for version 1.
absentSubscriberSM is for version >1.
*/

func init() {
	a := RoutingInfoForSmArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := RoutingInfoForSmRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
RoutingInfoForSmArg operation arg.
gprsSupportIndicator is set only if the SMS-GMSC supports receiving of two numbers from the HLR.
teleservice must be absent in version greater 1.

	RoutingInfoForSM-Arg ::= SEQUENCE {
		msisdn                    [0]  ISDN-AddressString,
		sm-RP-PRI                 [1]  BOOLEAN,
		serviceCentreAddress      [2]  AddressString,
		teleservice               [5]  TeleserviceCode        OPTIONAL,
		extensionContainer        [6]  ExtensionContainer     OPTIONAL,
		... ,
		gprsSupportIndicator      [7]  NULL                   OPTIONAL,
		sm-RP-MTI                 [8]  SM-RP-MTI              OPTIONAL,
		sm-RP-SMEA                [9]  SM-RP-SMEA             OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		sm-deliveryNotIntended    [10] SM-DeliveryNotIntended OPTIONAL,
		ip-sm-gwGuidanceIndicator [11] NULL                   OPTIONAL,
		imsi                      [12] IMSI                   OPTIONAL,
		t4-Trigger-Indicator      [14] NULL                   OPTIONAL,
		singleAttemptDelivery     [13] NULL                   OPTIONAL,
		correlationID             [15] CorrelationID          OPTIONAL,
		smsf-supportIndicator     [16] NULL                   OPTIONAL }

	SM-RP-SMEA ::= OCTET STRING (SIZE (1..12))
		-- this parameter contains an address field which is encoded
		-- as defined in 3GPP TS 23.040. An address field contains 3 elements :
		--	address-length
		--	type-of-address
		--	address-value
*/
type RoutingInfoForSmArg struct {
	InvokeID int8 `json:"id"`

	MSISDN      common.AddressString `json:"msisdn"`
	SMRPPRI     bool                 `json:"sm-RP-PRI"`
	CenterAddr  common.AddressString `json:"serviceCentreAddress"`
	Teleservice *uint8               `json:"teleservice,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	SupportGPRS bool               `json:"gprsSupportIndicator,omitempty"`
	SMRPMTI     SMRPMTI            `json:"sm-RP-MTI,omitempty"`
	SMRPSMEA    common.OctetString `json:"sm-RP-SMEA,omitempty"`
}

func (sri RoutingInfoForSmArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", sri.Name(), sri.InvokeID)
	fmt.Fprintf(buf, "\n%smsisdn:               %s", gsmap.LogPrefix, sri.MSISDN)
	fmt.Fprintf(buf, "\n%ssm-RP-PRI:            %t", gsmap.LogPrefix, sri.SMRPPRI)
	fmt.Fprintf(buf, "\n%sserviceCentreAddress: %s", gsmap.LogPrefix, sri.CenterAddr)
	if sri.Teleservice != nil {
		fmt.Fprintf(buf, "\n%steleservice         :%x", gsmap.LogPrefix, *sri.Teleservice)
	}
	// Extension
	if sri.SupportGPRS {
		fmt.Fprint(buf, "\n| gprsSupportIndicator:")
	}
	if sri.SMRPMTI != 0 {
		fmt.Fprint(buf, "\n| sm-RP-MTI:           ", sri.SMRPMTI)
	}
	if len(sri.SMRPSMEA) != 0 {
		fmt.Fprint(buf, "\n| sm-RP-SMEA:          ", sri.SMRPSMEA)
	}
	return buf.String()
}

func (sri RoutingInfoForSmArg) GetInvokeID() int8            { return sri.InvokeID }
func (RoutingInfoForSmArg) GetLinkedID() *int8               { return nil }
func (RoutingInfoForSmArg) Code() byte                       { return 45 }
func (RoutingInfoForSmArg) Name() string                     { return "RoutingInfoForSM-Arg" }
func (RoutingInfoForSmArg) DefaultContext() gsmap.AppContext { return ShortMsgGateway1 }

func (RoutingInfoForSmArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		RoutingInfoForSmArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.RoutingInfoForSmArg, e
	}
	c := tmp.RoutingInfoForSmArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (sri RoutingInfoForSmArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// msisdn, context_specific(80) + primitive(00) + 0(00)
	gsmap.WriteTLV(buf, 0x80, sri.MSISDN.Bytes())

	// sm-RP-PRI, context_specific(80) + primitive(00) + 1(01)
	if sri.SMRPPRI {
		gsmap.WriteTLV(buf, 0x81, []byte{0x01})
	} else {
		gsmap.WriteTLV(buf, 0x81, []byte{0x00})
	}

	// serviceCentreAddress, context_specific(80) + primitive(00) + 2(02)
	gsmap.WriteTLV(buf, 0x82, sri.CenterAddr.Bytes())

	// teleservice, context_specific(80) + primitive(00) + 5(05)
	if sri.Teleservice != nil {
		gsmap.WriteTLV(buf, 0x85, []byte{*sri.Teleservice})
	}

	// extensionContainer, context_specific(80) + constructed(20) + 6(06)

	// gprsSupportIndicator, context_specific(80) + primitive(00) + 7(07)
	if sri.SupportGPRS {
		gsmap.WriteTLV(buf, 0x87, nil)
	}

	// sm-RP-MTI, context_specific(80) + primitive(00) + 8(08)
	if tmp := sri.SMRPMTI.marshal(); tmp != nil {
		gsmap.WriteTLV(buf, 0x88, tmp)
	}

	// sm-RP-SMEA, context_specific(80) + primitive(00) + 9(09)
	if len(sri.SMRPSMEA) != 0 {
		gsmap.WriteTLV(buf, 0x89, sri.SMRPSMEA)
	}

	// RoutingInfoForSM-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (RoutingInfoForSmArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// RoutingInfoForSM-Arg, universal(00) + constructed(20) + sequence(10)
	sri := RoutingInfoForSmArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// msisdn, context_specific(80) + primitive(00) + 0(00)
	if _, v, e := gsmap.ReadTLV(buf, 0x80); e != nil {
		return nil, e
	} else if sri.MSISDN, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// sm-RP-PRI, context_specific(80) + primitive(00) + 1(01)
	if _, v, e := gsmap.ReadTLV(buf, 0x81); e != nil {
		return nil, e
	} else if len(v) != 1 {
		return nil, gsmap.UnexpectedTLV("invalid parameter value")
	} else {
		sri.SMRPPRI = v[0] != 0x00
	}

	// serviceCentreAddress, context_specific(80) + primitive(00) + 2(02)
	if _, v, e := gsmap.ReadTLV(buf, 0x82); e != nil {
		return nil, e
	} else if sri.CenterAddr, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return sri, nil
	} else if e != nil {
		return nil, e
	}

	// teleservice, context_specific(80) + primitive(00) + 5(05)
	if t == 0x85 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedEnumValue(v)
		}
		sri.Teleservice = &(v[0])

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return sri, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, context_specific(80) + constructed(20) + 6(06)
	if t == 0x86 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return sri, nil
		} else if e != nil {
			return nil, e
		}
	}

	// gprsSupportIndicator, context_specific(80) + primitive(00) + 7(07)
	if t == 0x87 {
		sri.SupportGPRS = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return sri, nil
		} else if e != nil {
			return nil, e
		}
	}

	// sm-RP-MTI, context_specific(80) + primitive(00) + 8(08)
	if t == 0x88 {
		if e = sri.SMRPMTI.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return sri, nil
		} else if e != nil {
			return nil, e
		}
	}

	// sm-RP-SMEA, context_specific(80) + primitive(00) + 9(09)
	if t == 0x89 {
		sri.SMRPSMEA = v
		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return sri, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return sri, nil
}

/*
RoutingInfoForSmRes operation res.
mwd-Set must be absent in version greater 1.

	RoutingInfoForSM-Res ::= SEQUENCE {
		imsi                     IMSI,
		locationInfoWithLMSI [0] LocationInfoWithLMSI,
		mwd-Set              [2] BOOLEAN              OPTIONAL,
		extensionContainer   [4] ExtensionContainer   OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		ip-sm-gwGuidance     [5] IP-SM-GW-Guidance    OPTIONAL }
*/
type RoutingInfoForSmRes struct {
	InvokeID int8 `json:"id"`

	IMSI         teldata.TBCD         `json:"imsi"`
	LocationInfo LocationInfoWithLMSI `json:"locationInfoWithLMSI"`
	MWD          bool                 `json:"mwd-Set,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (sri RoutingInfoForSmRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", sri.Name(), sri.InvokeID)
	fmt.Fprintf(buf, "\n%simsi: %s", gsmap.LogPrefix, sri.IMSI)
	fmt.Fprintf(buf, "\n%slocationInfoWithLMSI:", gsmap.LogPrefix)
	fmt.Fprintf(buf, "\n%s| networkNode-Number: %s",
		gsmap.LogPrefix, sri.LocationInfo.NodeNumber.Address)
	if !sri.LocationInfo.LMSI.IsEmpty() {
		fmt.Fprintf(buf, "\n%s| lmsi:               %s",
			gsmap.LogPrefix, sri.LocationInfo.LMSI)
	}
	// Extension
	if sri.LocationInfo.NodeNumber.IsGPRS {
		fmt.Fprintf(buf, "\n%s| gprsNodeIndicator: ", gsmap.LogPrefix)
	}
	if sri.LocationInfo.AdditionalNumber.Address.Digits.Length() != 0 {
		nt := ""
		if sri.LocationInfo.AdditionalNumber.IsGPRS {
			nt = "SGSN"
		} else {
			nt = "MSC"
		}
		fmt.Fprintf(buf, "\n%s| additional-Number: %s(%s)",
			gsmap.LogPrefix, sri.LocationInfo.AdditionalNumber.Address, nt)
	}
	if sri.MWD {
		fmt.Fprintf(buf, "\n%smwd-Set           :%t", gsmap.LogPrefix, sri.MWD)
	}
	return buf.String()
}

func (sri RoutingInfoForSmRes) GetInvokeID() int8 { return sri.InvokeID }
func (RoutingInfoForSmRes) Code() byte            { return 45 }
func (RoutingInfoForSmRes) Name() string          { return "RoutingInfoForSM-Res" }

func (RoutingInfoForSmRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		RoutingInfoForSmRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.RoutingInfoForSmRes, e
	}
	c := tmp.RoutingInfoForSmRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (sri RoutingInfoForSmRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// imsi, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, sri.IMSI.Bytes())

	// locationInfoWithLMSI, context_specific(80) + constructed(20) + 0(00)
	gsmap.WriteTLV(buf, 0xa0, sri.LocationInfo.marshal())

	// mwd-Set, context_specific(80) + primitive(00) + 2(02)
	if sri.MWD {
		gsmap.WriteTLV(buf, 0x82, []byte{0x01})
	}

	// extensionContainer, context_specific(80) + constructed(20) + 4(04)

	// RoutingInfoForSm-Res, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (RoutingInfoForSmRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// RoutingInfoForSM-Res, universal(00) + constructed(20) + sequence(10)
	sri := RoutingInfoForSmRes{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// imsi, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else {
		sri.IMSI = v
	}

	// locationInfoWithLMSI, context_specific(80) + constructed(20) + 0(00)
	if _, v, e := gsmap.ReadTLV(buf, 0xa0); e != nil {
		return nil, e
	} else if e = sri.LocationInfo.unmarshal(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return sri, nil
	} else if e != nil {
		return nil, e
	}

	// mwd-Set, context_specific(80) + primitive(00) + 2(02)
	if t == 0x82 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedTLV("invalid parameter value")
		}
		sri.MWD = v[0] != 0x00

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return sri, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, context_specific(80) + constructed(20) + 4(04)
	if t == 0xa4 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}
		/*
			if t, _, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return sri, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return sri, nil
}

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

	MSISDN     common.AddressString `json:"msisdn"`
	CenterAddr common.AddressString `json:"serviceCentreAddress"`
	Outcome    Outcome              `json:"sm-DeliveryOutcome"`
	AbsentDiag common.AbsentDiag    `json:"absentSubscriberDiagnosticSM,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	SupportGPRS       bool              `json:"gprsSupportIndicator,omitempty"`
	OutcomeIsGPRS     bool              `json:"deliveryOutcomeIndicator,omitempty"`
	AdditionalOutcome Outcome           `json:"additionalSM-DeliveryOutcome,omitempty"`
	AdditionalDiag    common.AbsentDiag `json:"additionalAbsentSubscriberDiagnosticSM,omitempty"`
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
	if rsds.AbsentDiag != 0 && rsds.AbsentDiag <= common.TempUnavailable {
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
	if rsds.AdditionalDiag != 0 && rsds.AdditionalDiag <= common.TempUnavailable {
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
	} else if rsds.MSISDN, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// serviceCentreAddress, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if rsds.CenterAddr, e = common.DecodeAddressString(v); e != nil {
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

	MSISDN common.AddressString `json:"storedMSISDN,omitempty"`
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
		InvokeID int8                  `json:"id"`
		MSISDN   *common.AddressString `json:"storedMSISDN,omitempty"`
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
		if rsds.MSISDN, e = common.DecodeAddressString(v); e != nil {
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
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}

	return rsds, nil
}

/*
informServiceCentre  OPERATION ::= {	--Timer s
	ARGUMENT
		InformServiceCentreArg
	CODE	local:63 }
*/

func init() {
	a := InformServiceCentreArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a
}

/*
InformServiceCentreArg operation arg.
InformServiceCentre must not be used in version 1.

	InformServiceCentreArg ::= SEQUENCE {
		storedMSISDN       ISDN-AddressString OPTIONAL,
		mw-Status          MW-Status          OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		... ,
		-- ^^^^^^^^ R99 ^^^^^^^^
		absentSubscriberDiagnosticSM                AbsentSubscriberDiagnosticSM OPTIONAL,
		additionalAbsentSubscriberDiagnosticSM  [0] AbsentSubscriberDiagnosticSM OPTIONAL,
		smsf3gppAbsentSubscriberDiagnosticSM    [1] AbsentSubscriberDiagnosticSM OPTIONAL,
		smsfNon3gppAbsentSubscriberDiagnosticSM [2] AbsentSubscriberDiagnosticSM OPTIONAL }
*/
type InformServiceCentreArg struct {
	InvokeID int8 `json:"id"`

	MSISDN   common.AddressString `json:"storedMSISDN,omitempty"`
	MWStatus MWStatus             `json:"mw-Status,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (isc InformServiceCentreArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", isc.Name(), isc.InvokeID)
	if !isc.MSISDN.IsEmpty() {
		fmt.Fprintf(buf, "\n%smsisdn:    %s", gsmap.LogPrefix, isc.MSISDN)
	}
	if isc.MWStatus != 0 {
		fmt.Fprintf(buf, "\n%sMW-Status: %s", gsmap.LogPrefix, isc.MWStatus)
	}
	return buf.String()
}

func (isc InformServiceCentreArg) MarshalJSON() ([]byte, error) {
	j := struct {
		InvokeID int8                  `json:"id"`
		MSISDN   *common.AddressString `json:"storedMSISDN,omitempty"`
		MWStatus MWStatus              `json:"mw-Status,omitempty"`
		// Extension
	}{
		InvokeID: isc.InvokeID,
		MWStatus: isc.MWStatus}
	if !isc.MSISDN.IsEmpty() {
		j.MSISDN = &isc.MSISDN
	}
	return json.Marshal(j)
}

func (isc InformServiceCentreArg) GetInvokeID() int8            { return isc.InvokeID }
func (InformServiceCentreArg) GetLinkedID() *int8               { return nil }
func (InformServiceCentreArg) Code() byte                       { return 63 }
func (InformServiceCentreArg) Name() string                     { return "InformServiceCentre-Arg" }
func (InformServiceCentreArg) DefaultContext() gsmap.AppContext { return 0 }

func (InformServiceCentreArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		InformServiceCentreArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.InformServiceCentreArg, e
	}
	c := tmp.InformServiceCentreArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (isc InformServiceCentreArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// storedMSISDN, universal(00) + primitive(00) + octet_string(04)
	if !isc.MSISDN.IsEmpty() {
		gsmap.WriteTLV(buf, 0x04, isc.MSISDN.Bytes())
	}

	// mw-Status, universal(00) + primitive(00) + bit_string(03)
	if isc.MWStatus != 0 {
		gsmap.WriteTLV(buf, 0x03, isc.MWStatus.marshal())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// InformServiceCentre-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (InformServiceCentreArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// InformServiceCentre-Arg, universal(00) + constructed(20) + sequence(10)
	isc := InformServiceCentreArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return isc, nil
	} else if e != nil {
		return nil, e
	}

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		if isc.MSISDN, e = common.DecodeAddressString(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isc, nil
		} else if e != nil {
			return nil, e
		}
	}

	// mw-Status, universal(00) + primitive(00) + bit_string(03)
	if t == 0x03 {
		if e = isc.MWStatus.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isc, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}
		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return isc, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return isc, nil
}
