package ifc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/teldata"
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

	MSISDN      gsmap.AddressString `json:"msisdn"`
	SMRPPRI     bool                `json:"sm-RP-PRI"`
	CenterAddr  gsmap.AddressString `json:"serviceCentreAddress"`
	Teleservice *uint8              `json:"teleservice,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	SupportGPRS bool              `json:"gprsSupportIndicator,omitempty"`
	SMRPMTI     SMRPMTI           `json:"sm-RP-MTI,omitempty"`
	SMRPSMEA    gsmap.OctetString `json:"sm-RP-SMEA,omitempty"`
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
	} else if sri.MSISDN, e = gsmap.DecodeAddressString(v); e != nil {
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
	} else if sri.CenterAddr, e = gsmap.DecodeAddressString(v); e != nil {
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
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
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
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
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
