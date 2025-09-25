package ife

import (
	"bytes"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/common"
	"github.com/fkgi/teldata"
)

/*
RPAddr is address information for SM-RP-OA/DA.
roaming number must not be used in version greater 1.
noSM-RP-DA and noSM-RP-OA must not be used in version 1.

	SM-RP-DA ::= CHOICE {
		imsi                   [0] IMSI,               // for MT
		lmsi                   [1] LMSI,               // for MT
		roamingNumber          [3] ISDN-AddressString, // for MT
		serviceCentreAddressDA [4] AddressString,      // for MO
		noSM-RP-DA             [5] NULL }
	SM-RP-OA ::= CHOICE {
		msisdn                 [2] ISDN-AddressString, // for MO
		serviceCentreAddressOA [4] AddressString,      // for MT
		noSM-RP-OA             [5] NULL }
*/

type RpAddr interface {
	RpAddrID() []byte
}

type rpAddrJSON struct {
	IMSI       teldata.IMSI         `json:"imsi"`
	LMSI       teldata.LMSI         `json:"lmsi"`
	MSISDN     common.AddressString `json:"msisdn"`
	RoamingNum common.AddressString `json:"roamingNumber"`
	SCAddr     common.AddressString `json:"serviceCentreAddress"`
}

func (r rpAddrJSON) getRpAddr() RpAddr {
	if !r.IMSI.IsEmpty() {
		return RpIMSI{IMSI: r.IMSI}
	} else if !r.LMSI.IsEmpty() {
		return RpLMSI{LMSI: r.LMSI}
	} else if !r.MSISDN.IsEmpty() {
		return RpMSISDN{MSISDN: r.MSISDN}
	} else if !r.RoamingNum.IsEmpty() {
		return RpRoamingNumber{RoamingNum: r.RoamingNum}
	} else if !r.SCAddr.IsEmpty() {
		return RpSCAddress{SCAddr: r.SCAddr}
	}
	return nil
}

func unmarshalRPAddr(buf *bytes.Buffer) (RpAddr, error) {
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e != nil {
		return nil, e
	}
	switch t {
	case 0x80:
		// imsi, context_specific(80) + primitive(00) + 0(00)
		ret := RpIMSI{}
		if ret.IMSI, e = teldata.DecodeIMSI(v); e != nil {
			return nil, e
		}
		return ret, nil
	case 0x81:
		// lmsi, context_specific(80) + primitive(00) + 0(01)
		ret := RpLMSI{}
		if ret.LMSI, e = teldata.DecodeLMSI(v); e != nil {
			return nil, e
		}
		return ret, nil
	case 0x82:
		// msisdn, context_specific(80) + primitive(00) + 2(02)
		ret := RpMSISDN{}
		if ret.MSISDN, e = common.DecodeAddressString(v); e != nil {
			return nil, e
		}
		return ret, nil
	case 0x83:
		// roamingNumber, context_specific(80) + primitive(00) + 3(03)
		ret := RpRoamingNumber{}
		if ret.RoamingNum, e = common.DecodeAddressString(v); e != nil {
			return nil, e
		}
		return ret, nil
	case 0x84:
		// serviceCenterAddress, context_specific(80) + primitive(00) + 4(04)
		ret := RpSCAddress{}
		if ret.SCAddr, e = common.DecodeAddressString(v); e != nil {
			return nil, e
		}
		return ret, nil
	case 0x85:
		// noSM-RP-OA/DA, context_specific(80) + primitive(00) + 5(05)
		return nil, nil
	default:
		return nil, gsmap.UnexpectedTag([]byte{0x80, 0x81, 0x82, 0x83, 0x84, 0x85}, t)
	}
}

func marshalRPAddr(r RpAddr, buf *bytes.Buffer) {
	switch addr := r.(type) {
	case RpIMSI:
		gsmap.WriteTLV(buf, 0x80, addr.IMSI.Bytes())
	case RpLMSI:
		gsmap.WriteTLV(buf, 0x81, addr.LMSI.Bytes())
	case RpMSISDN:
		gsmap.WriteTLV(buf, 0x82, addr.MSISDN.Bytes())
	case RpRoamingNumber:
		gsmap.WriteTLV(buf, 0x83, addr.RoamingNum.Bytes())
	case RpSCAddress:
		gsmap.WriteTLV(buf, 0x84, addr.SCAddr.Bytes())
	default:
		gsmap.WriteTLV(buf, 0x85, nil)
	}
}

type RpIMSI struct {
	IMSI teldata.IMSI `json:"imsi"`
}

func (RpIMSI) RpAddrID() []byte {
	return []byte{0x80}
}

func (r RpIMSI) String() string {
	return "IMSI: " + r.IMSI.String()
}

type RpLMSI struct {
	LMSI teldata.LMSI `json:"lmsi"`
}

func (RpLMSI) RpAddrID() []byte {
	return []byte{0x81}
}

func (r RpLMSI) String() string {
	return "LMSI: " + r.LMSI.String()
}

type RpMSISDN struct {
	MSISDN common.AddressString `json:"msisdn"`
}

func (RpMSISDN) RpAddrID() []byte {
	return []byte{0x82}
}

func (r RpMSISDN) String() string {
	return "MSISDN: " + r.MSISDN.String()
}

type RpRoamingNumber struct {
	RoamingNum common.AddressString `json:"roamingNumber"`
}

func (RpRoamingNumber) RpAddrID() []byte {
	return []byte{0x83}
}

func (r RpRoamingNumber) String() string {
	return "RoamingNumber: " + r.RoamingNum.String()
}

type RpSCAddress struct {
	SCAddr common.AddressString `json:"serviceCentreAddress"`
}

func (RpSCAddress) RpAddrID() []byte {
	return []byte{0x84}
}

func (r RpSCAddress) String() string {
	return "SC-Address: " + r.SCAddr.String()
}
