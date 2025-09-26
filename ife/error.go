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
SubscriberBusyForMT_SMS error operation.
subscriberBusyForMT-SMS must not be used in version 1.
If GprsConnectionSuspended is not understood it shall be discarded.

	subscriberBusyForMT-SMS  ERROR ::= {
		PARAMETER
			SubBusyForMT-SMS-Param -- optional
		CODE local:31 }

	SubBusyForMT-SMS-Param ::= SEQUENCE {
		extensionContainer      ExtensionContainer OPTIONAL,
		... ,
		gprsConnectionSuspended NULL               OPTIONAL }
*/
type SubscriberBusyForMT_SMS struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	GprsConnectionSuspended bool `json:"gprsConnectionSuspended,omitempty"`
}

func init() {
	c := SubscriberBusyForMT_SMS{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err SubscriberBusyForMT_SMS) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.GprsConnectionSuspended {
		fmt.Fprintf(buf, "\n%sgprsConnectionSuspended:", gsmap.LogPrefix)
	}
	return buf.String()
}

func (err SubscriberBusyForMT_SMS) GetInvokeID() int8 { return err.InvokeID }
func (SubscriberBusyForMT_SMS) Code() byte            { return 31 }
func (SubscriberBusyForMT_SMS) Name() string          { return "SubscriberBusyForMT-SMS" }

func (SubscriberBusyForMT_SMS) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		SubscriberBusyForMT_SMS
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.SubscriberBusyForMT_SMS, e
	}
	c := tmp.SubscriberBusyForMT_SMS
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err SubscriberBusyForMT_SMS) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// GprsConnectionSuspended, universal(00) + primitive(00) + null(05)
	if err.GprsConnectionSuspended {
		gsmap.WriteTLV(buf, 0x05, nil)
	}

	if buf.Len() != 0 {
		// SubBusyForMT-SMS-Param, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (SubscriberBusyForMT_SMS) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	err := SubscriberBusyForMT_SMS{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	}

	// SubBusyForMT-SMS-Param, universal(00) + constructed(20) + sequence(10)
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// GprsConnectionSuspended, universal(00) + primitive(00) + null(05)
	if t == 0x05 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedEnumValue(v)
		}
		err.GprsConnectionSuspended = true
	}

	return err, nil
}

/*
SM_DeliveryFailure error operation.
sm-DeliveryFailureCauseWithDiagnostic must not be used in version 1.
sm-EnumeratedDeliveryFailureCause must not be used in version greater 1.

	sm-DeliveryFailure  ERROR ::= {
		PARAMETER
			SM-DeliveryFailureCause
		CODE local:32 }

	SM-DeliveryFailureCause ::= CHOICE {
		sm-DeliveryFailureCauseWithDiagnostic SM-DeliveryFailureCauseWithDiagnostic,
		sm-EnumeratedDeliveryFailureCause     SM-EnumeratedDeliveryFailureCause }

	SM-DeliveryFailureCauseWithDiagnostic ::= SEQUENCE {
		sm-EnumeratedDeliveryFailureCause SM-EnumeratedDeliveryFailureCause,
		diagnosticInfo                    SignalInfo                         OPTIONAL,
		extensionContainer                ExtensionContainer                 OPTIONAL,
		... }
*/
type SM_DeliveryFailure struct {
	InvokeID int8 `json:"id"`

	NotExtensible bool                 `json:"notExtensible,omitempty"`
	Cause         DeliveryFailureCause `json:"sm-EnumeratedDeliveryFailureCause"`
	Diag          gsmap.OctetString    `json:"diagnosticInfo,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := SM_DeliveryFailure{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err SM_DeliveryFailure) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	fmt.Fprintf(buf, "\n%ssm-EnumeratedDeliveryFailureCause: %s",
		gsmap.LogPrefix, err.Cause)
	if len(err.Diag) != 0 {
		fmt.Fprintf(buf, "\n%sdiagnosticInfo: %s", gsmap.LogPrefix, err.Diag)
	}
	return buf.String()
}

func (err SM_DeliveryFailure) GetInvokeID() int8 { return err.InvokeID }
func (SM_DeliveryFailure) Code() byte            { return 32 }
func (SM_DeliveryFailure) Name() string          { return "SM-DeliveryFailure" }

func (SM_DeliveryFailure) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		SM_DeliveryFailure
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.SM_DeliveryFailure, e
	}
	c := tmp.SM_DeliveryFailure
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err SM_DeliveryFailure) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// sm-EnumeratedDeliveryFailureCause, universal(00) + primitive(00) + enum(0a)
	if tmp := err.Cause.marshal(); len(tmp) != 0 {
		gsmap.WriteTLV(buf, 0x0a, tmp)
	} else {
		gsmap.WriteTLV(buf, 0x0a, []byte{0x00})
	}

	if err.NotExtensible {
		return buf.Bytes()
	}

	// diagnosticInfo, universal(00) + primitive(00) + octet_string(04)
	if len(err.Diag) != 0 {
		gsmap.WriteTLV(buf, 0x04, err.Diag)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// SM-DeliveryFailureCauseWithDiagnostic, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (SM_DeliveryFailure) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	err := SM_DeliveryFailure{InvokeID: id}

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e != nil {
		return nil, e
	}
	// sm-EnumeratedDeliveryFailureCause, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Cause.unmarshal(v); e != nil {
			return nil, e
		}
		err.NotExtensible = true
		return err, nil
	}

	// SM-DeliveryFailureCauseWithDiagnostic, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		buf = bytes.NewBuffer(v)
	} else {
		return nil, gsmap.UnexpectedTag([]byte{0x30}, t)
	}

	// sm-EnumeratedDeliveryFailureCause, universal(00) + primitive(00) + enum(0a)
	if _, v, e := gsmap.ReadTLV(buf, 0x0a); e != nil {
		return nil, e
	} else if e = err.Cause.unmarshal(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e = gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	}

	// diagnosticInfo, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		err.Diag = v

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
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

	return err, nil
}
