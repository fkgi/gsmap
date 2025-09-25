package ife

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
	ShortMsgRelay1   gsmap.AppContext = 0x0004000001001501
	ShortMsgMORelay2 gsmap.AppContext = 0x0004000001001502
	ShortMsgMORelay3 gsmap.AppContext = 0x0004000001001503
)

/*
mo-ForwardSM  OPERATION ::= {
	ARGUMENT
		MO-ForwardSM-Arg
	RESULT
		MO-ForwardSM-Res -- optional
	ERRORS {
		systemFailure        |
		unexpectedDataValue  |
		facilityNotSupported |
		sm-DeliveryFailure }
	CODE local:46 }

ForwardSM ::= OPERATION -- Timer ml
	ARGUMENT
		forwardSM-Arg ForwardSM-Arg
	RESULT
	ERRORS {
		SystemFailure,
		DataMissing, -- DataMissing must not be used in version 1
		UnexpectedDataValue,
		FacilityNotSupported,
		UnidentifiedSubscriber,
		IllegalSubscriber,
		IllegalEquipment, -- IllegalEquipment must not be used in version 1
		AbsentSubscriber,
		SubscriberBusyForMT-SMS, -- SubscriberBusyForMT-SMS must not be used in version 1
		SM-DeliveryFailure }
*/

