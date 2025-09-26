package ifd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/teldata"
)

/*
deleteSubscriberData  OPERATION ::= {	--Timer m
	ARGUMENT
		DeleteSubscriberDataArg
	RESULT
		DeleteSubscriberDataRes -- optional
	ERRORS {
		dataMissing            |
		unexpectedDataValue    |
		unidentifiedSubscriber }
	CODE	local:8 }
*/

func init() {
	a := DeleteSubscriberDataArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := DeleteSubscriberDataRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
DeleteSubscriberDataArg operation arg.

	DeleteSubscriberDataArg ::= SEQUENCE {
		imsi                           [0]  IMSI,
		basicServiceList               [1]  BasicServiceList       OPTIONAL,
		-- The exception handling for reception of unsupported/not allocated basicServiceCodes is defined in clause 6.8.2
		ss-List                        [2]  SS-List                OPTIONAL,
		roamingRestrictionDueToUnsupportedFeature [4] NULL         OPTIONAL,
		regionalSubscriptionIdentifier [5]  ZoneCode               OPTIONAL,
		vbsGroupIndication             [7]  NULL                   OPTIONAL,
		vgcsGroupIndication            [8]  NULL                   OPTIONAL,
		camelSubscriptionInfoWithdraw  [9]  NULL                   OPTIONAL,
		extensionContainer             [6]  ExtensionContainer     OPTIONAL,
		...,
		gprsSubscriptionDataWithdraw   [10] GPRSSubscriptionDataWithdraw OPTIONAL,
		roamingRestrictedInSgsnDueToUnsuppportedFeature [11] NULL  OPTIONAL,
		lsaInformationWithdraw         [12] LSAInformationWithdraw OPTIONAL,
		gmlc-ListWithdraw              [13] NULL                   OPTIONAL,
		istInformationWithdraw         [14] NULL                   OPTIONAL,
		specificCSI-Withdraw           [15] SpecificCSI-Withdraw   OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		chargingCharacteristicsWithdraw	[16] NULL              OPTIONAL,
		stn-srWithdraw                  [17] NULL              OPTIONAL,
		epsSubscriptionDataWithdraw     [18] EPS-SubscriptionDataWithdraw OPTIONAL,
		apn-oi-replacementWithdraw      [19] NULL              OPTIONAL,
		csg-SubscriptionDeleted         [20] NULL              OPTIONAL,
		subscribedPeriodicTAU-RAU-TimerWithdraw	[22] NULL      OPTIONAL,
		subscribedPeriodicLAU-TimerWithdraw     [23] NULL      OPTIONAL,
		subscribed-vsrvccWithdraw       [21] NULL              OPTIONAL,
		vplmn-Csg-SubscriptionDeleted   [24] NULL              OPTIONAL,
		additionalMSISDN-Withdraw       [25] NULL              OPTIONAL,
		cs-to-ps-SRVCC-Withdraw         [26] NULL              OPTIONAL,
		imsiGroupIdList-Withdraw        [27] NULL              OPTIONAL,
		userPlaneIntegrityProtectionWithdraw         [28] NULL OPTIONAL,
		dl-Buffering-Suggested-Packet-Count-Withdraw [29] NULL OPTIONAL,
		ue-UsageTypeWithdraw            [30] NULL              OPTIONAL,
		reset-idsWithdraw               [31] NULL              OPTIONAL,
		iab-OperationWithdraw           [32] NULL              OPTIONAL }

	BasicServiceList ::= SEQUENCE SIZE (1..maxNumOfBasicServices) OF Ext-BasicServiceCode
*/
type DeleteSubscriberDataArg struct {
	InvokeID int8 `json:"id"`

	IMSI                                      teldata.IMSI `json:"imsi"`
	BasicServiceList                          []svcCode    `json:"basicServiceList,omitempty"`
	SsList                                    []uint8      `json:"ss-List,omitempty"`
	RoamingRestrictionDueToUnsupportedFeature bool         `json:"roamingRestrictionDueToUnsupportedFeature,omitempty"`
	RegionalSubscriptionIdentifier            data16       `json:"regionalSubscriptionIdentifier,omitempty"`
	VBSGroupIndication                        bool         `json:"vbsGroupIndication,omitempty"`
	VGCSGroupIndication                       bool         `json:"vgcsGroupIndication,omitempty"`
	CamelSubscriptionInfoWithdraw             bool         `json:"camelSubscriptionInfoWithdraw,omitempty"`
	// ExtensionContainer                     ExtensionContainer `json:"extensionContainer,omitempty"`
	// GPRSSubscriptionDataWithdraw                    gprsSubscriptionDataWithdraw `json:"gprsSubscriptionDataWithdraw,omitempty"`
	RoamingRestrictedInSgsnDueToUnsuppportedFeature bool `json:"roamingRestrictedInSgsnDueToUnsuppportedFeature,omitempty"`
	// LSAInformationWithdraw                          lsaInformationWithdraw       `json:"lsaInformationWithdraw,omitempty"`
	GMLCListWithdraw       bool                `json:"gmlc-ListWithdraw,omitempty"`
	ISTInformationWithdraw bool                `json:"istInformationWithdraw,omitempty"`
	SpecificCSIWithdraw    specificCSIWithdraw `json:"specificCSI-Withdraw,omitempty"`
}

func (dsd DeleteSubscriberDataArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", dsd.Name(), dsd.InvokeID)

	if !dsd.IMSI.IsEmpty() {
		fmt.Fprintf(buf, "\n%simsi: %s", gsmap.LogPrefix, dsd.IMSI)
	}
	for _, s := range dsd.BasicServiceList {
		fmt.Fprintf(buf, "\n%s| | basicServiceGroup: %x", gsmap.LogPrefix, s)
	}
	if len(dsd.SsList) != 0 {
		fmt.Fprintf(buf, "\n%sss-List: %v", gsmap.LogPrefix, dsd.SsList)
	}
	if dsd.RoamingRestrictionDueToUnsupportedFeature {
		fmt.Fprintf(buf, "\n%sroamingRestrictionDueToUnsupportedFeature:", gsmap.LogPrefix)
	}
	if len(dsd.RegionalSubscriptionIdentifier) != 0 {
		fmt.Fprintf(buf, "\n%sregionalSubscriptionIdentifier: %x",
			gsmap.LogPrefix, dsd.RegionalSubscriptionIdentifier)
	}
	if dsd.VBSGroupIndication {
		fmt.Fprintf(buf, "\n%svbsGroupIndication:", gsmap.LogPrefix)
	}
	if dsd.VGCSGroupIndication {
		fmt.Fprintf(buf, "\n%svgcsGroupIndication:", gsmap.LogPrefix)
	}
	if dsd.CamelSubscriptionInfoWithdraw {
		fmt.Fprintf(buf, "\n%scamelSubscriptionInfoWithdraw:", gsmap.LogPrefix)
	}
	// ExtensionContainer
	// GPRSSubscriptionDataWithdraw
	if dsd.RoamingRestrictedInSgsnDueToUnsuppportedFeature {
		fmt.Fprintf(buf, "\n%sroamingRestrictedInSgsnDueToUnsuppportedFeature:", gsmap.LogPrefix)
	}
	// LSAInformationWithdraw
	if dsd.GMLCListWithdraw {
		fmt.Fprintf(buf, "\n%sgmlc-ListWithdraw:", gsmap.LogPrefix)
	}
	if dsd.ISTInformationWithdraw {
		fmt.Fprintf(buf, "\n%sistInformationWithdraw:", gsmap.LogPrefix)
	}
	if dsd.SpecificCSIWithdraw != 0 {
		fmt.Fprintf(buf, "\n%sspecificCSI-Withdraw: %s", gsmap.LogPrefix, dsd.SpecificCSIWithdraw)
	}
	return buf.String()
}

func (dsd DeleteSubscriberDataArg) GetInvokeID() int8            { return dsd.InvokeID }
func (DeleteSubscriberDataArg) GetLinkedID() *int8               { return nil }
func (DeleteSubscriberDataArg) Code() byte                       { return 8 }
func (DeleteSubscriberDataArg) Name() string                     { return "DeleteSubscriberData-Arg" }
func (DeleteSubscriberDataArg) DefaultContext() gsmap.AppContext { return SubscriberDataMngt1 }

func (DeleteSubscriberDataArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		DeleteSubscriberDataArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.DeleteSubscriberDataArg, e
	}
	c := tmp.DeleteSubscriberDataArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (dsd DeleteSubscriberDataArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// imsi, context_specific(80) + primitive(00) + 0(00)
	if !dsd.IMSI.IsEmpty() {
		gsmap.WriteTLV(buf, 0x80, dsd.IMSI.Bytes())
	}

	// basicServiceList, context_specific(80) + constructed(20) + 1(01)
	if len(dsd.BasicServiceList) != 0 {
		buf2 := new(bytes.Buffer)
		for _, sc := range dsd.BasicServiceList {
			switch sc.Type {
			case 2: // basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
				gsmap.WriteTLV(buf2, 0x82, []byte{sc.Code})
			case 3: // basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
				gsmap.WriteTLV(buf2, 0x83, []byte{sc.Code})
			}
		}
		gsmap.WriteTLV(buf, 0xa1, buf2.Bytes())
	}

	// ss-List, context_specific(80) + constructed(20) + 2(02)
	if len(dsd.SsList) != 0 {
		gsmap.WriteTLV(buf, 0xa2, marshalCodeList(dsd.SsList))
	}

	// roamingRestrictionDueToUnsupportedFeature, context_specific(80) + primitive(00) + 4(04)
	if dsd.RoamingRestrictionDueToUnsupportedFeature {
		gsmap.WriteTLV(buf, 0x84, nil)
	}

	// regionalSubscriptionIdentifier, context_specific(80) + primitive(00) + 5(05)
	if len(dsd.RegionalSubscriptionIdentifier) != 0 {
		gsmap.WriteTLV(buf, 0x85, dsd.RegionalSubscriptionIdentifier.marshal())
	}

	// vbsGroupIndication, context_specific(80) + primitive(00) + 7(07)
	if dsd.VBSGroupIndication {
		gsmap.WriteTLV(buf, 0x87, nil)
	}

	// vgcsGroupIndication, context_specific(80) + primitive(00) + 8(08)
	if dsd.VGCSGroupIndication {
		gsmap.WriteTLV(buf, 0x88, nil)
	}

	// camelSubscriptionInfoWithdraw, context_specific(80) + primitive(00) + 9(09)
	if dsd.CamelSubscriptionInfoWithdraw {
		gsmap.WriteTLV(buf, 0x89, nil)
	}

	// extensionContainer, context_specific(80) + constructed(20) + 6(06)

	// gprsSubscriptionDataWithdraw, context_specific(80) + primitive(00) + 10(0A)

	// roamingRestrictedInSgsnDueToUnsuppportedFeature, context_specific(80) + primitive(00) + 11(0B)
	if dsd.RoamingRestrictedInSgsnDueToUnsuppportedFeature {
		gsmap.WriteTLV(buf, 0x8b, nil)
	}

	// lsaInformationWithdraw, context_specific(80) + primitive(00) + 12(0C)

	// gmlc-ListWithdraw, context_specific(80) + primitive(00) + 13(0D)
	if dsd.GMLCListWithdraw {
		gsmap.WriteTLV(buf, 0x8d, nil)
	}

	// istInformationWithdraw, context_specific(80) + primitive(00) + 14(0E)
	if dsd.ISTInformationWithdraw {
		gsmap.WriteTLV(buf, 0x8e, nil)
	}

	// specificCSI-Withdraw, context_specific(80) + primitive(00) + 15(0F)
	if dsd.SpecificCSIWithdraw != 0 {
		gsmap.WriteTLV(buf, 0x8f, dsd.SpecificCSIWithdraw.marshal())
	}

	// DeleteSubscriberDataArg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (DeleteSubscriberDataArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// DeleteSubscriberDataArg, universal(00) + constructed(20) + sequence(10)
	dsd := DeleteSubscriberDataArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// imsi, context_specific(80) + primitive(00) + 0(00)
	if _, v, e := gsmap.ReadTLV(buf, 0x80); e != nil {
		return nil, e
	} else if dsd.IMSI, e = teldata.DecodeIMSI(v); e != nil {
		return nil, e
	}

	// optional TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return dsd, nil
	} else if e != nil {
		return nil, e
	}

	// basicServiceList, context_specific(80) + constructed(20) + 1(01)
	if t == 0xa1 {
		dsd.BasicServiceList = []svcCode{}
		buf2 := bytes.NewBuffer(v)
		for {
			t2, v2, e := gsmap.ReadTLV(buf2, 0x00)
			if e == io.EOF {
				break
			} else if e != nil {
				return nil, e
			}

			switch t2 {
			case 0x82:
				// basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
				if len(v2) == 0 {
					return nil, gsmap.UnexpectedTLV("length must >0")
				}
				dsd.BasicServiceList = append(dsd.BasicServiceList, svcCode{Type: 2, Code: v2[0]})
			case 0x83:
				// basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
				if len(v2) == 0 {
					return nil, gsmap.UnexpectedTLV("length must >0")
				}
				dsd.BasicServiceList = append(dsd.BasicServiceList, svcCode{Type: 3, Code: v2[0]})
			}
		}
	}

	// ss-List, context_specific(80) + constructed(20) + 2(02)
	if t == 0xa2 {
		if dsd.SsList, e = unmarshalCodeList(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// roamingRestrictionDueToUnsupportedFeature, context_specific(80) + primitive(00) + 4(04)
	if t == 0x84 {
		dsd.RoamingRestrictionDueToUnsupportedFeature = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// regionalSubscriptionIdentifier, context_specific(80) + primitive(00) + 5(05)
	if t == 0x85 {
		if e = dsd.RegionalSubscriptionIdentifier.unmarshal(v); e != nil {
			return nil, e
		}
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// vbsGroupIndication, context_specific(80) + primitive(00) + 7(07)
	if t == 0x87 {
		dsd.VBSGroupIndication = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// vgcsGroupIndication, context_specific(80) + primitive(00) + 8(08)
	if t == 0x88 {
		dsd.VGCSGroupIndication = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// camelSubscriptionInfoWithdraw, context_specific(80) + primitive(00) + 9(09)
	if t == 0x89 {
		dsd.CamelSubscriptionInfoWithdraw = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}
	// extensionContainer, context_specific(80) + constructed(20) + 6(06)

	// gprsSubscriptionDataWithdraw, context_specific(80) + primitive(00) + 10(0A)

	// roamingRestrictedInSgsnDueToUnsuppportedFeature, context_specific(80) + primitive(00) + 11(0B)
	if t == 0x8b {
		dsd.RoamingRestrictedInSgsnDueToUnsuppportedFeature = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// lsaInformationWithdraw, context_specific(80) + primitive(00) + 12(0C)

	// gmlc-ListWithdraw, context_specific(80) + primitive(00) + 13(0D)
	if t == 0x8d {
		dsd.GMLCListWithdraw = true
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// istInformationWithdraw, context_specific(80) + primitive(00) + 14(0E)
	if t == 0x8e {
		dsd.ISTInformationWithdraw = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return dsd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// specificCSI-Withdraw, context_specific(80) + primitive(00) + 15(0F)
	if t == 0x8f {
		if e = dsd.SpecificCSIWithdraw.unmarshal(v); e != nil {
			return nil, e
		}
		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return dsd, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return dsd, nil
}

/*
specificCSIWithdraw

	SpecificCSI-Withdraw ::= BIT STRING {
		o-csi (0),
		ss-csi (1),
		tif-csi (2),
		d-csi (3),
		vt-csi (4),
		mo-sms-csi (5),
		m-csi (6),
		gprs-csi (7),
		-- ^^^^^^ R99 ^^^^^^
		t-csi (8),
		mt-sms-csi (9),
		mg-csi (10),
		o-IM-CSI (11),
		d-IM-CSI (12),
		vt-IM-CSI (13) } (SIZE(8..32))

	-- exception handling:
	-- bits 11 to 31 shall be ignored if received by a non-IP Multimedia Core Network entity.
	-- bits 0-10 and 14-31 shall be ignored if received by an IP Multimedia Core Network entity.
	-- bits 11-13 are only applicable in an IP Multimedia Core Network.
	-- Bit 8 and bits 11-13 are only applicable for the NoteSubscriberDataModified operation.
*/
type specificCSIWithdraw uint8

func (o specificCSIWithdraw) String() string {
	return fmt.Sprintf("%08b", o)
}

func (o *specificCSIWithdraw) UnmarshalJSON(b []byte) (e error) {
	var s string
	if e = json.Unmarshal(b, &s); e != nil {
		return
	}
	tmp, e := strconv.ParseUint(s, 2, 8)
	if e == nil {
		*o = specificCSIWithdraw(tmp)
	}
	return
}

func (o specificCSIWithdraw) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(o), 2))
}

func (o *specificCSIWithdraw) unmarshal(b []byte) error {
	if len(b) < 2 {
		return gsmap.UnexpectedTLV("length must <2")
	}
	*o = specificCSIWithdraw(b[1])
	return nil
}

func (o specificCSIWithdraw) marshal() []byte {
	return []byte{0x01, byte(o)}
}

/*
DeleteSubscriberDataRes

	DeleteSubscriberDataRes ::= SEQUENCE {
		regionalSubscriptionResponse [0] RegionalSubscriptionResponse OPTIONAL,
		extensionContainer               ExtensionContainer           OPTIONAL,
		...}
*/
type DeleteSubscriberDataRes struct {
	InvokeID                     int8                    `json:"id"`
	RegionalSubscriptionResponse regionalSubscriptionRes `json:"regionalSubscriptionResponse,omitempty"`
	// ExtensionContainer           ExtensionContainer           `json:"extensionContainer,omitempty"`
}

func (dsd DeleteSubscriberDataRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", dsd.Name(), dsd.InvokeID)
	if dsd.RegionalSubscriptionResponse != 0 {
		fmt.Fprintf(buf, "\n%sregionalSubscriptionResponse: %s", gsmap.LogPrefix, dsd.RegionalSubscriptionResponse)
	}
	return buf.String()
}
func (dsd DeleteSubscriberDataRes) GetInvokeID() int8 { return dsd.InvokeID }
func (DeleteSubscriberDataRes) Code() byte            { return 8 }
func (DeleteSubscriberDataRes) Name() string          { return "DeleteSubscriberData-Res" }

func (DeleteSubscriberDataRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		DeleteSubscriberDataRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.DeleteSubscriberDataRes, e
	}
	c := tmp.DeleteSubscriberDataRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (dsd DeleteSubscriberDataRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)
	// regionalSubscriptionResponse, context_specific(80) + primitive(00) + 0(00)
	if dsd.RegionalSubscriptionResponse != 0 {
		gsmap.WriteTLV(buf, 0x80, dsd.RegionalSubscriptionResponse.marshal())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// DeleteSubscriberDataRes, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (DeleteSubscriberDataRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// DeleteSubscriberDataRes, universal(00) + constructed(20) + sequence(10)
	dsd := DeleteSubscriberDataRes{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// optional TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return dsd, nil
	} else if e != nil {
		return nil, e
	}

	// regionalSubscriptionResponse, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if e = dsd.RegionalSubscriptionResponse.unmarshal(v); e != nil {
			return nil, e
		}
		// if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
		// 	return dsd, nil
		// } else if e != nil {
		// 	return nil, e
		// }
	}
	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return dsd, nil
}
