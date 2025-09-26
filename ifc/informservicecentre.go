package ifc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
informServiceCentre  OPERATION ::= {	--Timer s
	ARGUMENT
		InformServiceCentreArg
	CODE	local:63 }
*/

func init() {
	a := InformServiceCentreArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a
}

/*
InformServiceCentreArg operation arg.
InformServiceCentre must not be used in version 1.

	InformServiceCentreArg ::= SEQUENCE {
		storedMSISDN       ISDN-AddressString OPTIONAL,
		mw-Status          MW-Status          OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		... ,
		-- ^^^^^^^^ R99 ^^^^^^^^
		absentSubscriberDiagnosticSM                AbsentSubscriberDiagnosticSM OPTIONAL,
		additionalAbsentSubscriberDiagnosticSM  [0] AbsentSubscriberDiagnosticSM OPTIONAL,
		smsf3gppAbsentSubscriberDiagnosticSM    [1] AbsentSubscriberDiagnosticSM OPTIONAL,
		smsfNon3gppAbsentSubscriberDiagnosticSM [2] AbsentSubscriberDiagnosticSM OPTIONAL }
*/
type InformServiceCentreArg struct {
	InvokeID int8 `json:"id"`

	MSISDN   gsmap.AddressString `json:"storedMSISDN,omitempty"`
	MWStatus MWStatus            `json:"mw-Status,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (isc InformServiceCentreArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", isc.Name(), isc.InvokeID)
	if !isc.MSISDN.IsEmpty() {
		fmt.Fprintf(buf, "\n%smsisdn:    %s", gsmap.LogPrefix, isc.MSISDN)
	}
	if isc.MWStatus != 0 {
		fmt.Fprintf(buf, "\n%sMW-Status: %s", gsmap.LogPrefix, isc.MWStatus)
	}
	return buf.String()
}

func (isc InformServiceCentreArg) MarshalJSON() ([]byte, error) {
	j := struct {
		InvokeID int8                 `json:"id"`
		MSISDN   *gsmap.AddressString `json:"storedMSISDN,omitempty"`
		MWStatus MWStatus             `json:"mw-Status,omitempty"`
		// Extension
	}{
		InvokeID: isc.InvokeID,
		MWStatus: isc.MWStatus}
	if !isc.MSISDN.IsEmpty() {
		j.MSISDN = &isc.MSISDN
	}
	return json.Marshal(j)
}

func (isc InformServiceCentreArg) GetInvokeID() int8            { return isc.InvokeID }
func (InformServiceCentreArg) GetLinkedID() *int8               { return nil }
func (InformServiceCentreArg) Code() byte                       { return 63 }
func (InformServiceCentreArg) Name() string                     { return "InformServiceCentre-Arg" }
func (InformServiceCentreArg) DefaultContext() gsmap.AppContext { return 0 }

func (InformServiceCentreArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		InformServiceCentreArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.InformServiceCentreArg, e
	}
	c := tmp.InformServiceCentreArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (isc InformServiceCentreArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// storedMSISDN, universal(00) + primitive(00) + octet_string(04)
	if !isc.MSISDN.IsEmpty() {
		gsmap.WriteTLV(buf, 0x04, isc.MSISDN.Bytes())
	}

	// mw-Status, universal(00) + primitive(00) + bit_string(03)
	if isc.MWStatus != 0 {
		gsmap.WriteTLV(buf, 0x03, isc.MWStatus.marshal())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// InformServiceCentre-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (InformServiceCentreArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// InformServiceCentre-Arg, universal(00) + constructed(20) + sequence(10)
	isc := InformServiceCentreArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return isc, nil
	} else if e != nil {
		return nil, e
	}

	// msisdn, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		if isc.MSISDN, e = gsmap.DecodeAddressString(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isc, nil
		} else if e != nil {
			return nil, e
		}
	}

	// mw-Status, universal(00) + primitive(00) + bit_string(03)
	if t == 0x03 {
		if e = isc.MWStatus.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return isc, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return nil, e
		}
		/*
			if t, v, e = gsmap.ReadTLV(buf, nil); e == io.EOF {
				return isc, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return isc, nil
}
