package ifc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/teldata"
)

/*
LocationInfoWithLMSI

	LocationInfoWithLMSI ::= SEQUENCE {
		networkNode-Number [1] ISDN-AddressString,
		lmsi                   LMSI               OPTIONAL,
		extensionContainer     ExtensionContainer OPTIONAL,
		...,
		gprsNodeIndicator                    [5]  NULL                       OPTIONAL,
		additional-Number                    [6]  Additional-Number          OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		networkNodeDiameterAddress           [7]  NetworkNodeDiameterAddress OPTIONAL,
		additionalNetworkNodeDiameterAddress [8]  NetworkNodeDiameterAddress OPTIONAL,
		thirdNumber                          [9]  Additional-Number          OPTIONAL,
		thirdNetworkNodeDiameterAddress      [10] NetworkNodeDiameterAddress OPTIONAL,
		imsNodeIndicator                     [11] NULL                       OPTIONAL,
		smsf-3gpp-Number                     [12] ISDN-AddressString         OPTIONAL,
		smsf-3gpp-DiameterAddress            [13] NetworkNodeDiameterAddress OPTIONAL,
		smsf-non-3gpp-Number                 [14] ISDN-AddressString         OPTIONAL,
		smsf-non-3gpp-DiameterAddress        [15] NetworkNodeDiameterAddress OPTIONAL,
		smsf-3gpp-address-indicator          [16] NULL                       OPTIONAL,
		smsf-non-3gpp-address-indicator      [17] NULL                       OPTIONAL }

	Additional-Number ::= CHOICE {
		msc-Number  [0] ISDN-AddressString,
		sgsn-Number [1] ISDN-AddressString }

gprsNodeIndicator is set only if the SGSN number is sent as the Network Node Number.
msc-number can be the MSC number or the SMS Router number or the MME number for MT SMS.
sgsn-number can be the SGSN number or the SMS Router number.
*/
type LocationInfoWithLMSI struct {
	NodeNumber NodeNumber
	LMSI       teldata.LMSI
	// Extension
	// NumberIsGPRS      bool
	AdditionalNumber NodeNumber
}

type NodeNumber struct {
	Address gsmap.AddressString
	IsGPRS  bool
}

func (l LocationInfoWithLMSI) MarshalJSON() ([]byte, error) {
	type AdditionalNumber struct {
		MSC  *gsmap.AddressString `json:"msc-Number,omitempty"`
		SGSN *gsmap.AddressString `json:"sgsn-Number,omitempty"`
	}
	j := struct {
		NN   gsmap.AddressString `json:"networkNode-Number"`
		LMSI *teldata.LMSI       `json:"lmsi,omitempty"`
		GPRS bool                `json:"gprsNodeIndicator,omitempty"`
		ANN  *AdditionalNumber   `json:"additional-Number,omitempty"`
	}{
		NN:   l.NodeNumber.Address,
		GPRS: l.NodeNumber.IsGPRS,
	}

	if !l.LMSI.IsEmpty() {
		j.LMSI = &l.LMSI
	}
	if l.AdditionalNumber.Address.IsEmpty() {
	} else if l.AdditionalNumber.IsGPRS {
		j.ANN = &AdditionalNumber{SGSN: &l.AdditionalNumber.Address}
	} else {
		j.ANN = &AdditionalNumber{MSC: &l.AdditionalNumber.Address}
	}
	return json.Marshal(j)
}

func (l *LocationInfoWithLMSI) UnmarshalJSON(b []byte) (e error) {
	j := struct {
		NN   gsmap.AddressString `json:"networkNode-Number"`
		LMSI *teldata.LMSI       `json:"lmsi,omitempty"`
		GPRS bool                `json:"gprsNodeIndicator,omitempty"`
		ANN  *struct {
			MSC  *gsmap.AddressString `json:"msc-Number,omitempty"`
			SGSN *gsmap.AddressString `json:"sgsn-Number,omitempty"`
		} `json:"additional-Number,omitempty"`
	}{}

	if e = json.Unmarshal(b, &j); e == nil {
		l.NodeNumber.Address = j.NN
		l.NodeNumber.IsGPRS = j.GPRS
		if j.LMSI != nil {
			l.LMSI = *j.LMSI
		}
		if j.ANN != nil {
			if j.ANN.MSC != nil {
				l.AdditionalNumber.Address = *j.ANN.MSC
			} else if j.ANN.SGSN != nil {
				l.AdditionalNumber.Address = *j.ANN.SGSN
				l.AdditionalNumber.IsGPRS = true
			}
		}
	}
	return
}

