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

/*
Identity parameter.

	Identity ::= CHOICE {
		imsi          IMSI,
		imsi-WithLMSI IMSI-WithLMSI}

	IMSI-WithLMSI ::= SEQUENCE {
		imsi IMSI,
		lmsi LMSI,
		-- a special value 00000000 indicates that the LMSI is not in use
		...}
*/
type Identity struct {
	IMSI teldata.IMSI `json:"imsi"`
	LMSI teldata.LMSI `json:"lmsi,omitempty"`
}

func (i Identity) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "\n%s| imsi: %s", gsmap.LogPrefix, i.IMSI)
	if !i.LMSI.IsEmpty() {
		fmt.Fprintf(buf, "\n%s| lmsi: %s", gsmap.LogPrefix, i.LMSI)
	}
	return buf.String()
}

func (i *Identity) unmarshalFrom(buf *bytes.Buffer) error {
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e != nil {
		return e
	}

	switch t {
	case 0x04:
		i.IMSI, e = teldata.DecodeIMSI(v)
	case 0x30:
		buf = bytes.NewBuffer(v)
		if _, v, e = gsmap.ReadTLV(buf, 0x04); e != nil {
		} else if i.IMSI, e = teldata.DecodeIMSI(v); e != nil {
		} else if _, v, e = gsmap.ReadTLV(buf, 0x04); e != nil {
		} else if i.LMSI, e = teldata.DecodeLMSI(v); e != nil {
		}
	default:
		e = gsmap.UnexpectedTag([]byte{0x04, 0x30}, t)
	}
	return e
}

func (i Identity) marshalTo(buf *bytes.Buffer) error {
	if i.LMSI.IsEmpty() {
		// imsi, universal(00) + primitive(00) + octet_string(04)
		gsmap.WriteTLV(buf, 0x04, i.IMSI.Bytes())
	} else {
		buf2 := new(bytes.Buffer)
		// imsi, universal(00) + primitive(00) + octet_string(04)
		gsmap.WriteTLV(buf2, 0x04, i.IMSI.Bytes())
		// lmsi, universal(00) + primitive(00) + octet_string(04)
		gsmap.WriteTLV(buf2, 0x04, i.LMSI.Bytes())
		// imsi-WithLMSI, universal(00) + constructed(20) + sequence(10)
		gsmap.WriteTLV(buf, 0x30, buf.Bytes())
	}
	return nil
}

/*
CancellationType parameter.

	CancellationType ::= ENUMERATED {
		updateProcedure        (0),
		subscriptionWithdraw   (1),
		...,
		-- ^^^^^^^^ R99 ^^^^^^^^
		initialAttachProcedure (2)}
		-- The HLR shall not send values other than listed above
*/
type CancellationType byte

const (
	_ CancellationType = iota
	UpdateProcedure
	SubscriptionWithdraw
)

func (c CancellationType) String() string {
	switch c {
	case UpdateProcedure:
		return "updateProcedure"
	case SubscriptionWithdraw:
		return "subscriptionWithdraw"
	}
	return ""
}

func (c *CancellationType) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "updateProcedure":
		*c = UpdateProcedure
	case "subscriptionWithdraw":
		*c = SubscriptionWithdraw
	default:
		*c = 0
	}
	return
}

