package ifc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/common"
)

const (
	ShortMsgAlert1 gsmap.AppContext = 0x0004000001001701
	ShortMsgAlert2 gsmap.AppContext = 0x0004000001001702
)

/*
alertServiceCentreWithoutResult must not be used in version greater 1.

alertServiceCentreWithoutResult OPERATION ::= {
	ARGUMENT
    	AlertServiceCentreArg
	CODE	local:49 }
*/

func init() {
	c := AlertServiceCentreWithoutResult{}
	gsmap.ArgMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

/*
AlertServiceCentreWithoutResult operation arg.

		AlertServiceCentreArg ::= SEQUENCE {
			msisdn                        ISDN-AddressString,
			serviceCentreAddress          AddressString,
	     	... }
*/
type AlertServiceCentreWithoutResult struct {
	InvokeID int8 `json:"id"`

	MSISDN     common.AddressString `json:"msisdn"`
	CenterAddr common.AddressString `json:"serviceCentreAddress"`
}

func (al AlertServiceCentreWithoutResult) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "AlertServiceCentre-Arg (ID=%d)", al.InvokeID)
	fmt.Fprintf(buf, "\n%smsisdn:               %s", gsmap.LogPrefix, al.MSISDN)
	fmt.Fprintf(buf, "\n%sserviceCentreAddress: %s", gsmap.LogPrefix, al.CenterAddr)
	return buf.String()
}

func (al AlertServiceCentreWithoutResult) GetInvokeID() int8             { return al.InvokeID }
func (AlertServiceCentreWithoutResult) GetLinkedID() *int8               { return nil }
func (AlertServiceCentreWithoutResult) Code() byte                       { return 49 }
func (AlertServiceCentreWithoutResult) Name() string                     { return "AlertServiceCentreWithoutResult" }
func (AlertServiceCentreWithoutResult) DefaultContext() gsmap.AppContext { return ShortMsgAlert1 }

func (AlertServiceCentreWithoutResult) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		AlertServiceCentreWithoutResult
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.AlertServiceCentreWithoutResult, e
	}
	c := tmp.AlertServiceCentreWithoutResult
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (al AlertServiceCentreWithoutResult) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, al.MSISDN.Bytes())

	// serviceCentreAddress, universal(00) + primitive(00) + octet_string(04)
	b := gsmap.WriteTLV(buf, 0x04, al.CenterAddr.Bytes())

	// AlertServiceCentreArg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, b)
}

func (AlertServiceCentreWithoutResult) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// AlertServiceCentreArg, universal(00) + constructed(20) + sequence(10)
	al := AlertServiceCentreWithoutResult{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if al.MSISDN, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// serviceCentreAddress, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if al.CenterAddr, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	return al, nil
}

/*
alertServiceCentremust not be used in version 1.

alertServiceCentre  OPERATION ::= {	--Timer s
	ARGUMENT
		AlertServiceCentreArg
	RETURN RESULT TRUE
	ERRORS {
		systemFailure      |
		dataMissing        |
		unexpectedDataValue }
	CODE	local:64 }
*/

func init() {
	c := AlertServiceCentreArg{}
	gsmap.ArgMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

/*
AlertServiceCentreArg operation arg.

	AlertServiceCentreArg ::= SEQUENCE {
		msisdn                        ISDN-AddressString,
		serviceCentreAddress          AddressString,
		...,
		-- ^^^^^^ R99 ^^^^^^
		imsi                          IMSI                       OPTIONAL,
		correlationID                 CorrelationID              OPTIONAL,
		maximumUeAvailabilityTime [0] Time                       OPTIONAL,
		smsGmscAlertEvent         [1] SmsGmsc-Alert-Event        OPTIONAL,
		smsGmscDiameterAddress    [2] NetworkNodeDiameterAddress OPTIONAL,
		newSGSNNumber             [3] ISDN-AddressString         OPTIONAL,
		newSGSNDiameterAddress    [4] NetworkNodeDiameterAddress OPTIONAL,
		newMMENumber              [5] ISDN-AddressString         OPTIONAL,
		newMMEDiameterAddress     [6] NetworkNodeDiameterAddress OPTIONAL,
		newMSCNumber              [7] ISDN-AddressString         OPTIONAL }
*/
type AlertServiceCentreArg struct {
	InvokeID int8 `json:"id"`

	MSISDN     common.AddressString `json:"msisdn"`
	CenterAddr common.AddressString `json:"serviceCentreAddress"`
}

func (al AlertServiceCentreArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", al.Name(), al.InvokeID)
	fmt.Fprintf(buf, "\n%smsisdn:               %s", gsmap.LogPrefix, al.MSISDN)
	fmt.Fprintf(buf, "\n%sserviceCentreAddress: %s", gsmap.LogPrefix, al.CenterAddr)
	return buf.String()
}

func (al AlertServiceCentreArg) GetInvokeID() int8             { return al.InvokeID }
func (AlertServiceCentreArg) GetLinkedID() *int8               { return nil }
func (AlertServiceCentreArg) Code() byte                       { return 64 }
func (AlertServiceCentreArg) Name() string                     { return "AlertServiceCentre-Arg" }
func (AlertServiceCentreArg) DefaultContext() gsmap.AppContext { return 0 }

func (AlertServiceCentreArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		AlertServiceCentreArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.AlertServiceCentreArg, e
	}
	c := tmp.AlertServiceCentreArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (al AlertServiceCentreArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, al.MSISDN.Bytes())

	// serviceCentreAddress, universal(00) + primitive(00) + octet_string(04)
	b := gsmap.WriteTLV(buf, 0x04, al.CenterAddr.Bytes())

	// AlertServiceCentreArg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, b)
}

func (AlertServiceCentreArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// AlertServiceCentreArg, universal(00) + constructed(20) + sequence(10)
	al := AlertServiceCentreArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if al.MSISDN, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// serviceCentreAddress, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if al.CenterAddr, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	return al, nil
}
