package ifd

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
	SubscriberDataMngt1 gsmap.AppContext = 0x0004000001001001
	SubscriberDataMngt2 gsmap.AppContext = 0x0004000001001002
	SubscriberDataMngt3 gsmap.AppContext = 0x0004000001001003
)

/*
insertSubscriberData  OPERATION ::= {
	ARGUMENT
		InsertSubscriberDataArg
	RESULT
		InsertSubscriberDataRes -- optional
	ERRORS {
		dataMissing |
		unexpectedDataValue |
		unidentifiedSubscriber}
	CODE local:7 }
*/

func init() {
	a := InsertSubscriberDataArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := InsertSubscriberDataRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
InsertSubscriberDataArg operation arg.

	InsertSubscriberDataArg ::= SEQUENCE {
		imsi                           [0]  IMSI                           OPTIONAL,
		msisdn                         [1]  ISDN-AddressString             OPTIONAL,
		category                       [2]  Category                       OPTIONAL,
		subscriberStatus               [3]  SubscriberStatus               OPTIONAL,
		bearerServiceList              [4]  BearerServiceList              OPTIONAL,
		-- The exception handling for reception of unsupported / not allocated
		-- bearerServiceCodes is defined in clause 8.8.1
		teleserviceList                [6]  TeleserviceList                OPTIONAL,
		-- The exception handling for reception of unsupported / not allocated
		-- teleserviceCodes is defined in clause 8.8.1
		provisionedSS                  [7]  Ext-SS-InfoList                OPTIONAL,
		odb-Data                       [8]  ODB-Data                       OPTIONAL,
		roamingRestrictionDueToUnsupportedFeature [9]  NULL                OPTIONAL,
		regionalSubscriptionData       [10] ZoneCodeList                   OPTIONAL,
		vbsSubscriptionData            [11] VBSDataList                    OPTIONAL,
		vgcsSubscriptionData           [12] VGCSDataList                   OPTIONAL,
		vlrCamelSubscriptionInfo       [13] VlrCamelSubscriptionInfo       OPTIONAL
		extensionContainer             [14] ExtensionContainer             OPTIONAL,
		... ,
		naea-PreferredCI               [15] NAEA-PreferredCI               OPTIONAL,
		-- naea-PreferredCI is included at the discretion of the HLR operator.
		gprsSubscriptionData           [16] GPRSSubscriptionData           OPTIONAL,
		roamingRestrictedInSgsnDueToUnsupportedFeature [23] NULL           OPTIONAL,
		networkAccessMode	           [24] NetworkAccessMode              OPTIONAL,
		lsaInformation                 [25] LSAInformation                 OPTIONAL,
		lmu-Indicator                  [21] NULL                           OPTIONAL,
		lcsInformation                 [22] LCSInformation                 OPTIONAL,
		istAlertTimer                  [26] IST-AlertTimerValue            OPTIONAL,
		superChargerSupportedInHLR     [27] AgeIndicator                   OPTIONAL,
		mc-SS-Info                     [28] MC-SS-Info                     OPTIONAL,
		cs-AllocationRetentionPriority [29] CS-AllocationRetentionPriority OPTIONAL,
		sgsn-CAMEL-SubscriptionInfo    [17] SGSN-CAMEL-SubscriptionInfo    OPTIONAL,
		chargingCharacteristics        [18]	ChargingCharacteristics        OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		accessRestrictionData                 [19] AccessRestrictionData          OPTIONAL,
		ics-Indicator                         [20] BOOLEAN                        OPTIONAL,
		eps-SubscriptionData                  [31] EPS-SubscriptionData           OPTIONAL,
		csg-SubscriptionDataList              [32] CSG-SubscriptionDataList       OPTIONAL,
		ue-ReachabilityRequestIndicator       [33] NULL                           OPTIONAL,
		sgsn-Number                           [34] ISDN-AddressString             OPTIONAL,
		mme-Name                              [35] DiameterIdentity               OPTIONAL,
		subscribedPeriodicRAUTAUtimer         [36] SubscribedPeriodicRAUTAUtimer  OPTIONAL,
		vplmnLIPAAllowed                      [37] NULL                           OPTIONAL,
		mdtUserConsent                        [38] BOOLEAN                        OPTIONAL,
		subscribedPeriodicLAUtimer            [39] SubscribedPeriodicLAUtimer     OPTIONAL,
		vplmn-Csg-SubscriptionDataList        [40] VPLMN-CSG-SubscriptionDataList OPTIONAL,
		additionalMSISDN                      [41] ISDN-AddressString             OPTIONAL,
		psAndSMS-OnlyServiceProvision         [42] NULL                           OPTIONAL,
		smsInSGSNAllowed                      [43] NULL                           OPTIONAL,
		cs-to-ps-SRVCC-Allowed-Indicator      [44] NULL                           OPTIONAL,
		pcscf-Restoration-Request             [45] NULL                           OPTIONAL,
		adjacentAccessRestrictionDataList     [46] AdjacentAccessRestrictionDataList OPTIONAL,
		imsi-Group-Id-List                    [47] IMSI-GroupIdList               OPTIONAL,
		ueUsageType                           [48] UE-UsageType                   OPTIONAL,
		userPlaneIntegrityProtectionIndicator [49] NULL                           OPTIONAL,
		dl-Buffering-Suggested-Packet-Count   [50] DL-Buffering-Suggested-Packet-Count OPTIONAL,
		reset-Id-List                         [51] Reset-Id-List                  OPTIONAL,
		eDRX-Cycle-Length-List                [52] EDRX-Cycle-Length-List         OPTIONAL,
		ext-AccessRestrictionData             [53] Ext-AccessRestrictionData      OPTIONAL,
		iab-Operation-Allowed-Indicator       [54] NULL                           OPTIONAL }
		-- If the Network Access Mode parameter is sent, it shall be present only in
		-- the first sequence if seqmentation is used

	Category ::= OCTET STRING (SIZE (1))
		-- The internal structure is defined in ITU-T Rec Q.763.
	AgeIndicator ::= OCTET STRING (SIZE (1..6))
		-- The internal structure of this parameter is implementation specific.
	IST-AlertTimerValue ::= INTEGER (15..255)
	CS-AllocationRetentionPriority ::= OCTET STRING (SIZE (1))
		-- This data type encodes each priority level defined in 3GPP TS 23.107 [154] as the
		--binary value of the priority level.
*/
type InsertSubscriberDataArg struct {
	InvokeID int8 `json:"id"`

	IMSI                     teldata.IMSI         `json:"imsi,omitempty"`
	MSISDN                   common.AddressString `json:"msisdn,omitempty"`
	Category                 byte                 `json:"category,omitempty"`
	SubscriberStatus         subscriberStatus     `json:"subscriberStatus,omitempty"`
	BsList                   []uint8              `json:"bearerServiceList,omitempty"`
	TsList                   []uint8              `json:"teleserviceList,omitempty"`
	ProvisionedSS            ssInfoList           `json:"provisionedSS,omitempty"`
	OdbData                  odbData              `json:"odb-Data,omitempty"`
	RoamingRestriction       bool                 `json:"roamingRestrictionDueToUnsupportedFeature,omitempty"`
	RegionalSubscriptionData []data16             `json:"regionalSubscriptionData,omitempty"`
	VbsSubscriptionData      []vBroadcastData     `json:"vbsSubscriptionData,omitempty"`
	VgcsSubscriptionData     []vGroupCallData     `json:"vgcsSubscriptionData,omitempty"`
	// VlrCamelSubscriptionInfo VlrCamelSubsInfo     `json:"vlrCamelSubscriptionInfo,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	NaeaPreferredCI *NAEAPreferredCI `json:"naea-PreferredCI,omitempty"`
	// GprsSubscriptionData          GPRSSubscriptionData          `json:"gprsSubscriptionData,omitempty"`
	RoamingRestrictedInSgsn bool         `json:"roamingRestrictedInSgsnDueToUnsupportedFeature,omitempty"`
	AccessMode              nwAccessMode `json:"networkAccessMode,omitempty"`
	// LsaInformation                LSAInformation                `json:"lsaInformation,omitempty"`
	LmuIndicator bool `json:"lmu-Indicator,omitempty"`
	// LcsInformation                LCSInformation                `json:"lcsInformation,omitempty"`
	IstAlertTimer uint8 `json:"istAlertTimer,omitempty"`
	// SuperChargerSupportedInHLR    AgeIndicator                  `json:"superChargerSupportedInHLR,omitempty"`
	// McSSInfo                      McSSInfo                      `json:"mc-SS-Info,omitempty"`
	CsAllocationRetentionPriority byte `json:"cs-AllocationRetentionPriority,omitempty"`
	// SgsnCAMELSubscriptionInfo     SGSNCAMELSubscriptionInfo `json:"sgsn-CAMEL-SubscriptionInfo,omitempty"`
	ChargingCharacteristics data16 `json:"chargingCharacteristics,omitempty"`
}

func (isd InsertSubscriberDataArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", isd.Name(), isd.InvokeID)

	if !isd.IMSI.IsEmpty() {
		fmt.Fprintf(buf, "\n%simsi: %s", gsmap.LogPrefix, isd.IMSI)
	}
	if !isd.MSISDN.IsEmpty() {
		fmt.Fprintf(buf, "\n%smsisdn: %s", gsmap.LogPrefix, isd.MSISDN)
	}
	if isd.Category != 0 {
		fmt.Fprintf(buf, "\n%scategory: %x", gsmap.LogPrefix, isd.Category)
	}
	if isd.SubscriberStatus != 0 {
		fmt.Fprintf(buf, "\n%ssubscriberStatus: %s", gsmap.LogPrefix, isd.SubscriberStatus)
	}
	if len(isd.BsList) != 0 {
		fmt.Fprintf(buf, "\n%sbearerServiceList: %v", gsmap.LogPrefix, isd.BsList)
	}
	if len(isd.TsList) != 0 {
		fmt.Fprintf(buf, "\n%steleserviceList: %v", gsmap.LogPrefix, isd.TsList)
	}
	if len(isd.ProvisionedSS) != 0 {
		fmt.Fprintf(buf, "\n%sprovisionedSS: %v", gsmap.LogPrefix, isd.ProvisionedSS)
	}
	if isd.OdbData.GeneralData != 0 {
		fmt.Fprintf(buf, "\n%sodb-Data: %v", gsmap.LogPrefix, isd.OdbData)
	}
	if isd.RoamingRestriction {
		fmt.Fprintf(buf, "\n%sroamingRestrictionDueToUnsupportedFeature:", gsmap.LogPrefix)
	}
	if len(isd.RegionalSubscriptionData) != 0 {
		fmt.Fprintf(buf, "\n%sregionalSubscriptionData: %v", gsmap.LogPrefix, isd.RegionalSubscriptionData)
	}
	if len(isd.VbsSubscriptionData) != 0 {
		fmt.Fprintf(buf, "\n%svbsSubscriptionData:", gsmap.LogPrefix)
		for _, d := range isd.VbsSubscriptionData {
			fmt.Fprintf(buf, "\n%s| %s", gsmap.LogPrefix, d)
		}
	}
	if len(isd.VgcsSubscriptionData) != 0 {
		fmt.Fprintf(buf, "\n%svgcsSubscriptionData:", gsmap.LogPrefix)
		for _, d := range isd.VgcsSubscriptionData {
			fmt.Fprintf(buf, "\n%s| %s", gsmap.LogPrefix, d)
		}
	}
	// VlrCamelSubscriptionInfo
	if isd.NaeaPreferredCI != nil {
		fmt.Fprintf(buf, "\n%snaea-PreferredCI: %s", gsmap.LogPrefix, isd.NaeaPreferredCI)
	}
	// GprsSubscriptionData
	if isd.RoamingRestrictedInSgsn {
		fmt.Fprintf(buf, "\n%sroamingRestrictedInSgsnDueToUnsupportedFeature:",
			gsmap.LogPrefix)
	}
	if isd.AccessMode != 0 {
		fmt.Fprintf(buf, "\n%snetworkAccessMode: %s", gsmap.LogPrefix, isd.AccessMode)
	}
	// LsaInformation
	if isd.LmuIndicator {
		fmt.Fprintf(buf, "\n%slmu-Indicator:", gsmap.LogPrefix)
	}
	// LcsInformation
	if isd.IstAlertTimer >= 15 {
		fmt.Fprintf(buf, "\n%sistAlertTimer: %d", gsmap.LogPrefix, isd.IstAlertTimer)
	}
	// SuperChargerSupportedInHLR
	// McSSInfo
	if isd.CsAllocationRetentionPriority != 0 {
		fmt.Fprintf(buf, "\n%scs-AllocationRetentionPriority: %d",
			gsmap.LogPrefix, isd.CsAllocationRetentionPriority)
	}
	// SgsnCAMELSubscriptionInfo
	if !isd.ChargingCharacteristics.IsEmpty() {
		fmt.Fprintf(buf, "\n%schargingCharacteristics: %s",
			gsmap.LogPrefix, isd.ChargingCharacteristics)
	}
	return buf.String()
}

func (isd InsertSubscriberDataArg) GetInvokeID() int8            { return isd.InvokeID }
func (InsertSubscriberDataArg) GetLinkedID() *int8               { return nil }
func (InsertSubscriberDataArg) Code() byte                       { return 7 }
func (InsertSubscriberDataArg) Name() string                     { return "InsertSubscriberData-Arg" }
func (InsertSubscriberDataArg) DefaultContext() gsmap.AppContext { return SubscriberDataMngt1 }

func (InsertSubscriberDataArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		InsertSubscriberDataArg
	}{}
	e := json.Unmarshal(v, &tmp)
	if e != nil {
		return tmp.InsertSubscriberDataArg, e
	}
	c := tmp.InsertSubscriberDataArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, e
}

