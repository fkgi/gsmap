package ifd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/teldata"
)

/*
readyForSM OPERATION ::= {
	ARGUMENT
		ReadyForSM-Arg
	RESULT
		ReadyForSM-Res -- optional
	ERRORS {
		dataMissing          |
		unexpectedDataValue  |
		facilityNotSupported |
		unknownSubscriber}
	CODE local:66 }
*/

func init() {
	a := ReadyForSmArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := ReadyForSmRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
ReadyForSmArg operation arg.

	ReadyForSM-Arg ::= SEQUENCE {
		imsi                 [0] IMSI,
		alertReason              AlertReason,
		alertReasonIndicator     NULL               OPTIONAL,
		extensionContainer       ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		additionalAlertReasonIndicator [1] NULL OPTIONAL,
		maximumUeAvailabilityTime          Time OPTIONAL }

alertReasonIndicator is set only when the alertReason sent to HLR is for GPRS.
*/
type ReadyForSmArg struct {
	InvokeID int8 `json:"id"`

	IMSI    teldata.IMSI `json:"imsi"`
	Reason  alertReason  `json:"alertReason"`
	ForGPRS bool         `json:"alertReasonIndicator,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (rsm ReadyForSmArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", rsm.Name(), rsm.InvokeID)
	fmt.Fprintf(buf, "\n%simsi:        %s", gsmap.LogPrefix, rsm.IMSI)
	fmt.Fprintf(buf, "\n%salertReason: %s", gsmap.LogPrefix, rsm.Reason)
	if rsm.ForGPRS {
		fmt.Fprintf(buf, "\n%salertReasonIndicator:", gsmap.LogPrefix)
	}
	// Extension
	return buf.String()
}

func (rsm ReadyForSmArg) GetInvokeID() int8            { return rsm.InvokeID }
func (ReadyForSmArg) GetLinkedID() *int8               { return nil }
func (ReadyForSmArg) Code() byte                       { return 66 }
func (ReadyForSmArg) Name() string                     { return "ReadyForSM-Arg" }
func (ReadyForSmArg) DefaultContext() gsmap.AppContext { return MwdMngt1 }

func (ReadyForSmArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		ReadyForSmArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.ReadyForSmArg, e
	}
	c := tmp.ReadyForSmArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (rsm ReadyForSmArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// imsi, context_specific(80) + primitive(00) + 0(00)
	gsmap.WriteTLV(buf, 0x80, rsm.IMSI.Bytes())

	// alertReason, universal(00) + primitive(00) + enum(0a)
	if b := rsm.Reason.marshal(); b != nil {
		gsmap.WriteTLV(buf, 0x0a, b)
	} else {
		gsmap.WriteTLV(buf, 0x0a, []byte{0x00})
	}

	// alertReasonIndicator, universal(00) + primitive(00) + null(05)
	if rsm.ForGPRS {
		gsmap.WriteTLV(buf, 0x05, nil)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// ReadyForSM-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (ReadyForSmArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// ReadyForSM-Arg, universal(00) + constructed(20) + sequence(10)
	rsm := ReadyForSmArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// imsi, context_specific(80) + primitive(00) + 0(00)
	if _, v, e := gsmap.ReadTLV(buf, 0x80); e != nil {
		return nil, e
	} else if rsm.IMSI, e = teldata.DecodeIMSI(v); e != nil {
		return nil, e
	}

	// alertReason, universal(00) + primitive(00) + enum(0a)
	if _, v, e := gsmap.ReadTLV(buf, 0x0a); e != nil {
		return nil, e
	} else if e = rsm.Reason.unmarshal(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return rsm, nil
	} else if e != nil {
		return nil, e
	}

	// alertReasonIndicator, universal(00) + primitive(00) + null(05)
	if t == 0x05 {
		rsm.ForGPRS = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rsm, nil
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
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return rsm, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return rsm, nil
}

/*
ReadyForSmRes operation res.

	ReadyForSM-Res ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		...}
*/
type ReadyForSmRes struct {
	InvokeID int8 `json:"id"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (rsm ReadyForSmRes) String() string {
	return fmt.Sprintf("%s (ID=%d)", rsm.Name(), rsm.InvokeID)
}

func (rsm ReadyForSmRes) GetInvokeID() int8 { return rsm.InvokeID }
func (ReadyForSmRes) Code() byte            { return 66 }
func (ReadyForSmRes) Name() string          { return "ReadyForSM-Res" }

func (ReadyForSmRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		ReadyForSmRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.ReadyForSmRes, e
	}
	c := tmp.ReadyForSmRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (rsm ReadyForSmRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// ReadyForSM-Res, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (ReadyForSmRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// ReadyForSM-Res, universal(00) + constructed(20) + sequence(10)
	rsm := ReadyForSmRes{InvokeID: id}
	if buf.Len() == 0 {
		return rsm, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return rsm, nil
	} else if e != nil {
		return nil, e
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}

	return rsm, nil
}

/*
alertReason

	AlertReason ::= ENUMERATED {
		ms-Present       (0),
		memoryAvailable  (1)}
*/
type alertReason byte

const (
	_ alertReason = iota
	SMPresent
	MemoryAvailable
)

func (a alertReason) String() string {
	switch a {
	case SMPresent:
		return "ms-Present"
	case MemoryAvailable:
		return "memoryAvailable"
	}
	return ""
}

func (a *alertReason) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "ms-Present":
		*a = SMPresent
	case "memoryAvailable":
		*a = MemoryAvailable
	default:
		*a = 0
	}
	return
}

func (a alertReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *alertReason) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*a = SMPresent
	case 1:
		*a = MemoryAvailable
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (a alertReason) marshal() []byte {
	switch a {
	case SMPresent:
		return []byte{0x00}
	case MemoryAvailable:
		return []byte{0x01}
	}
	return nil
}
