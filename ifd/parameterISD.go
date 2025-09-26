package ifd

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
	TeleserviceList ::= SEQUENCE SIZE (1..maxNumOfTeleservices) OF
		Ext-TeleserviceCode
	Ext-TeleserviceCode ::= OCTET STRING (SIZE (1..5))
		-- OCTET 1:
			-- bits 87654321: group (bits 8765) and specific service (bits 4321)
		-- OCTETS 2-5: reserved for future use.

	BearerServiceList ::= SEQUENCE SIZE (1..maxNumOfBearerServices) OF
		Ext-BearerServiceCode
	Ext-BearerServiceCode ::= OCTET STRING (SIZE (1..5))
		-- OCTET 1:
			-- plmn-specific bearer services:
				-- bits 87654321: defined by the HPLMN operator
			-- rest of bearer services:
				-- bit 8: 0 (unused)
				-- bits 7654321: group (bits 7654), and rate, if applicable (bits 321)
		-- OCTETS 2-5: reserved for future use.

	SS-List ::= SEQUENCE SIZE (1..maxNumOfSS) OF SS-Code
	SS-Code ::= OCTET STRING (SIZE (1))
		-- bits 87654321: group (bits 8765), and specific service (bits 4321)
*/

func marshalCodeList(l []uint8) []byte {
	buf := new(bytes.Buffer)
	for _, c := range l {
		// universal(00) + primitive(00) + octet_string(04)
		gsmap.WriteTLV(buf, 0x04, []byte{c})
	}
	return buf.Bytes()
}

func unmarshalCodeList(b []byte) ([]uint8, error) {
	l := []uint8{}
	buf := bytes.NewBuffer(b)
	for {
		// universal(00) + primitive(00) + octet_string(04)
		if _, v, e := gsmap.ReadTLV(buf, 0x04); e == io.EOF {
			break
		} else if e != nil {
			return nil, e
		} else {
			l = append(l, v[0])
		}
	}
	return l, nil
}

/*
subscriberStatus

	SubscriberStatus ::= ENUMERATED {
		serviceGranted            (0),
		operatorDeterminedBarring (1) }
*/
type subscriberStatus byte

const (
	_ subscriberStatus = iota
	ServiceGranted
	OperatorDeterminedBarring
)

func (s subscriberStatus) String() string {
	switch s {
	case ServiceGranted:
		return "serviceGranted"
	case OperatorDeterminedBarring:
		return "operatorDeterminedBarring"
	}
	return ""
}

func (s *subscriberStatus) UnmarshalJSON(b []byte) (e error) {
	var d string
	e = json.Unmarshal(b, &d)
	switch d {
	case "serviceGranted":
		*s = ServiceGranted
	case "operatorDeterminedBarring":
		*s = OperatorDeterminedBarring
	default:
		*s = 0
	}
	return
}

func (s subscriberStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *subscriberStatus) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*s = ServiceGranted
	case 1:
		*s = OperatorDeterminedBarring
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (s subscriberStatus) marshal() []byte {
	switch s {
	case ServiceGranted:
		return []byte{0x00}
	case OperatorDeterminedBarring:
		return []byte{0x01}
	}
	return nil
}