func init() {
	a := MOForwardSMArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := MOForwardSMRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
MOForwardSMArg operation arg.

	MO-ForwardSM-Arg ::= SEQUENCE {
		sm-RP-DA               SM-RP-DA,
		sm-RP-OA               SM-RP-OA,
		sm-RP-UI               SignalInfo,
		extensionContainer     ExtensionContainer OPTIONAL,
		... ,
		imsi                   IMSI               OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		correlationID      [0] CorrelationID      OPTIONAL,
		sm-DeliveryOutcome [1] SM-DeliveryOutcome OPTIONAL }

	SM-RP-DA ::= CHOICE {
		serviceCentreAddressDA [4] AddressString,
		noSM-RP-DA             [5] NULL }
	SM-RP-OA ::= CHOICE {
		msisdn                 [2] ISDN-AddressString,
		noSM-RP-OA             [5] NULL }

	ForwardSM-Arg ::= SEQUENCE {
		sm-RP-DA           SM-RP-DA,
		sm-RP-OA           SM-RP-OA,
		sm-RP-UI           SignalInfo,
		moreMessagesToSend NULL       OPTIONAL, // for MT
		... }

moreMessagesToSend must be absent in version 1.
*/
type MOForwardSMArg struct {
	InvokeID int8 `json:"id"`

	SMRPDA RpAddr             `json:"sm-RP-DA"`
	SMRPOA RpAddr             `json:"sm-RP-OA"`
	SMRPUI common.OctetString `json:"sm-RP-UI"`
	MMS    bool               `json:"moreMessagesToSend,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	IMSI teldata.IMSI `json:"imsi,omitempty"`
}

func (mo MOForwardSMArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", mo.Name(), mo.InvokeID)
	fmt.Fprintf(buf, "\n%ssm-RP-DA: %v", gsmap.LogPrefix, mo.SMRPDA)
	fmt.Fprintf(buf, "\n%ssm-RP-OA: %v", gsmap.LogPrefix, mo.SMRPOA)
	fmt.Fprintf(buf, "\n%ssm-RP-UI: %s", gsmap.LogPrefix, mo.SMRPUI)
	if mo.MMS {
		fmt.Fprintf(buf, "\n%smoreMessagesToSend:", gsmap.LogPrefix)
	}
	if !mo.IMSI.IsEmpty() {
		fmt.Fprintf(buf, "\n%simsi:     %s", gsmap.LogPrefix, mo.IMSI)
	}
	return buf.String()
}

func (mo MOForwardSMArg) GetInvokeID() int8             { return mo.InvokeID }
func (MOForwardSMArg) GetLinkedID() *int8               { return nil }
func (MOForwardSMArg) Code() byte                       { return 46 }
func (MOForwardSMArg) Name() string                     { return "MO-ForwardSM-Arg" }
func (MOForwardSMArg) DefaultContext() gsmap.AppContext { return ShortMsgRelay1 }

func (MOForwardSMArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8      `json:"id"`
		SMRPDA   rpAddrJSON `json:"sm-RP-DA"`
		SMRPOA   rpAddrJSON `json:"sm-RP-OA"`
		MOForwardSMArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.MOForwardSMArg, e
	}
	c := tmp.MOForwardSMArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	c.SMRPDA = tmp.SMRPDA.getRpAddr()
	c.SMRPOA = tmp.SMRPOA.getRpAddr()
	return c, nil
}

func (mo MOForwardSMArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// sm-RP-DA
	marshalRPAddr(mo.SMRPDA, buf)

	// sm-RP-OA
	marshalRPAddr(mo.SMRPOA, buf)

	// sm-RP-UI, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, mo.SMRPUI)

	// moreMessagesToSend, universal(00) + primitive(00) + null(05)
	if mo.MMS {
		gsmap.WriteTLV(buf, 0x05, nil)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// imsi, universal(00) + primitive(00) + octet_string(04)
	if len(mo.IMSI) != 0 {
		gsmap.WriteTLV(buf, 0x04, mo.IMSI.Bytes())
	}

	// MOForwardSM-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (MOForwardSMArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// MOForwardSM-Arg, universal(00) + constructed(20) + sequence(10)
	mo := MOForwardSMArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// sm-RP-DA
	var e error
	if mo.SMRPDA, e = unmarshalRPAddr(buf); e != nil {
		return nil, e
	}

	// sm-RP-OA
	if mo.SMRPOA, e = unmarshalRPAddr(buf); e != nil {
		return nil, e
	}

	// sm-RP-UI, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else {
		mo.SMRPUI = v
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return mo, nil
	} else if e != nil {
		return nil, e
	}

	// moreMessagesToSend, universal(00) + primitive(00) + null(05)
	if t == 0x05 {
		mo.MMS = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return mo, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return mo, nil
		} else if e != nil {
			return nil, e
		}
	}

	// imsi, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		if mo.IMSI, e = teldata.DecodeIMSI(v); e != nil {
			return nil, e
		}

		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return mo, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return mo, nil
}

/*
MOForwardSMRes operation res.

	MO-ForwardSM-Res ::= SEQUENCE {
		sm-RP-UI           SignalInfo         OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		...}
*/
type MOForwardSMRes struct {
	InvokeID int8 `json:"id"`

	SMRPUI common.OctetString `json:"sm-RP-UI,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (mo MOForwardSMRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", mo.Name(), mo.InvokeID)
	if len(mo.SMRPUI) != 0 {
		fmt.Fprintf(buf, "\n%ssm-RP-UI: %s", gsmap.LogPrefix, mo.SMRPUI)
	}
	return buf.String()
}

func (mo MOForwardSMRes) GetInvokeID() int8 { return mo.InvokeID }
func (MOForwardSMRes) Code() byte           { return 46 }
func (MOForwardSMRes) Name() string         { return "MO-ForwardSM-Res" }

func (MOForwardSMRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		MOForwardSMRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.MOForwardSMRes, e
	}
	c := tmp.MOForwardSMRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (mo MOForwardSMRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// sm-RP-UI, universal(00) + primitive(00) + octet_string(04)
	if len(mo.SMRPUI) != 0 {
		gsmap.WriteTLV(buf, 0x04, mo.SMRPUI)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// MOForwardSM-Res, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (MOForwardSMRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// MOForwardSM-Res, universal(00) + constructed(20) + sequence(10)
	mo := MOForwardSMRes{InvokeID: id}
	if buf.Len() == 0 {
		return mo, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return mo, nil
	} else if e != nil {
		return nil, e
	}

	// sm-RP-UI, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		mo.SMRPUI = v

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return mo, nil
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

	return mo, nil
}