func (l LocationInfoWithLMSI) marshal() []byte {
	buf := new(bytes.Buffer)

	// networkNode-Number, context_specific(80) + primitive(00) + 1(01)
	gsmap.WriteTLV(buf, 0x81, l.NodeNumber.Address.Bytes())

	// lmsi, universal(00) + primitive(00) + octet_string(04)
	if !l.LMSI.IsEmpty() {
		gsmap.WriteTLV(buf, 0x04, l.LMSI.Bytes())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// gprsNodeIndicator, context_specific(80) + primitive(00) + 5(05)
	if l.NodeNumber.IsGPRS {
		gsmap.WriteTLV(buf, 0x85, nil)
	}

	// additional-Number, context_specific(80) + primitive(00) + 6(06)
	if !l.AdditionalNumber.Address.IsEmpty() {
		var t byte
		if !l.AdditionalNumber.IsGPRS {
			// msc-Number, context_specific(80) + primitive(00) + 0(00)
			t = 0x80
		} else {
			// sgsn-Number, context_specific(80) + primitive(00) + 1(01)
			t = 0x81
		}
		gsmap.WriteTLV(buf, 0x86,
			gsmap.WriteTLV(new(bytes.Buffer), t, l.AdditionalNumber.Address.Bytes()))
	}

	return buf.Bytes()
}

func (l *LocationInfoWithLMSI) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	// networkNode-Number, context_specific(80) + primitive(00) + 1(01)
	if _, v, e := gsmap.ReadTLV(buf, 0x81); e != nil {
		return e
	} else if l.NodeNumber.Address, e = gsmap.DecodeAddressString(v); e != nil {
		return e
	}

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	// lmsi, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		if l.LMSI, e = teldata.DecodeLMSI(v); e != nil {
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

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// gprsNodeIndicator, context_specific(80) + primitive(00) + 5(05)
	if t == 0x85 {
		l.NodeNumber.IsGPRS = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// additional-Number, context_specific(80) + primitive(00) + 6(06)
	if t == 0x86 {
		if t, v, e = gsmap.ReadTLV(bytes.NewBuffer(v), 0x00); e != nil {
			return e
		}
		switch t {
		case 0x80: // msc-Number, context_specific(80) + primitive(00) + 0(00)
			l.AdditionalNumber.IsGPRS = false
		case 0x81: // sgsn-Number, context_specific(80) + primitive(00) + 1(01)
			l.AdditionalNumber.IsGPRS = true
		default:
			return gsmap.UnexpectedTag([]byte{0x80, 0x81}, t)
		}
		if l.AdditionalNumber.Address, e = gsmap.DecodeAddressString(v); e != nil {
			return e
		}
		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return nil
			} else if e != nil {
				return e
			}
		*/
	}

	return nil
}

/*
Outcome

	SM-DeliveryOutcome ::= ENUMERATED {
		memoryCapacityExceeded (0),
		absentSubscriber       (1),
		successfulTransfer     (2)}
*/
type Outcome byte

const (
	_ Outcome = iota
	MemoryCapacityExceeded
	AbsentSubscriber
	SuccessfulTransfer
)

func (o Outcome) String() string {
	switch o {
	case MemoryCapacityExceeded:
		return "memoryCapacityExceeded"
	case AbsentSubscriber:
		return "absentSubscriber"
	case SuccessfulTransfer:
		return "successfulTransfer"
	}
	return ""
}

func (o *Outcome) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "memoryCapacityExceeded":
		*o = MemoryCapacityExceeded
	case "absentSubscriber":
		*o = AbsentSubscriber
	case "successfulTransfer":
		*o = SuccessfulTransfer
	default:
		*o = 0
	}
	return
}