/*
odbData

	ODB-Data ::= SEQUENCE {
		odb-GeneralData	ODB-GeneralData,
		odb-HPLMN-Data	ODB-HPLMN-Data	OPTIONAL,
		extensionContainer	ExtensionContainer	OPTIONAL,
		...}
*/
type odbData struct {
	GeneralData odbGeneralData `json:"odb-GeneralData"`
	HplmnData   odbHplmnData   `json:"odb-HPLMN-Data,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (o odbData) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s| odb-GeneralData: %s", gsmap.LogPrefix, o.GeneralData)
	if o.HplmnData != 0 {
		fmt.Fprintf(buf, "\n%s| odb-HPLMN-Data: %s", gsmap.LogPrefix, o.HplmnData)
	}
	return buf.String()
}

func (o odbData) marshal() []byte {
	buf := new(bytes.Buffer)

	// odb-GeneralData, universal(00) + primitive(00) + bit_string(03)
	gsmap.WriteTLV(buf, 0x03, o.GeneralData.marshal())

	// odb-HPLMN-Data, universal(00) + primitive(00) + bit_string(03)
	if o.HplmnData != 0 {
		gsmap.WriteTLV(buf, 0x03, o.HplmnData.marshal())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return buf.Bytes()
}

func (o *odbData) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	// odb-GeneralData, universal(00) + primitive(00) + bit_string(03)
	if _, v, e := gsmap.ReadTLV(buf, 0x03); e != nil {
		return e
	} else if e = o.GeneralData.unmarshal(v); e != nil {
		return e
	}

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	// odb-HPLMN-Data, universal(00) + primitive(00) + bit_string(03)
	if t == 0x03 {
		if e = o.HplmnData.unmarshal(v); e != nil {
			return e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return e
		}
		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return rsm, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return nil
}

/*
odbGeneralData

	ODB-GeneralData ::= BIT STRING {
		allOG-CallsBarred  (0),
		internationalOGCallsBarred  (1),
		internationalOGCallsNotToHPLMN-CountryBarred  (2),
		premiumRateInformationOGCallsBarred  (3),
		premiumRateEntertainementOGCallsBarred  (4),
		ss-AccessBarred  (5),
		interzonalOGCallsBarred (6),
		interzonalOGCallsNotToHPLMN-CountryBarred (7),
		interzonalOGCallsAndInternationalOGCallsNotToHPLMN-CountryBarred (8),
		allECT-Barred (9),
		chargeableECT-Barred (10),
		internationalECT-Barred (11),
		interzonalECT-Barred (12),
		doublyChargeableECT-Barred (13),
		multipleECT-Barred (14),
		-- ^^^^^^ R99 ^^^^^^
		allPacketOrientedServicesBarred (15),
		roamerAccessToHPLMN-AP-Barred  (16),
		roamerAccessToVPLMN-AP-Barred  (17),
		roamingOutsidePLMNOG-CallsBarred  (18),
		allIC-CallsBarred  (19),
		roamingOutsidePLMNIC-CallsBarred  (20),
		roamingOutsidePLMNICountryIC-CallsBarred  (21),
		roamingOutsidePLMN-Barred  (22),
		roamingOutsidePLMN-CountryBarred  (23),
		registrationAllCF-Barred  (24),
		registrationCFNotToHPLMN-Barred  (25),
		registrationInterzonalCF-Barred  (26),
		registrationInterzonalCFNotToHPLMN-Barred  (27),
		registrationInternationalCF-Barred  (28)} (SIZE (15..32))

	-- exception handling: reception of unknown bit assignments in the
	-- ODB-GeneralData type shall be treated like unsupported ODB-GeneralData
	-- When the ODB-GeneralData type is removed from the HLR for a given subscriber,
	-- in NoteSubscriberDataModified operation sent toward the gsmSCF
	-- all bits shall be set to "O".
*/
type odbGeneralData uint16

func (o odbGeneralData) String() string {
	return fmt.Sprintf("%016b", o)
}

func (o *odbGeneralData) UnmarshalJSON(b []byte) (e error) {
	var s string
	if e = json.Unmarshal(b, &s); e != nil {
		return
	}
	tmp, e := strconv.ParseUint(s, 2, 16)
	if e == nil {
		*o = odbGeneralData(tmp)
	}
	return
}

func (o odbGeneralData) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(o), 2))
}

func (o *odbGeneralData) unmarshal(b []byte) error {
	if len(b) < 3 {
		return gsmap.UnexpectedTLV("length must <3")
	}
	*o = (odbGeneralData(b[1]) << 8) | odbGeneralData(b[2]&0xfe)
	return nil
}

func (o odbGeneralData) marshal() []byte {
	return []byte{0x01, byte(o >> 8), byte(o)}
}

/*
odbHplmnData

	ODB-HPLMN-Data ::= BIT STRING {
		plmn-SpecificBarringType1  (0),
		plmn-SpecificBarringType2  (1),
		plmn-SpecificBarringType3  (2),
		plmn-SpecificBarringType4  (3)} (SIZE (4..32))

	-- exception handling: reception of unknown bit assignments in the
	-- ODB-HPLMN-Data type shall be treated like unsupported ODB-HPLMN-Data
	-- When the ODB-HPLMN-Data type is removed from the HLR for a given subscriber,
	-- in NoteSubscriberDataModified operation sent toward the gsmSCF
	-- all bits shall be set to "O".
*/
type odbHplmnData uint8

func (o odbHplmnData) String() string {
	return fmt.Sprintf("%08b", o)
}

func (o *odbHplmnData) UnmarshalJSON(b []byte) (e error) {
	var s string
	if e = json.Unmarshal(b, &s); e != nil {
		return
	}
	tmp, e := strconv.ParseUint(s, 2, 8)
	if e == nil {
		*o = odbHplmnData(tmp)
	}
	return
}

func (o odbHplmnData) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(o), 2))
}

func (o *odbHplmnData) unmarshal(b []byte) error {
	if len(b) < 2 {
		return gsmap.UnexpectedTLV("length must <2")
	}
	*o = odbHplmnData(b[1] & 0xf0)
	return nil
}

func (o odbHplmnData) marshal() []byte {
	return []byte{0x01, byte(o)}
}

/*
data16 is 2 octet string data.

	ZoneCodeList ::= SEQUENCE SIZE (1..maxNumOfZoneCodes)
		OF ZoneCode
	ZoneCode ::= OCTET STRING (SIZE (2))
		-- internal structure is defined in TS 3GPP TS 23.003 [17]

	ChargingCharacteristics ::= OCTET STRING (SIZE (2))
		-- Octets are coded according to 3GPP TS 32.215.
*/
type data16 [2]byte

func (d data16) String() string {
	return hex.EncodeToString(d[:])
}

func (d *data16) UnmarshalJSON(b []byte) error {
	var s string
	if e := json.Unmarshal(b, &s); e != nil {
		return e
	}
	if tmp, e := hex.DecodeString(s); e != nil {
		return e
	} else if len(tmp) != 2 {
		return fmt.Errorf("data length is not 2 octets")
	} else {
		d[0] = tmp[0]
		d[1] = tmp[1]
	}
	return nil
}

func (d data16) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d data16) IsEmpty() bool {
	return d[0] == 0 && d[1] == 0
}

func (d data16) marshal() []byte {
	return d[:]
}

func (d *data16) unmarshal(b []byte) error {
	if len(b) != 2 {
		return gsmap.UnexpectedTLV("length must 2")
	}
	d[0] = b[0]
	d[1] = b[1]
	return nil
}

func marshalData16List(l []data16) []byte {
	buf := new(bytes.Buffer)
	for _, d := range l {
		// universal(00) + primitive(00) + octet_string(04)
		gsmap.WriteTLV(buf, 0x04, d[:])
	}
	return buf.Bytes()
}

func unmarshalData16List(b []byte) ([]data16, error) {
	l := []data16{}
	buf := bytes.NewBuffer(b)
	for {
		// universal(00) + primitive(00) + octet_string(04)
		if _, v, e := gsmap.ReadTLV(buf, 0x04); e == io.EOF {
			break
		} else if e != nil {
			return nil, e
		} else if len(v) != 2 {
			return nil, gsmap.UnexpectedTLV("length must 2")
		} else {
			l = append(l, data16{v[0], v[1]})
		}
	}
	return l, nil
}

/*
gID

	GroupId  ::= TBCD-STRING (SIZE (3))
	-- When Group-Id is less than six characters in length, the TBCD filler (1111)
	-- is used to fill unused half octets.
	-- Refers to the Group Identification as specified in 3GPP TS 23.003 and 3GPP TS 43.068/ 43.069
*/
type gID [3]byte

func (d gID) String() string {
	return hex.EncodeToString(d[:])
}

func (d *gID) UnmarshalJSON(b []byte) error {
	var s string
	if e := json.Unmarshal(b, &s); e != nil {
		return e
	}
	if tmp, e := hex.DecodeString(s); e != nil {
		return e
	} else if len(tmp) != 3 {
		return fmt.Errorf("data length is not 3 octets")
	} else {
		d[0] = tmp[0]
		d[1] = tmp[1]
		d[2] = tmp[2]
	}
	return nil
}

func (d gID) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

/*
vGroupCallData

	VGCSDataList ::= SEQUENCE SIZE (1..maxNumOfVGCSGroupIds) OF
		VoiceGroupCallData
	VoiceGroupCallData  ::= SEQUENCE {
		groupId                     GroupId,
		extensionContainer          ExtensionContainer      OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		additionalSubscriptions     AdditionalSubscriptions OPTIONAL,
		additionalInfo          [0] AdditionalInfo          OPTIONAL,
		longGroupId             [1] Long-GroupId            OPTIONAL }

	-- groupId shall be filled with six TBCD fillers (1111)if the longGroupId is present
	-- VoiceGroupCallData containing a longGroupId shall not be sent to VLRs that did not
	-- indicate support of long Group IDs within the Update Location or Restore Data
	-- request message
*/
type vGroupCallData struct {
	GroupID gID `json:"groupId"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (d vGroupCallData) String() string {
	return fmt.Sprintf("groupId=%s", d.GroupID)
}

func (d vGroupCallData) marshal() []byte {
	buf := new(bytes.Buffer)
	// groupId, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, d.GroupID[:])

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return buf.Bytes()
}

func (d *vGroupCallData) unmarshal(v []byte) error {
	buf := bytes.NewBuffer(v)

	// groupId, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return e
	} else if len(v) != 3 {
		return fmt.Errorf("invalid length")
	} else {
		d.GroupID[0] = v[0]
		d.GroupID[1] = v[1]
		d.GroupID[2] = v[2]
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return e
		}

		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return nil
			} else if e != nil {
				return e
			}
		*/
	}

	return nil
}

