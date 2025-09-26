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
restoreData  OPERATION ::= {	--Timer m
	ARGUMENT
		RestoreDataArg
	RESULT
		RestoreDataRes
	ERRORS {
		systemFailure       |
		dataMissing         |
		unexpectedDataValue |
		unknownSubscriber   }
	CODE	local:57 }
*/

func init() {
	a := RestoreDataArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := RestoreDataRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
RestoreDataArg

	RestoreDataArg ::= SEQUENCE {
		imsi               IMSI,
		lmsi               LMSI               OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		... ,
		vlr-Capability       [6] VLR-Capability OPTIONAL,
		-- ^^^^^^^^ R99 ^^^^^^^^
		restorationIndicator [7] NULL           OPTIONAL }
*/
type RestoreDataArg struct {
	InvokeID int8 `json:"id"`

	IMSI teldata.IMSI `json:"imsi"`
	LMSI teldata.LMSI `json:"lmsi,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	VlrCapability vlrCapability `json:"vlr-Capability,omitempty"`
}

func (ul RestoreDataArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", ul.Name(), ul.InvokeID)
	fmt.Fprintf(buf, "\n%simsi:       %s", gsmap.LogPrefix, ul.IMSI)
	if !ul.LMSI.IsEmpty() {
		fmt.Fprintf(buf, "\n%slmsi:        %x", gsmap.LogPrefix, ul.LMSI.Bytes())
	}
	// Extension
	if s := ul.VlrCapability.String(); s != "" {
		fmt.Fprintf(buf, "\n%svlr-Capability:%s", gsmap.LogPrefix, s)
	}
	return buf.String()
}

func (rd RestoreDataArg) MarshalJSON() ([]byte, error) {
	j := struct {
		InvokeID int8 `json:"id"`

		IMSI          teldata.IMSI   `json:"imsi"`
		LMSI          *teldata.LMSI  `json:"lmsi,omitempty"`
		VlrCapability *vlrCapability `json:"vlr-Capability,omitempty"`
	}{
		InvokeID: rd.InvokeID,
		IMSI:     rd.IMSI}
	if !rd.LMSI.IsEmpty() {
		j.LMSI = &rd.LMSI
	}
	if rd.VlrCapability.String() != "" {
		j.VlrCapability = &rd.VlrCapability
	}
	return json.Marshal(j)
}

func (rd RestoreDataArg) GetInvokeID() int8             { return rd.InvokeID }
func (RestoreDataArg) GetLinkedID() *int8               { return nil }
func (RestoreDataArg) Code() byte                       { return 57 }
func (RestoreDataArg) Name() string                     { return "RestoreData-Arg" }
func (RestoreDataArg) DefaultContext() gsmap.AppContext { return NetworkLocUp1 }

func (RestoreDataArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		RestoreDataArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.RestoreDataArg, e
	}
	c := tmp.RestoreDataArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (rd RestoreDataArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// imsi, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, rd.IMSI.Bytes())

	// lmsi, universal(00) + primitive(00) + octet_string(04)
	if !rd.LMSI.IsEmpty() {
		gsmap.WriteTLV(buf, 0x04, rd.LMSI.Bytes())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// vlr-Capability, context_specific(80) + constructed(20) + 6(06)
	if tmp := rd.VlrCapability.marshal(); len(tmp) != 0 {
		gsmap.WriteTLV(buf, 0xa6, tmp)
	}

	// RestoreData-Res, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (RestoreDataArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// RestoreData-Res, universal(00) + constructed(20) + sequence(10)
	ul := RestoreDataArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// imsi, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if ul.IMSI, e = teldata.DecodeIMSI(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return ul, nil
	} else if e != nil {
		return nil, e
	}

	// lmsi, universal(00) + primitive(00) + octet_string(04)
	if t == 0x04 {
		if ul.LMSI, e = teldata.DecodeLMSI(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return ul, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return ul, nil
		} else if e != nil {
			return nil, e
		}
	}

	// vlr-Capability, context_specific(80) + constructed(20) + 6(06)
	if t == 0xa6 {
		if e = ul.VlrCapability.unmarshal(v); e != nil {
			return nil, e
		}

		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return ul, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return ul, nil
}

/*
RestoreDataRes

	RestoreDataRes ::= SEQUENCE {
		hlr-Number         ISDN-AddressString,
		msNotReachable     NULL               OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		...}
*/
type RestoreDataRes struct {
	InvokeID int8 `json:"id"`

	HlrNumber      gsmap.AddressString `json:"hlr-Number"`
	MsNotReachable bool                `json:"msNotReachable,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (rd RestoreDataRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", rd.Name(), rd.InvokeID)
	fmt.Fprintf(buf, "\n%shlr-Number: %s", gsmap.LogPrefix, rd.HlrNumber)
	if rd.MsNotReachable {
		fmt.Fprintf(buf, "\n%smsNotReachable:", gsmap.LogPrefix)
	}
	return buf.String()
}

func (rd RestoreDataRes) GetInvokeID() int8 { return rd.InvokeID }
func (RestoreDataRes) Code() byte           { return 57 }
func (RestoreDataRes) Name() string         { return "RestoreData-Res" }

func (RestoreDataRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		RestoreDataRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.RestoreDataRes, e
	}
	c := tmp.RestoreDataRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (rd RestoreDataRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// hlr-Number, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, rd.HlrNumber.Bytes())

	// msNotReachable, universal(00) + primitive(00) + null(05)
	if rd.MsNotReachable {
		gsmap.WriteTLV(buf, 0x05, nil)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// RestoreData-Res, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (RestoreDataRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// RestoreData-Res, universal(00) + constructed(20) + sequence(10)
	rd := RestoreDataRes{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// hlr-Number, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if rd.HlrNumber, e = gsmap.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return rd, nil
	} else if e != nil {
		return nil, e
	}

	// msNotReachable, universal(00) + primitive(00) + null(05)
	if t == 0x05 {
		rd.MsNotReachable = true

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return rd, nil
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

	return rd, nil
}
