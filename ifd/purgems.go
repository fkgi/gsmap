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
purgeMS  OPERATION ::= {	--Timer m
	ARGUMENT
		PurgeMS-Arg
	RESULT
		PurgeMS-Res -- optional
	ERRORS{
		dataMissing         |
		unexpectedDataValue |
		unknownSubscriber   }
	CODE local:67 }
*/

func init() {
	a := PurgeMSArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := PurgeMSRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
PurgeMSArg

	PurgeMS-Arg ::= [3] SEQUENCE {
		imsi                   IMSI,
		vlr-Number         [0] ISDN-AddressString OPTIONAL,
		sgsn-Number        [1] ISDN-AddressString OPTIONAL,
		extensionContainer     ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^^^ R99 ^^^^^^^^
		locationInformation     [2] LocationInformation     OPTIONAL,
		locationInformationGPRS [3] LocationInformationGPRS OPTIONAL,
		locationInformationEPS  [4] LocationInformationEPS  OPTIONAL }
*/
type PurgeMSArg struct {
	InvokeID int8 `json:"id"`

	IMSI       teldata.IMSI        `json:"imsi"`
	VlrNumber  gsmap.AddressString `json:"vlr-Number,omitempty"`
	SgsnNumber gsmap.AddressString `json:"sgsn-Number,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (pm PurgeMSArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", pm.Name(), pm.InvokeID)
	fmt.Fprintf(buf, "\n%simsi: %s", gsmap.LogPrefix, pm.IMSI)
	if !pm.VlrNumber.IsEmpty() {
		fmt.Fprintf(buf, "\n%svlr-Number: %s", gsmap.LogPrefix, pm.VlrNumber)
	}
	if !pm.SgsnNumber.IsEmpty() {
		fmt.Fprintf(buf, "\n%ssgsn-Number: %s", gsmap.LogPrefix, pm.SgsnNumber)
	}
	// Extension
	return buf.String()
}

func (pm PurgeMSArg) MarshalJSON() ([]byte, error) {
	j := struct {
		InvokeID int8 `json:"id"`

		IMSI       teldata.IMSI         `json:"imsi"`
		VlrNumber  *gsmap.AddressString `json:"vlr-Number,omitempty"`
		SgsnNumber *gsmap.AddressString `json:"sgsn-Number,omitempty"`
	}{
		InvokeID: pm.InvokeID,
		IMSI:     pm.IMSI}
	if !pm.VlrNumber.IsEmpty() {
		j.VlrNumber = &pm.VlrNumber
	}
	if !pm.SgsnNumber.IsEmpty() {
		j.SgsnNumber = &pm.SgsnNumber
	}
	return json.Marshal(j)
}

func (pm PurgeMSArg) GetInvokeID() int8             { return pm.InvokeID }
func (PurgeMSArg) GetLinkedID() *int8               { return nil }
func (PurgeMSArg) Code() byte                       { return 67 }
func (PurgeMSArg) Name() string                     { return "PurgeMS-Arg" }
func (PurgeMSArg) DefaultContext() gsmap.AppContext { return 0 }

func (PurgeMSArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		PurgeMSArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.PurgeMSArg, e
	}
	c := tmp.PurgeMSArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (pm PurgeMSArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// imsi, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, pm.IMSI.Bytes())

	// vlr-Number, context_specific(80) + primitive(00) + 0(00)
	if !pm.VlrNumber.IsEmpty() {
		gsmap.WriteTLV(buf, 0x80, pm.VlrNumber.Bytes())
	}

	// sgsn-Number, context_specific(80) + primitive(00) + 1(01)
	if !pm.SgsnNumber.IsEmpty() {
		gsmap.WriteTLV(buf, 0x81, pm.SgsnNumber.Bytes())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// PurgeMS-Arg, context_specific(80) + constructed(20) + 3(03)
	return gsmap.WriteTLV(new(bytes.Buffer), 0xa3, buf.Bytes())
}

func (PurgeMSArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// PurgeMS-Arg, context_specific(80) + constructed(20) + 3(03)
	pm := PurgeMSArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0xa3); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// imsi, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if pm.IMSI, e = teldata.DecodeIMSI(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return pm, nil
	} else if e != nil {
		return nil, e
	}

	// vlr-Number, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		if pm.VlrNumber, e = gsmap.DecodeAddressString(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return pm, nil
		} else if e != nil {
			return nil, e
		}
	}

	// sgsn-Number, context_specific(80) + primitive(00) + 1(01)
	if t == 0x81 {
		if pm.SgsnNumber, e = gsmap.DecodeAddressString(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return pm, nil
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

	return pm, nil
}

/*
PurgeMSRes

	PurgeMS-Res ::= SEQUENCE {
		freezeTMSI         [0] NULL               OPTIONAL,
		freezeP-TMSI       [1] NULL               OPTIONAL,
		extensionContainer     ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^^^ R99 ^^^^^^^^
		freezeM-TMSI [2] NULL OPTIONAL }
*/
type PurgeMSRes struct {
	InvokeID int8 `json:"id"`

	FreezeTMSI  bool `json:"freezeTMSI,omitempty"`
	FreezePTMSI bool `json:"freezeP-TMSI,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (pm PurgeMSRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", pm.Name(), pm.InvokeID)
	if pm.FreezeTMSI {
		fmt.Fprintf(buf, "\n%sreezeTMSI:", gsmap.LogPrefix)
	}
	if pm.FreezePTMSI {
		fmt.Fprintf(buf, "\n%sreezeP-TMSI:", gsmap.LogPrefix)
	}
	return buf.String()
}

func (pm PurgeMSRes) GetInvokeID() int8             { return pm.InvokeID }
func (PurgeMSRes) GetLinkedID() *int8               { return nil }
func (PurgeMSRes) Code() byte                       { return 67 }
func (PurgeMSRes) Name() string                     { return "PurgeMS-Res" }
func (PurgeMSRes) DefaultContext() gsmap.AppContext { return 0 }

func (PurgeMSRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		PurgeMSRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.PurgeMSRes, e
	}
	c := tmp.PurgeMSRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (pm PurgeMSRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// freezeTMSI, context_specific(80) + primitive(00) + 0(00)
	if pm.FreezeTMSI {
		gsmap.WriteTLV(buf, 0x80, nil)
	}

	// freezeP-TMSI, context_specific(80) + primitive(00) + 1(01)
	if pm.FreezePTMSI {
		gsmap.WriteTLV(buf, 0x81, nil)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// PurgeMS-Res, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (PurgeMSRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// PurgeMS-Res, universal(00) + constructed(20) + sequence(10)
	pm := PurgeMSRes{InvokeID: id}
	if buf.Len() == 0 {
		return pm, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return pm, nil
	} else if e != nil {
		return nil, e
	}

	// freezeTMSI, context_specific(80) + primitive(00) + 0(00)
	if t == 0x80 {
		pm.FreezeTMSI = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return pm, nil
		} else if e != nil {
			return nil, e
		}
	}

	// freezeP-TMSI, context_specific(80) + primitive(00) + 1(01)
	if t == 0x81 {
		pm.FreezePTMSI = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return pm, nil
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

	return pm, nil
}