/*
vBroadcastData

	VBSDataList ::= SEQUENCE SIZE (1..maxNumOfVBSGroupIds) OF
		VoiceBroadcastData
	VoiceBroadcastData ::= SEQUENCE {
		groupid                  GroupId,
		broadcastInitEntitlement NULL               OPTIONAL,
		extensionContainer       ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		longGroupId	[0] Long-GroupId	OPTIONAL }

		-- groupId shall be filled with six TBCD fillers (1111)if the longGroupId is present
		-- VoiceBroadcastData containing a longGroupId shall not be sent to VLRs that did not
		-- indicate support of long Group IDs within the Update Location or Restore Data
		-- request message
*/
type vBroadcastData struct {
	GroupID       gID  `json:"groupId"`
	BroadcastInit bool `json:"broadcastInitEntitlement,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (d vBroadcastData) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "groupId=%s", d.GroupID)
	if d.BroadcastInit {
		fmt.Fprint(buf, ", broadcastInitEntitlement")
	}
	return buf.String()
}

func (d vBroadcastData) marshal() []byte {
	buf := new(bytes.Buffer)
	// groupId, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, d.GroupID[:])

	// broadcastInitEntitlement, universal(00) + primitive(00) + null(05)
	if d.BroadcastInit {
		gsmap.WriteTLV(buf, 0x05, nil)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return buf.Bytes()
}

func (d *vBroadcastData) unmarshal(v []byte) error {
	buf := bytes.NewBuffer(v)
	// groupId, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return e
	} else if len(v) != 3 {
		return fmt.Errorf("invalid length")
	} else {
		d.GroupID[0] = v[0]
		d.GroupID[1] = v[1]
		d.GroupID[2] = v[2]
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	// broadcastInitEntitlement, universal(00) + primitive(00) + null(05)
	if t == 0x05 {
		d.BroadcastInit = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return e
		}

		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return nil
			} else if e != nil {
				return e
			}
		*/
	}

	return nil
}

/*
VlrCamelSubsInfo

	VlrCamelSubscriptionInfo ::= SEQUENCE {
		o-CSI                         [0]  O-CSI                         OPTIONAL,
		extensionContainer            [1]  ExtensionContainer            OPTIONAL,
		...,
		ss-CSI                        [2]  SS-CSI                        OPTIONAL,
		o-BcsmCamelTDP-CriteriaList   [4]  O-BcsmCamelTDPCriteriaList    OPTIONAL,
		tif-CSI                       [3]  NULL                          OPTIONAL,
		m-CSI                         [5]  M-CSI                         OPTIONAL,
		mo-sms-CSI                    [6]  SMS-CSI                       OPTIONAL,
		vt-CSI                        [7]  T-CSI                         OPTIONAL,
		t-BCSM-CAMEL-TDP-CriteriaList [8]  T-BCSM-CAMEL-TDP-CriteriaList OPTIONAL,
		d-CSI                         [9]  D-CSI                         OPTIONAL,
		mt-sms-CSI                    [10] SMS-CSI                       OPTIONAL,
		mt-smsCAMELTDP-CriteriaList   [11] MT-smsCAMELTDP-CriteriaList   OPTIONAL
		}
*/
type VlrCamelSubsInfo struct {
	// ToDo
}

/*
NAEAPreferredCI

	NAEA-PreferredCI ::= SEQUENCE {
		naea-PreferredCIC	[0] NAEA-CIC,
		extensionContainer	[1] ExtensionContainer	OPTIONAL,
		...}
	NAEA-CIC ::= OCTET STRING (SIZE (3))
		-- The internal structure is defined by the Carrier Identification
		-- parameter in ANSI T1.113.3. Carrier codes between "000" and "999" may
		-- be encoded as 3 digits using "000" to "999" or as 4 digits using
		-- "0000" to "0999". Carrier codes between "1000" and "9999" are encoded
		-- using 4 digits.
*/
type NAEAPreferredCI struct {
	PrefCIC [3]byte `json:"naea-PreferredCIC"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (n NAEAPreferredCI) String() string {
	return fmt.Sprintf("naea-PreferredCIC=%x", n.PrefCIC)
}

func (n *NAEAPreferredCI) UnmarshalJSON(b []byte) (e error) {
	tmp := struct {
		PrefCIC string `json:"naea-PreferredCIC"`
	}{}
	if e = json.Unmarshal(b, &tmp); e != nil {
		return
	}
	if b, e = hex.DecodeString(tmp.PrefCIC); e != nil {
		return
	}
	if len(b) != 3 {
		e = gsmap.UnexpectedTLV("length must 3")
	} else {
		n.PrefCIC[0] = tmp.PrefCIC[0]
		n.PrefCIC[1] = tmp.PrefCIC[1]
		n.PrefCIC[2] = tmp.PrefCIC[2]
	}

	return
}

func (n NAEAPreferredCI) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrefCIC string `json:"naea-PreferredCIC"`
	}{
		PrefCIC: hex.EncodeToString(n.PrefCIC[:])})
}

