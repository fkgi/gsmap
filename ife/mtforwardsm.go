package ife

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
mt-ForwardSM OPERATION ::= {
	ARGUMENT
		MT-ForwardSM-Arg
	RESULT
		MT-ForwardSM-Res -- optional
	ERRORS {
		systemFailure           |
		dataMissing             |
		unexpectedDataValue     |
		facilityNotSupported    |
		unidentifiedSubscriber  |
		illegalSubscriber       |
		illegalEquipment        |
		subscriberBusyForMT-SMS |
		sm-DeliveryFailure      |
		absentSubscriberSM }
	CODE local:44 }
*/

func init() {
	a := MTForwardSMArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := MTForwardSMRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
MTForwardSMArg operation arg.

	MT-ForwardSM-Arg ::= SEQUENCE {
		sm-RP-DA           SM-RP-DA,
		sm-RP-OA           SM-RP-OA,
		sm-RP-UI           SignalInfo,
		moreMessagesToSend NULL               OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		smDeliveryTimer               SM-DeliveryTimerValue      OPTIONAL,
		smDeliveryStartTime           Time                       OPTIONAL,
		smsOverIP-OnlyIndicator   [0] NULL                       OPTIONAL,
		correlationID             [1] CorrelationID              OPTIONAL,
		maximumRetransmissionTime [2] Time                       OPTIONAL,
		smsGmscAddress            [3] ISDN-AddressString         OPTIONAL,
		smsGmscDiameterAddress    [4] NetworkNodeDiameterAddress OPTIONAL }

	SignalInfo ::= OCTET STRING (SIZE (1..maxSignalInfoLength))
*/
type MTForwardSMArg struct {
	InvokeID int8 `json:"id"`

	SMRPDA RpAddr            `json:"sm-RP-DA"`
	SMRPOA RpAddr            `json:"sm-RP-OA"`
	SMRPUI gsmap.OctetString `json:"sm-RP-UI"`
	MMS    bool              `json:"moreMessagesToSend,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (mt MTForwardSMArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", mt.Name(), mt.InvokeID)
	fmt.Fprintf(buf, "\n%ssm-RP-DA: %v", gsmap.LogPrefix, mt.SMRPDA)
	fmt.Fprintf(buf, "\n%ssm-RP-OA: %v", gsmap.LogPrefix, mt.SMRPOA)
	fmt.Fprintf(buf, "\n%ssm-RP-UI: %s", gsmap.LogPrefix, mt.SMRPUI)
	if mt.MMS {
		fmt.Fprintf(buf, "\n%smoreMessagesToSend:", gsmap.LogPrefix)
	}
	return buf.String()
}

func (mt MTForwardSMArg) GetInvokeID() int8             { return mt.InvokeID }
func (MTForwardSMArg) GetLinkedID() *int8               { return nil }
func (MTForwardSMArg) Code() byte                       { return 44 }
func (MTForwardSMArg) Name() string                     { return "MT-ForwardSM-Arg" }
func (MTForwardSMArg) DefaultContext() gsmap.AppContext { return 0 }

func (MTForwardSMArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8      `json:"id"`
		SMRPDA   rpAddrJSON `json:"sm-RP-DA"`
		SMRPOA   rpAddrJSON `json:"sm-RP-OA"`
		MTForwardSMArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.MTForwardSMArg, e
	}
	c := tmp.MTForwardSMArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	c.SMRPDA = tmp.SMRPDA.getRpAddr()
	c.SMRPOA = tmp.SMRPOA.getRpAddr()
	return c, nil
}

func (mt MTForwardSMArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// sm-RP-DA
	marshalRPAddr(mt.SMRPDA, buf)

	// sm-RP-OA
	marshalRPAddr(mt.SMRPOA, buf)

	// sm-RP-UI, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, mt.SMRPUI)

	// moreMessagesToSend, universal(00) + primitive(00) + null(05)
	if mt.MMS {
		gsmap.WriteTLV(buf, 0x05, nil)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// MTForwardSM-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (MTForwardSMArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// MTForwardSM-Arg, universal(00) + constructed(20) + sequence(10)
	mt := MTForwardSMArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// sm-RP-DA
	var e error
	if mt.SMRPDA, e = unmarshalRPAddr(buf); e != nil {
		return nil, e
	}

	// sm-RP-OA
	if mt.SMRPOA, e = unmarshalRPAddr(buf); e != nil {
		return nil, e
	}

	// sm-RP-UI
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else {
		mt.SMRPUI = v
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return mt, nil
	} else if e != nil {
		return nil, e
	}

	// moreMessagesToSend, universal(00) + primitive(00) + null(05)
	if t == 0x05 {
		mt.MMS = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return mt, nil
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

	return mt, nil
}

/*
MTForwardSMRes operation res.

	MT-ForwardSM-Res ::= SEQUENCE {
		sm-RP-UI           SignalInfo         OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		... }
*/
type MTForwardSMRes struct {
	InvokeID int8 `json:"id"`

	SMRPUI gsmap.OctetString `json:"sm-RP-UI,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (mt MTForwardSMRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", mt.Name(), mt.InvokeID)
	if len(mt.SMRPUI) != 0 {
		fmt.Fprintf(buf, "\n%ssm-RP-UI: %s", gsmap.LogPrefix, mt.SMRPUI)
	}
	return buf.String()
}

func (mt MTForwardSMRes) GetInvokeID() int8 { return mt.InvokeID }
func (MTForwardSMRes) Code() byte           { return 44 }
func (MTForwardSMRes) Name() string         { return "MT-ForwardSM-Res" }

func (MTForwardSMRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		MTForwardSMRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.MTForwardSMRes, e
	}
	c := tmp.MTForwardSMRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (mt MTForwardSMRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// sm-RP-UI, universal(00) + primitive(00) + octet_string(04)
	if len(mt.SMRPUI) != 0 {
		gsmap.WriteTLV(buf, 0x04, mt.SMRPUI)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// MOForwardSM-Res, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (MTForwardSMRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// MTForwardSM-Res, universal(00) + constructed(20) + sequence(10)
	mt := MTForwardSMRes{InvokeID: id}
	if buf.Len() == 0 {
		return mt, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return mt, nil
	} else if e != nil {
		return nil, e
	}

	// sm-RP-UI, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		mt.SMRPUI = v
		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return mt, nil
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

	return mt, nil
}