func (isd InsertSubscriberDataArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// imsi, context_specific(80) + primitive(00) + 0(00)
	if !isd.IMSI.IsEmpty() {
		gsmap.WriteTLV(buf, 0x80, isd.IMSI.Bytes())
	}

	// msisdn, context_specific(80) + primitive(00) + 1(01)
	if !isd.MSISDN.IsEmpty() {
		gsmap.WriteTLV(buf, 0x81, isd.MSISDN.Bytes())
	}

	// category, context_specific(80) + primitive(00) + 2(02)
	if isd.Category != 0 {
		gsmap.WriteTLV(buf, 0x82, []byte{isd.Category})
	}

	// subscriberStatus, context_specific(80) + primitive(00) + 3(03)
	if d := isd.SubscriberStatus.marshal(); d != nil {
		gsmap.WriteTLV(buf, 0x83, d)
	}

	// bearerServiceList, context_specific(80) + constructed(20) + 4(04)
	if len(isd.BsList) != 0 {
		gsmap.WriteTLV(buf, 0xa4, marshalCodeList(isd.BsList))
	}

	// teleserviceList, context_specific(80) + constructed(20) + 6(06)
	if len(isd.TsList) != 0 {
		gsmap.WriteTLV(buf, 0xa6, marshalCodeList(isd.TsList))
	}

	// provisionedSS, context_specific(80) + constructed(20) + 7(07)
	if len(isd.ProvisionedSS) != 0 {
		gsmap.WriteTLV(buf, 0xa7, isd.ProvisionedSS.marshal())
	}

	// odb-Data, context_specific(80) + constructed(20) + 8(08)
	if isd.OdbData.GeneralData != 0 {
		gsmap.WriteTLV(buf, 0xa8, isd.OdbData.marshal())
	}

	// roamingRestrictionDueToUnsupportedFeature, context_specific(80) + primitive(00) + 9(09)
	if isd.RoamingRestriction {
		gsmap.WriteTLV(buf, 0x89, nil)
	}

	// regionalSubscriptionData, context_specific(80) + constructed(20) + 10(0a)
	if len(isd.RegionalSubscriptionData) != 0 {
		gsmap.WriteTLV(buf, 0xaa, marshalData16List(isd.RegionalSubscriptionData))
	}

	// vbsSubscriptionData, context_specific(80) + constructed(20) + 11(0b)
	if len(isd.VbsSubscriptionData) != 0 {
		buf2 := new(bytes.Buffer)
		for _, gr := range isd.VbsSubscriptionData {
			// VoiceBroadcastData, universal(00) + constructed(20) + sequence(10)
			gsmap.WriteTLV(buf2, 0x30, gr.marshal())
		}
		gsmap.WriteTLV(buf, 0xab, buf2.Bytes())
	}

	// vgcsSubscriptionData, context_specific(80) + constructed(20) + 12(0c)
	if len(isd.VgcsSubscriptionData) != 0 {
		buf2 := new(bytes.Buffer)
		for _, gr := range isd.VgcsSubscriptionData {
			// VoiceGroupCallData, universal(00) + constructed(20) + sequence(10)
			gsmap.WriteTLV(buf2, 0x30, gr.marshal())
		}
		gsmap.WriteTLV(buf, 0xac, buf2.Bytes())
	}

	// vlrCamelSubscriptionInfo, context_specific(80) + constructed(20) + 13(0d)
	// extensionContainer, context_specific(80) + constructed(20) + 14(0e)

	// naea-PreferredCI, context_specific(80) + constructed(20) + 15(0f)
	if isd.NaeaPreferredCI != nil {
		gsmap.WriteTLV(buf, 0xaf, isd.NaeaPreferredCI.marshal())
	}

	// gprsSubscriptionData, context_specific(80) + constructed(20) + 16(10)

	// roamingRestrictedInSgsnDueToUnsupportedFeature, context_specific(80) + primitive(00) + 23(17)
	if isd.RoamingRestrictedInSgsn {
		gsmap.WriteTLV(buf, 0x97, nil)
	}

	// networkAccessMode, context_specific(80) + primitive(00) + 24(18)
	if d := isd.AccessMode.marshal(); d != nil {
		gsmap.WriteTLV(buf, 0x98, d)
	}

	// lsaInformation, context_specific(80) + constructed(20) + 25(19)

	// lmu-Indicator, context_specific(80) + primitive(00) + 21(15)
	if isd.LmuIndicator {
		gsmap.WriteTLV(buf, 0x95, nil)
	}

	// lcsInformation, context_specific(80) + constructed(20) + 22(16)

	// istAlertTimer, context_specific(80) + primitive(00) + 26(1a)
	if isd.IstAlertTimer >= 15 {
		gsmap.WriteTLV(buf, 0x9a, []byte{isd.IstAlertTimer})
	}

	// superChargerSupportedInHLR, context_specific(80) + primitive(00) + 27(1b)
	// mc-SS-Info, context_specific(80) + constructed(20) + 28(1c)

	// cs-AllocationRetentionPriority, context_specific(80) + primitive(00) + 29(1d)
	if isd.CsAllocationRetentionPriority != 0 {
		gsmap.WriteTLV(buf, 0x9d, []byte{isd.CsAllocationRetentionPriority})
	}

	// sgsn-CAMEL-SubscriptionInfo, context_specific(80) + constructed(20) + 17(11)

	// ChargingCharacteristics, context_specific(80) + primitive(00) + 18(12)
	if !isd.ChargingCharacteristics.IsEmpty() {
		gsmap.WriteTLV(buf, 0x92, isd.ChargingCharacteristics.marshal())
	}

	// InsertSubscriberData-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (InsertSubscriberDataArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// InsertSubscriberData-Arg, universal(00) + constructed(20) + sequence(10)
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}
	isd := InsertSubscriberDataArg{InvokeID: id}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return isd, nil
	} else if e != nil {
		return nil, e
	}

	// imsi, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if isd.IMSI, e = teldata.DecodeIMSI(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// msisdn, context_specific(80) + primitive(00) + 1(01)
	if t == 0x81 {
		if isd.MSISDN, e = common.DecodeAddressString(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// category, context_specific(80) + primitive(00) + 2(02)
	if t == 0x82 {
		isd.Category = v[0]

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// subscriberStatus, context_specific(80) + primitive(00) + 3(03)
	if t == 0x83 {
		if e = isd.SubscriberStatus.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// bearerServiceList, context_specific(80) + constructed(20) + 4(04)
	if t == 0xa4 {
		if isd.BsList, e = unmarshalCodeList(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// teleserviceList, context_specific(80) + constructed(20) + 6(06)
	if t == 0xa6 {
		if isd.TsList, e = unmarshalCodeList(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// provisionedSS, context_specific(80) + constructed(20) + 7(07)
	if t == 0xa7 {
		if isd.ProvisionedSS, e = isd.ProvisionedSS.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// odb-Data, context_specific(80) + constructed(20) + 8(08)
	if t == 0xa8 {
		if e = isd.OdbData.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// roamingRestrictionDueToUnsupportedFeature, context_specific(80) + primitive(00) + 9(09)
	if t == 0x89 {
		isd.RoamingRestriction = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// regionalSubscriptionData, context_specific(80) + constructed(20) + 10(0a)
	if t == 0xaa {
		if isd.RegionalSubscriptionData, e = unmarshalData16List(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// vbsSubscriptionData, context_specific(80) + constructed(20) + 11(0b)
	if t == 0xab {
		buf2 := bytes.NewBuffer(v)
		isd.VbsSubscriptionData = []vBroadcastData{}
		for {
			d := vBroadcastData{}
			if _, v, e = gsmap.ReadTLV(buf2, 0x30); e == io.EOF {
				break
			} else if e != nil {
				return nil, e
			} else if e = d.unmarshal(v); e != nil {
				return nil, e
			} else {
				isd.VbsSubscriptionData = append(isd.VbsSubscriptionData, d)
			}
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// vgcsSubscriptionData, context_specific(80) + constructed(20) + 12(0c)
	if t == 0xac {
		buf2 := bytes.NewBuffer(v)
		isd.VgcsSubscriptionData = []vGroupCallData{}
		for {
			d := vGroupCallData{}
			if _, v, e = gsmap.ReadTLV(buf2, 0x30); e == io.EOF {
				break
			} else if e != nil {
				return nil, e
			} else if e = d.unmarshal(v); e != nil {
				return nil, e
			} else {
				isd.VgcsSubscriptionData = append(isd.VgcsSubscriptionData, d)
			}
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// vlrCamelSubscriptionInfo, context_specific(80) + constructed(20) + 13(0d)
	if t == 0xad {
		// unmarshal data

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, context_specific(80) + constructed(20) + 14(0e)
	if t == 0xae {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// naea-PreferredCI, context_specific(80) + constructed(20) + 15(0f)
	if t == 0xaf {
		isd.NaeaPreferredCI = &NAEAPreferredCI{}
		if e = isd.NaeaPreferredCI.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// gprsSubscriptionData, context_specific(80) + constructed(20) + 16(10)
	if t == 0xb0 {
		// unmarshal data

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// roamingRestrictedInSgsnDueToUnsupportedFeature, context_specific(80) + primitive(00) + 23(17)
	if t == 0x97 {
		isd.RoamingRestrictedInSgsn = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// networkAccessMode, context_specific(80) + primitive(00) + 24(18)
	if t == 0x98 {
		if e = isd.AccessMode.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// lsaInformation, context_specific(80) + constructed(20) + 25(19)
	if t == 0xb9 {
		// unmarshal data

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// lmu-Indicator, context_specific(80) + primitive(00) + 21(15)
	if t == 0x95 {
		isd.LmuIndicator = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// lcsInformation, context_specific(80) + constructed(20) + 22(16)
	if t == 0xb6 {
		// unmarshal data

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// istAlertTimer, context_specific(80) + primitive(00) + 26(1a)
	if t == 0x9a {
		isd.IstAlertTimer = v[0]

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// superChargerSupportedInHLR, context_specific(80) + primitive(00) + 27(1b)
	if t == 0x9b {
		// unmarshal data

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// mc-SS-Info, context_specific(80) + constructed(20) + 28(1c)
	if t == 0xbc {
		// unmarshal data

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// cs-AllocationRetentionPriority, context_specific(80) + primitive(00) + 29(1d)
	if t == 0x9d {
		isd.CsAllocationRetentionPriority = v[0]

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// sgsn-CAMEL-SubscriptionInfo, context_specific(80) + constructed(20) + 17(11)
	if t == 0xb1 {
		// unmarshal data

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// ChargingCharacteristics, context_specific(80) + primitive(00) + 18(12)
	if t == 0x92 {
		if e = isd.ChargingCharacteristics.unmarshal(v); e != nil {
			return nil, e
		}

		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return isd, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return isd, nil
}

/*
InsertSubscriberDataRes operation res.

	InsertSubscriberDataRes ::= SEQUENCE {
		teleserviceList              [1]  TeleserviceList              OPTIONAL,
		bearerServiceList            [2]  BearerServiceList            OPTIONAL,
		ss-List                      [3]  SS-List                      OPTIONAL,
		odb-GeneralData              [4]  ODB-GeneralData              OPTIONAL,
		regionalSubscriptionResponse [5]  RegionalSubscriptionResponse OPTIONAL,
		supportedCamelPhases         [6]  SupportedCamelPhases         OPTIONAL,
		extensionContainer           [7]  ExtensionContainer           OPTIONAL,
		... ,
		-- ^^^^^^ R99 ^^^^^^
		offeredCamel4CSIs            [8]  OfferedCamel4CSIs     OPTIONAL,
		supportedFeatures            [9]  SupportedFeatures     OPTIONAL,
		ext-SupportedFeatures        [10] Ext-SupportedFeatures OPTIONAL }
*/
type InsertSubscriberDataRes struct {
	InvokeID int8 `json:"id"`

	TsList             []uint8                 `json:"teleserviceList,omitempty"`
	BsList             []uint8                 `json:"bearerServiceList,omitempty"`
	SsList             []uint8                 `json:"ss-List,omitempty"`
	OdbGeneralData     odbGeneralData          `json:"odb-GeneralData,omitempty"`
	RegionSubscription regionalSubscriptionRes `json:"regionalSubscriptionResponse,omitempty"`
	SupportedCamelPh   supportedCamelPh        `json:"supportedCamelPhases,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (isd InsertSubscriberDataRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", isd.Name(), isd.InvokeID)

	if len(isd.TsList) != 0 {
		fmt.Fprintf(buf, "\n%steleserviceList:   %v", gsmap.LogPrefix, isd.TsList)
	}
	if len(isd.BsList) != 0 {
		fmt.Fprintf(buf, "\n%sbearerServiceList: %v", gsmap.LogPrefix, isd.BsList)
	}
	if len(isd.SsList) != 0 {
		fmt.Fprintf(buf, "\n%sss-List:           %v", gsmap.LogPrefix, isd.SsList)
	}
	if isd.OdbGeneralData != 0 {
		fmt.Fprintf(buf, "\n%sodb-GeneralData: %x", gsmap.LogPrefix, isd.OdbGeneralData)
	}
	if isd.RegionSubscription != 0 {
		fmt.Fprintf(buf, "\n%sregionalSubscriptionResponse: %s",
			gsmap.LogPrefix, isd.RegionSubscription)
	}
	if isd.SupportedCamelPh != 0 {
		fmt.Fprintf(buf, "\n%ssupportedCamelPhases: %s",
			gsmap.LogPrefix, isd.SupportedCamelPh)
	}
	// Extension
	return buf.String()
}

func (isd InsertSubscriberDataRes) GetInvokeID() int8 { return isd.InvokeID }
func (InsertSubscriberDataRes) Code() byte            { return 7 }
func (InsertSubscriberDataRes) Name() string          { return "InsertSubscriberData-Res" }

func (InsertSubscriberDataRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		InsertSubscriberDataRes
	}{}
	e := json.Unmarshal(v, &tmp)
	if e != nil {
		return tmp.InsertSubscriberDataRes, e
	}
	c := tmp.InsertSubscriberDataRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, e
}

func (isd InsertSubscriberDataRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)
	// teleserviceList, context_specific(80) + constructed(20) + 1(01)
	if len(isd.TsList) != 0 {
		gsmap.WriteTLV(buf, 0xa1, marshalCodeList(isd.TsList))
	}

	// bearerServiceList, context_specific(80) + constructed(20) + 2(02)
	if len(isd.BsList) != 0 {
		gsmap.WriteTLV(buf, 0xa2, marshalCodeList(isd.BsList))
	}

	// ss-List, context_specific(80) + constructed(20) + 3(03)
	if len(isd.SsList) != 0 {
		gsmap.WriteTLV(buf, 0xa3, marshalCodeList(isd.SsList))
	}

	// odb-GeneralData, context_specific(80) + primitive(00) + 4(04)
	if isd.OdbGeneralData != 0 {
		gsmap.WriteTLV(buf, 0x84, isd.OdbGeneralData.marshal())
	}

	// regionalSubscriptionResponse, context_specific(80) + primitive(00) + 5(05)
	if b := isd.RegionSubscription.marshal(); b != nil {
		gsmap.WriteTLV(buf, 0x85, b)
	}

	// supportedCamelPhases, context_specific(80) + primitive(00) + 6(06)
	if isd.SupportedCamelPh != 0 {
		gsmap.WriteTLV(buf, 0x86, isd.SupportedCamelPh.marshal())
	}

	// extensionContainer, context_specific(80) + constructed(20) + 7(07)

	if buf.Len() != 0 {
		// InsertSubscriberData-Res, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (InsertSubscriberDataRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// InsertSubscriberData-Res, universal(00) + constructed(20) + sequence(10)
	isd := InsertSubscriberDataRes{InvokeID: id}
	if buf.Len() == 0 {
		return isd, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return isd, nil
	} else if e != nil {
		return nil, e
	}

	// teleserviceList, context_specific(80) + constructed(20) + 1(01)
	if t == 0xa1 {
		if isd.TsList, e = unmarshalCodeList(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// bearerServiceList, context_specific(80) + constructed(20) + 2(02)
	if t == 0xa2 {
		if isd.BsList, e = unmarshalCodeList(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// ss-List, context_specific(80) + constructed(20) + 3(03)
	if t == 0xa3 {
		if isd.SsList, e = unmarshalCodeList(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// odb-GeneralData, context_specific(80) + primitive(00) + 4(04)
	if t == 0x84 {
		if e = isd.OdbGeneralData.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// regionalSubscriptionResponse, context_specific(80) + primitive(00) + 5(05)
	if t == 0x85 {
		if e = isd.RegionSubscription.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// supportedCamelPhases, context_specific(80) + primitive(00) + 6(06)
	if t == 0x86 {
		if e = isd.SupportedCamelPh.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isd, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, context_specific(80) + constructed(20) + 7(07)
	if t == 0xa7 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}
		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return isd, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return isd, nil
}

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

/*
DeleteSubscriberDataArg

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
*/

/*
DeleteSubscriberDataRes

	DeleteSubscriberDataRes ::= SEQUENCE {
		regionalSubscriptionResponse [0] RegionalSubscriptionResponse OPTIONAL,
		extensionContainer               ExtensionContainer           OPTIONAL,
		...}
*/