func (n *NAEAPreferredCI) unmarshal(b []byte) error {
	buf := bytes.NewBuffer(b)

	// naea-PreferredCIC, context_specific(80) + constructed(20) + 0(00)
	if _, v, e := gsmap.ReadTLV(buf, 0xa0); e != nil {
		return e
	} else if len(v) != 3 {
		return gsmap.UnexpectedTLV("length must 3")
	} else {
		n.PrefCIC[0] = v[0]
		n.PrefCIC[1] = v[1]
		n.PrefCIC[2] = v[2]
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	// extensionContainer, context_specific(80) + constructed(20) + 1(01)
	if t == 0xa1 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return e
		}
	}

	return nil
}

func (n NAEAPreferredCI) marshal() []byte {
	buf := new(bytes.Buffer)

	// naea-PreferredCIC, context_specific(80) + constructed(20) + 0(00)
	gsmap.WriteTLV(buf, 0xa0, n.PrefCIC[:])

	// extensionContainer, context_specific(80) + constructed(20) + 1(01)

	return buf.Bytes()
}

/*
nwAccessMode

	NetworkAccessMode ::= ENUMERATED {
		packetAndCircuit (0),
		onlyCircuit      (1),
		onlyPacket       (2),
		...}
		-- if unknown values are received in NetworkAccessMode
		-- they shall be discarded.
*/
type nwAccessMode byte