func (c CancellationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *CancellationType) unmarshal(b []byte) error {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = UpdateProcedure
	case 1:
		*c = SubscriptionWithdraw
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (c CancellationType) marshal() []byte {
	switch c {
	case UpdateProcedure:
		return []byte{0x00}
	case SubscriptionWithdraw:
		return []byte{0x01}
	}
	return nil
}

/*
vlrCapability parameter.

	VLR-Capability ::= SEQUENCE{
		supportedCamelPhases           [0]  SupportedCamelPhases         OPTIONAL,
		extensionContainer                  ExtensionContainer           OPTIONAL,
		... ,
		solsaSupportIndicator          [2]  NULL                         OPTIONAL,
		istSupportIndicator            [1]  IST-SupportIndicator         OPTIONAL,
		superChargerSupportedInServingNetworkEntity [3] SuperChargerInfo OPTIONAL,
		longFTN-Supported              [4]  NULL                         OPTIONAL,
		-- ^^^^^^^^ R99 ^^^^^^^^
		supportedLCS-CapabilitySets    [5]  SupportedLCS-CapabilitySets  OPTIONAL,
		offeredCamel4CSIs              [6]  OfferedCamel4CSIs            OPTIONAL,
		supportedRAT-TypesIndicator    [7]  SupportedRAT-Types           OPTIONAL,
		longGroupID-Supported          [8]  NULL                         OPTIONAL,
		mtRoamingForwardingSupported   [9]  NULL                         OPTIONAL,
		msisdn-lessOperation-Supported [10] NULL                         OPTIONAL,
		reset-ids-Supported            [11] NULL                         OPTIONAL }

	SuperChargerInfo ::= CHOICE {
		sendSubscriberData   [0] NULL,
		subscriberDataStored [1] AgeIndicator }

	AgeIndicator ::= OCTET STRING (SIZE (1..6))
		-- The internal structure of this parameter is implementation specific.
*/
type vlrCapability struct {
	SupportedCamelPh supportedCamelPh `json:"supportedCamelPhases,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	SolsaSupport     bool       `json:"solsaSupportIndicator,omitempty"`
	IstSupport       istSupport `json:"istSupportIndicator,omitempty"`
	SuperChargerInfo []byte     `json:"superChargerSupportedInServingNetworkEntity,omitempty"`
	LongFTNSupport   bool       `json:"longFTN-Supported,omitempty"`
}

func (v vlrCapability) String() string {
	buf := new(strings.Builder)
	if v.SupportedCamelPh != 0 {
		fmt.Fprintf(buf, "\n%s| supportedCamelPhases: %s",
			gsmap.LogPrefix, v.SupportedCamelPh)
	}
	// Extension
	if v.SolsaSupport {
		fmt.Fprintf(buf, "\n%s| solsaSupportIndicator:", gsmap.LogPrefix)
	}
	if v.IstSupport != 0 {
		fmt.Fprintf(buf, "\n%s| istSupportIndicator: %s", gsmap.LogPrefix, v.IstSupport)
	}
	if v.SuperChargerInfo != nil {
		fmt.Fprintf(buf, "\n%s| superChargerSupportedInServingNetworkEntity:",
			gsmap.LogPrefix)
		if len(v.SuperChargerInfo) == 0 {
			fmt.Fprintf(buf, "\n%s| | sendSubscriberData:", gsmap.LogPrefix)
		} else {
			fmt.Fprintf(buf, "\n%s| | subscriberDataStored: %x",
				gsmap.LogPrefix, v.SuperChargerInfo)
		}
	}
	if v.LongFTNSupport {
		fmt.Fprintf(buf, "\n%s| longFTN-Supported:", gsmap.LogPrefix)
	}

	return buf.String()
}

func (vc *vlrCapability) marshal() []byte {
	buf := new(bytes.Buffer)

	// supportedCamelPhases, context_specific(80) + primitive(00) + 0(00)
	if vc.SupportedCamelPh != 0 {
		gsmap.WriteTLV(buf, 0x80, vc.SupportedCamelPh.marshal())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// solsaSupportIndicator, context_specific(80) + primitive(00) + 2(02)
	if vc.SolsaSupport {
		gsmap.WriteTLV(buf, 0x82, nil)
	}

	// istSupportIndicator, context_specific(80) + primitive(00) + 1(01)
	if b := vc.IstSupport.marshal(); b != nil {
		gsmap.WriteTLV(buf, 0x81, b)
	}

	// superChargerSupportedInServingNetworkEntity,
	// context_specific(80) + primitive(00) + 3(03)
	if vc.SuperChargerInfo != nil {
		if len(vc.SuperChargerInfo) == 0 {
			gsmap.WriteTLV(buf, 0x83,
				gsmap.WriteTLV(new(bytes.Buffer), 0x80, nil))
		} else {
			gsmap.WriteTLV(buf, 0x83,
				gsmap.WriteTLV(new(bytes.Buffer), 0x81, vc.SuperChargerInfo))
		}
	}

	// longFTN-Supported, context_specific(80) + primitive(00) + 4(04)
	if vc.LongFTNSupport {
		gsmap.WriteTLV(buf, 0x84, nil)
	}

	return buf.Bytes()
}

func (vc *vlrCapability) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	// supportedCamelPhases, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if e = vc.SupportedCamelPh.unmarshal(v); e != nil {
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
		if _, e = common.UnmarshalExtension(v); e != nil {
			return e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// solsaSupportIndicator, context_specific(80) + primitive(00) + 2(02)
	if t == 0x82 {
		vc.SolsaSupport = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// istSupportIndicator, context_specific(80) + primitive(00) + 1(01)
	if t == 0x81 {
		if e = vc.IstSupport.unmarshal(v); e != nil {
			return e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// superChargerSupportedInServingNetworkEntity,
	// context_specific(80) + primitive(00) + 3(03)
	if t == 0x83 {
		if t, v, e = gsmap.ReadTLV(bytes.NewBuffer(v), 0x00); e != nil {
			return e
		}
		switch t {
		case 0x80:
			vc.SuperChargerInfo = []byte{}
		case 0x81:
			vc.SuperChargerInfo = v
		}

		if t, _, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// longFTN-Supported, context_specific(80) + primitive(00) + 4(04)
	if t == 0x84 {
		vc.LongFTNSupport = true
	}

	return nil
}

/*
supportedCamelPh

	SupportedCamelPhases ::= BIT STRING {
		phase1 (0),
		phase2 (1),
		phase3 (2),
		-- ^^^^^^^^ R99 ^^^^^^^^
		phase4 (3)} (SIZE (1..16))
		-- A node shall mark in the BIT STRING all CAMEL Phases it supports.
		-- Other values than listed above shall be discarded.
*/
type supportedCamelPh byte

func (s supportedCamelPh) String() string {
	return fmt.Sprintf("ph1=%t, ph2=%t, ph3=%t",
		s&0x80 == 0x80, s&0x40 == 0x40, s&0x20 == 0x20)
}

func (s *supportedCamelPh) UnmarshalJSON(b []byte) (e error) {
	tmp := struct {
		Ph1 bool `json:"phase1,omitempty"`
		Ph2 bool `json:"phase2,omitempty"`
		Ph3 bool `json:"phase3,omitempty"`
	}{}
	if e = json.Unmarshal(b, &tmp); e != nil {
		return
	}
	if tmp.Ph1 {
		*s |= 0x80
	}
	if tmp.Ph2 {
		*s |= 0x40
	}
	if tmp.Ph3 {
		*s |= 0x20
	}
	return
}

func (s supportedCamelPh) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Ph1 bool `json:"phase1,omitempty"`
		Ph2 bool `json:"phase2,omitempty"`
		Ph3 bool `json:"phase3,omitempty"`
	}{
		Ph1: s&0x80 == 0x80,
		Ph2: s&0x40 == 0x40,
		Ph3: s&0x20 == 0x20})
}

func (s *supportedCamelPh) unmarshal(b []byte) error {
	if len(b) < 2 {
		return gsmap.UnexpectedTLV("length must <2")
	}
	if len(b) == 2 {
		switch b[0] {
		case 7:
			b[1] &= 0x80
		case 6:
			b[1] &= 0xc0
		}
	}
	*s = supportedCamelPh(b[1] & 0xe0)
	return nil
}

func (s supportedCamelPh) marshal() []byte {
	return []byte{0x05, byte(s)}
}

/*
istSupport

	IST-SupportIndicator ::=  ENUMERATED {
		basicISTSupported	(0),
		istCommandSupported	(1),
		...}
		-- exception handling:
		-- reception of values > 1 shall be mapped to ' istCommandSupported '
*/
type istSupport byte

const (
	_ istSupport = iota
	BasicIstSupported
	IstCommandSupported
)

func (i istSupport) String() string {
	switch i {
	case BasicIstSupported:
		return "basicISTSupported"
	case IstCommandSupported:
		return "istCommandSupported"
	}
	return ""
}

func (i *istSupport) UnmarshalJSON(b []byte) (e error) {
	var s string
	if e = json.Unmarshal(b, &s); e != nil {
		return
	}
	switch s {
	case "basicISTSupported":
		*i = BasicIstSupported
	case "istCommandSupported":
		*i = IstCommandSupported
	default:
		*i = 0
	}
	return
}

func (s istSupport) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *istSupport) unmarshal(b []byte) error {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*s = BasicIstSupported
	case 1:
		*s = IstCommandSupported
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (s istSupport) marshal() []byte {
	switch s {
	case BasicIstSupported:
		return []byte{0x00}
	case IstCommandSupported:
		return []byte{0x01}
	}
	return nil
}

/*
regionalSubscriptionRes

	RegionalSubscriptionResponse ::= ENUMERATED {
		networkNode-AreaRestricted (0),
		tooManyZoneCodes           (1),
		zoneCodesConflict          (2),
		regionalSubscNotSupported  (3)}
*/
type regionalSubscriptionRes byte

const (
	_ regionalSubscriptionRes = iota
	NetworkNodeAreaRestricted
	TooManyZoneCodes
	ZoneCodesConflict
	RegionalSubscNotSupported
)

func (r regionalSubscriptionRes) String() string {
	switch r {
	case NetworkNodeAreaRestricted:
		return "networkNode-AreaRestricted"
	case TooManyZoneCodes:
		return "tooManyZoneCodes"
	case ZoneCodesConflict:
		return "zoneCodesConflict"
	case RegionalSubscNotSupported:
		return "regionalSubscNotSupported"
	}
	return ""
}

func (r *regionalSubscriptionRes) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "networkNode-AreaRestricted":
		*r = NetworkNodeAreaRestricted
	case "tooManyZoneCodes":
		*r = TooManyZoneCodes
	case "zoneCodesConflict":
		*r = ZoneCodesConflict
	case "regionalSubscNotSupported":
		*r = RegionalSubscNotSupported
	default:
		*r = 0
	}
	return
}

func (r regionalSubscriptionRes) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *regionalSubscriptionRes) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*r = NetworkNodeAreaRestricted
	case 1:
		*r = TooManyZoneCodes
	case 2:
		*r = ZoneCodesConflict
	case 3:
		*r = RegionalSubscNotSupported
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (r regionalSubscriptionRes) marshal() []byte {
	switch r {
	case NetworkNodeAreaRestricted:
		return []byte{0x00}
	case TooManyZoneCodes:
		return []byte{0x01}
	case ZoneCodesConflict:
		return []byte{0x02}
	case RegionalSubscNotSupported:
		return []byte{0x03}
	}
	return nil
}