func (o Outcome) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *Outcome) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*o = MemoryCapacityExceeded
	case 1:
		*o = AbsentSubscriber
	case 2:
		*o = SuccessfulTransfer
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (o Outcome) marshal() []byte {
	switch o {
	case MemoryCapacityExceeded:
		return []byte{0x00}
	case AbsentSubscriber:
		return []byte{0x01}
	case SuccessfulTransfer:
		return []byte{0x02}
	}
	return nil
}

/*
MWStatus

	MW-Status ::= BIT STRING {
		sc-AddressNotIncluded (0),
		mnrf-Set              (1),
		mcef-Set              (2),
		mnrg-Set              (3),
		-- ^^^^^^^^ R99 ^^^^^^^^
		mnr5g-Set             (4),
		mnr5gn3g-Set          (5)} (SIZE (6..16))
*/
type MWStatus byte

func (s MWStatus) String() string {
	return fmt.Sprintf("sc-AddrNotIncluded=%t, mnrf=%t, mcef=%t, mnrg=%t",
		s&0x80 == 0x80, s&0x40 == 0x40, s&0x20 == 0x20, s&0x10 == 0x10)
}

func (s *MWStatus) UnmarshalJSON(b []byte) (e error) {
	tmp := struct {
		SCAddrNotIncluded bool `json:"sc-AddressNotIncluded,omitempty"`
		MNRF              bool `json:"mnrf-Set,omitempty"`
		MCEF              bool `json:"mcef-Set,omitempty"`
		MNRG              bool `json:"mnrg-Set,omitempty"`
	}{}
	if e = json.Unmarshal(b, &tmp); e != nil {
		return
	}
	if tmp.SCAddrNotIncluded {
		*s |= 0x80
	}
	if tmp.MNRF {
		*s |= 0x40
	}
	if tmp.MCEF {
		*s |= 0x20
	}
	if tmp.MNRG {
		*s |= 0x10
	}
	return

}

func (s MWStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SCAddrNotIncluded bool `json:"sc-AddressNotIncluded,omitempty"`
		MNRF              bool `json:"mnrf-Set,omitempty"`
		MCEF              bool `json:"mcef-Set,omitempty"`
		MNRG              bool `json:"mnrg-Set,omitempty"`
	}{
		SCAddrNotIncluded: s&0x80 == 0x80,
		MNRF:              s&0x40 == 0x40,
		MCEF:              s&0x20 == 0x20,
		MNRG:              s&0x10 == 0x10})
}

func (s *MWStatus) unmarshal(b []byte) error {
	if len(b) < 2 {
		return gsmap.UnexpectedTLV("length must <2")
	}
	if len(b) == 2 {
		switch b[0] {
		case 7:
			b[1] &= 0x80
		case 6:
			b[1] &= 0xc0
		case 5:
			b[1] &= 0xe0
		}
	}
	*s = MWStatus(b[1] & 0xf0)
	return nil
}

func (s MWStatus) marshal() []byte {
	return []byte{0x04, byte(s)}
}

/*
SMRPMTI

	SM-RP-MTI ::= INTEGER (0..10)
		-- 0 SMS Deliver
		-- 1 SMS Status Report
		-- other values are reserved for future use and shall be discarded if received
*/
type SMRPMTI byte

const (
	_ SMRPMTI = iota
	MTI_Deliver
	MTI_StatusReport
)

func (i SMRPMTI) String() string {
	switch i {
	case MTI_Deliver:
		return "SMS Deliver"
	case MTI_StatusReport:
		return "SMS Status Report"
	}
	return ""
}

func (i *SMRPMTI) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "SMS Deliver":
		*i = MTI_Deliver
	case "SMS Status Report":
		*i = MTI_StatusReport
	default:
		*i = 0
	}
	return
}

func (i SMRPMTI) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *SMRPMTI) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*i = MTI_Deliver
	case 1:
		*i = MTI_StatusReport
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (i SMRPMTI) marshal() []byte {
	switch i {
	case MTI_Deliver:
		return []byte{0x00}
	case MTI_StatusReport:
		return []byte{0x01}
	}
	return nil
}