const (
	_ nwAccessMode = iota
	PacketAndCircuit
	OnlyCircuit
	OnlyPacket
)

func (m nwAccessMode) String() string {
	switch m {
	case PacketAndCircuit:
		return "packetAndCircuit"
	case OnlyCircuit:
		return "onlyCircuit"
	case OnlyPacket:
		return "onlyPacket"
	}
	return ""
}

func (m *nwAccessMode) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "packetAndCircuit":
		*m = PacketAndCircuit
	case "onlyCircuit":
		*m = OnlyCircuit
	case "onlyPacket":
		*m = OnlyPacket
	default:
		*m = 0
	}
	return
}

func (m nwAccessMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *nwAccessMode) unmarshal(b []byte) error {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*m = PacketAndCircuit
	case 1:
		*m = OnlyCircuit
	case 2:
		*m = OnlyPacket
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (m nwAccessMode) marshal() []byte {
	switch m {
	case PacketAndCircuit:
		return []byte{0x00}
	case OnlyCircuit:
		return []byte{0x01}
	case OnlyPacket:
		return []byte{0x02}
	}
	return nil
}

/*
LSAInformation

	LSAInformation ::= SEQUENCE {
		completeDataListIncluded   NULL                   OPTIONAL,
		lsaOnlyAccessIndicator [1] LSAOnlyAccessIndicator OPTIONAL,
		lsaDataList            [2] LSADataList            OPTIONAL,
		extensionContainer     [3] ExtensionContainer     OPTIONAL,
		...}
	-- If segmentation is used, completeDataListIncluded may only be present in the
	-- first segment.

	LSAOnlyAccessIndicator ::= ENUMERATED {
		accessOutsideLSAsAllowed    (0),
		accessOutsideLSAsRestricted (1)}
	LSADataList ::= SEQUENCE SIZE (1..maxNumOfLSAs) OF
		LSAData
	LSAData ::= SEQUENCE {
		lsaIdentity            [0] LSAIdentity,
		lsaAttributes          [1] LSAAttributes,
		lsaActiveModeIndicator [2] NULL               OPTIONAL,
		extensionContainer     [3] ExtensionContainer OPTIONAL,
		...}
*/
type LSAInformation struct {
	Complete bool `json:"completeDataListIncluded,omitempty"`
	// Extension
}

/*
LCSInformation

	LCSInformation ::= SEQUENCE {
		gmlc-List                    [0] GMLC-List                OPTIONAL,
		lcs-PrivacyExceptionList     [1] LCS-PrivacyExceptionList OPTIONAL,
		molr-List                    [2] MOLR-List                OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		add-lcs-PrivacyExceptionList [3] LCS-PrivacyExceptionList OPTIONAL }
		-- add-lcs-PrivacyExceptionList may be sent only if lcs-PrivacyExceptionList is
		-- present and contains four instances of LCS-PrivacyClass. If the mentioned condition
		-- is not satisfied the receiving node shall discard add-lcs-PrivacyExceptionList.
		-- If an LCS-PrivacyClass is received both in lcs-PrivacyExceptionList and in
		-- add-lcs-PrivacyExceptionList with the same SS-Code, then the error unexpected
		-- data value shall be returned.

	GMLC-List ::= SEQUENCE SIZE (1..maxNumOfGMLC) OF
		ISDN-AddressString
		-- if segmentation is used, the complete GMLC-List shall be sent in one segment

	LCS-PrivacyExceptionList ::= SEQUENCE SIZE (1..maxNumOfPrivacyClass) OF
		LCS-PrivacyClass
	LCS-PrivacyClass ::= SEQUENCE {
		ss-Code                  SS-Code,
		ss-Status                Ext-SS-Status,
		notificationToMSUser [0] NotificationToMSUser	OPTIONAL,
		-- notificationToMSUser may be sent only for SS-codes callSessionRelated
		-- and callSessionUnrelated. If not received for SS-codes callSessionRelated
		-- and callSessionUnrelated,
		-- the default values according to 3GPP TS 23.271 shall be assumed.
		externalClientList   [1] ExternalClientList	OPTIONAL,
		-- externalClientList may be sent only for SS-code callSessionUnrelated to a
		-- visited node that does not support LCS Release 4 or later versions.
		-- externalClientList may be sent only for SS-codes callSessionUnrelated and
		-- callSessionRelated to a visited node that supports LCS Release 4 or later versions.
		plmnClientList       [2] PLMNClientList	OPTIONAL,
		-- plmnClientList may be sent only for SS-code plmnoperator.
		extensionContainer   [3] ExtensionContainer	OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		ext-externalClientList	[4] Ext-ExternalClientList	OPTIONAL,
		-- Ext-externalClientList may be sent only if the visited node supports LCS Release 4 or
		-- later versions, the user did specify more than 5 clients, and White Book SCCP is used.
		serviceTypeList	[5]	ServiceTypeList	OPTIONAL
		-- serviceTypeList may be sent only for SS-code serviceType and if the visited node
		-- supports LCS Release 5 or later versions.
		--
		-- if segmentation is used, the complete LCS-PrivacyClass shall be sent in one segment
	}

	MOLR-List ::= SEQUENCE SIZE (1..maxNumOfMOLR-Class) OF
		MOLR-Class
	MOLR-Class ::= SEQUENCE {
		ss-Code                SS-Code,
		ss-Status              Ext-SS-Status,
		extensionContainer [0] ExtensionContainer OPTIONAL,
		...}
*/

/*
MCSSInfo

	MC-SS-Info ::= SEQUENCE {
		ss-Code            [0] SS-Code,
		ss-Status          [1] Ext-SS-Status,
		nbrSB              [2] MaxMC-Bearers,
		nbrUser            [3] MC-Bearers,
		extensionContainer [4] ExtensionContainer OPTIONAL,
		...}
*/

/*
	SGSN-CAMEL-SubscriptionInfo ::= SEQUENCE {
		gprs-CSI	[0]	GPRS-CSI	OPTIONAL,
		mo-sms-CSI	[1]	SMS-CSI	OPTIONAL,
		extensionContainer	[2]	ExtensionContainer	OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		mt-sms-CSI	[3]	SMS-CSI	OPTIONAL,
		mt-smsCAMELTDP-CriteriaList	[4]	MT-smsCAMELTDP-CriteriaList	OPTIONAL,
		mg-csi	[5]	MG-CSI	OPTIONAL
		}

	GPRS-CSI ::= SEQUENCE {
		gprs-CamelTDPDataList	[0] GPRS-CamelTDPDataList	OPTIONAL,
		camelCapabilityHandling	[1] CamelCapabilityHandling	OPTIONAL,
		extensionContainer	[2] ExtensionContainer	OPTIONAL,
		notificationToCSE	[3]	NULL	OPTIONAL,
		csi-Active	[4]	NULL	OPTIONAL,
		...}
	--	notificationToCSE and csi-Active shall not be present when GPRS-CSI is sent to SGSN.
	--	They may only be included in ATSI/ATM ack/NSDC message.
	--	GPRS-CamelTDPData and  camelCapabilityHandling shall be present in
	--	the GPRS-CSI sequence.
	--	If GPRS-CSI is segmented, gprs-CamelTDPDataList and camelCapabilityHandling shall be
	--	present in the first segment

	GPRS-CamelTDPDataList ::= SEQUENCE SIZE (1..maxNumOfCamelTDPData) OF
		GPRS-CamelTDPData
	--	GPRS-CamelTDPDataList shall not contain more than one instance of
	--	GPRS-CamelTDPData containing the same value for gprs-TriggerDetectionPoint.

	GPRS-CamelTDPData ::= SEQUENCE {
		gprs-TriggerDetectionPoint	[0] GPRS-TriggerDetectionPoint,
		serviceKey	[1] ServiceKey,
		gsmSCF-Address	[2] ISDN-AddressString,
		defaultSessionHandling	[3] DefaultGPRS-Handling,
		extensionContainer	[4] ExtensionContainer	OPTIONAL,
		...
		}

	DefaultGPRS-Handling ::= ENUMERATED {
		continueTransaction (0) ,
		releaseTransaction (1) ,
		...}
	-- exception handling:
	-- reception of values in range 2-31 shall be treated as "continueTransaction"
	-- reception of values greater than 31 shall be treated as "releaseTransaction"

	GPRS-TriggerDetectionPoint ::= ENUMERATED {
		attach	(1),
		attachChangeOfPosition	(2),
		pdp-ContextEstablishment	(11),
		pdp-ContextEstablishmentAcknowledgement	(12),
		pdp-ContextChangeOfPosition	(14),
		... }
	-- exception handling:
	-- For GPRS-CamelTDPData sequences containing this parameter with any
	-- other value than the ones listed the receiver shall ignore the whole
	-- GPRS-CamelTDPDatasequence.

	SMS-CSI ::= SEQUENCE {
		sms-CAMEL-TDP-DataList	[0] SMS-CAMEL-TDP-DataList	OPTIONAL,
		camelCapabilityHandling	[1] CamelCapabilityHandling	OPTIONAL,
		extensionContainer	[2] ExtensionContainer	OPTIONAL,
		notificationToCSE	[3] NULL	OPTIONAL,
		csi-Active	[4] NULL	OPTIONAL,
		...}
	--	notificationToCSE and csi-Active shall not be present
	--	when MO-SMS-CSI or MT-SMS-CSI is sent to VLR or SGSN.
	--	They may only be included in ATSI/ATM ack/NSDC message.
	--	SMS-CAMEL-TDP-Data and  camelCapabilityHandling shall be present in
	--	the SMS-CSI sequence.
	--	If SMS-CSI is segmented, sms-CAMEL-TDP-DataList and camelCapabilityHandling shall be
	--	present in the first segment

	SMS-CAMEL-TDP-DataList ::= SEQUENCE SIZE (1..maxNumOfCamelTDPData) OF
		SMS-CAMEL-TDP-Data
	--	SMS-CAMEL-TDP-DataList shall not contain more than one instance of
	--	SMS-CAMEL-TDP-Data containing the same value for sms-TriggerDetectionPoint.

	SMS-CAMEL-TDP-Data ::= SEQUENCE {
		sms-TriggerDetectionPoint	[0] SMS-TriggerDetectionPoint,
		serviceKey	[1] ServiceKey,
		gsmSCF-Address	[2] ISDN-AddressString,
		defaultSMS-Handling	[3] DefaultSMS-Handling,
		extensionContainer	[4] ExtensionContainer	OPTIONAL,
		...
		}

	SMS-TriggerDetectionPoint ::= ENUMERATED {
		sms-CollectedInfo (1),
		...,
		sms-DeliveryRequest (2)
		}
	--	exception handling:
	--	For SMS-CAMEL-TDP-Data and MT-smsCAMELTDP-Criteria sequences containing this
	--	parameter with any other value than the ones listed the receiver shall ignore
	--	the whole sequence.
	--
	--	If this parameter is received with any other value than sms-CollectedInfo
	--	in an SMS-CAMEL-TDP-Data sequence contained in mo-sms-CSI, then the receiver shall
	--	ignore the whole SMS-CAMEL-TDP-Data sequence.
	--
	--	If this parameter is received with any other value than sms-DeliveryRequest
	--	in an SMS-CAMEL-TDP-Data sequence contained in mt-sms-CSI then the receiver shall
	--	ignore the whole SMS-CAMEL-TDP-Data sequence.
	--
	--	If this parameter is received with any other value than sms-DeliveryRequest
	--	in an MT-smsCAMELTDP-Criteria sequence then the receiver shall
	--	ignore the whole MT-smsCAMELTDP-Criteria sequence.

	DefaultSMS-Handling ::= ENUMERATED {
		continueTransaction (0) ,
		releaseTransaction (1) ,
		...}
	--	exception handling:
	--	reception of values in range 2-31 shall be treated as "continueTransaction"
	--	reception of values greater than 31 shall be treated as "releaseTransaction"

*/
