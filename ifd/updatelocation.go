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
updateLocation OPERATION ::= {
	ARGUMENT
		UpdateLocationArg
	RESULT
		UpdateLocationRes
	ERRORS {
		systemFailure       |
		dataMissing         |
		unexpectedDataValue |
		unknownSubscriber   |
		roamingNotAllowed }
	CODE local:2 }
*/

func init() {
	a := UpdateLocationArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := UpdateLocationRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
UpdateLocationArg operation arg.

	UpdateLocationArg ::= SEQUENCE {
		imsi                             IMSI,
		msc-Number                  [1]  ISDN-AddressString,
		vlr-Number                       ISDN-AddressString,
		lmsi                        [10] LMSI               OPTIONAL,
		extensionContainer               ExtensionContainer OPTIONAL,
		... ,
		vlr-Capability              [6]  VLR-Capability     OPTIONAL,
		-- ^^^^^^^^ R99 ^^^^^^^^
		informPreviousNetworkEntity [11] NULL               OPTIONAL,
		cs-LCS-NotSupportedByUE     [12] NULL               OPTIONAL,
		v-gmlc-Address              [2]  GSN-Address        OPTIONAL,
		add-info                    [13] ADD-Info           OPTIONAL,
		pagingArea                  [14] PagingArea         OPTIONAL,
		skipSubscriberDataUpdate    [15] NULL               OPTIONAL,
		-- The skipSubscriberDataUpdate parameter in the UpdateLocationArg
		-- and the ADD-Info structures carry the same semantic.
		restorationIndicator [16] NULL                       OPTIONAL,
		eplmn-List           [3]  EPLMN-List                 OPTIONAL,
		mme-DiameterAddress  [4]  NetworkNodeDiameterAddress OPTIONAL
	}
*/
type UpdateLocationArg struct {
	InvokeID int8 `json:"id"`

	IMSI      teldata.IMSI        `json:"imsi"`
	MscNumber gsmap.AddressString `json:"msc-Number"`
	VlrNumber gsmap.AddressString `json:"vlr-Number"`
	LMSI      teldata.LMSI        `json:"lmsi,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	VlrCapability vlrCapability `json:"vlr-Capability,omitempty"`
}

func (ul UpdateLocationArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", ul.Name(), ul.InvokeID)
	fmt.Fprintf(buf, "\n%simsi:       %s", gsmap.LogPrefix, ul.IMSI)
	fmt.Fprintf(buf, "\n%smsc-Number: %s", gsmap.LogPrefix, ul.MscNumber)
	fmt.Fprintf(buf, "\n%svlr-Number: %s", gsmap.LogPrefix, ul.VlrNumber)
	if !ul.LMSI.IsEmpty() {
		fmt.Fprintf(buf, "\n%slmsi:        %x", gsmap.LogPrefix, ul.LMSI.Bytes())
	}
	// Extension
	if s := ul.VlrCapability.String(); s != "" {
		fmt.Fprintf(buf, "\n%svlr-Capability:%s", gsmap.LogPrefix, s)
	}
	return buf.String()
}

func (ul UpdateLocationArg) MarshalJSON() ([]byte, error) {
	j := struct {
		InvokeID int8 `json:"id"`

		IMSI          teldata.IMSI        `json:"imsi"`
		MscNumber     gsmap.AddressString `json:"msc-Number"`
		VlrNumber     gsmap.AddressString `json:"vlr-Number"`
		LMSI          *teldata.LMSI       `json:"lmsi,omitempty"`
		VlrCapability *vlrCapability      `json:"vlr-Capability,omitempty"`
	}{
		InvokeID:  ul.InvokeID,
		IMSI:      ul.IMSI,
		MscNumber: ul.MscNumber,
		VlrNumber: ul.VlrNumber,
	}
	if !ul.LMSI.IsEmpty() {
		j.LMSI = &ul.LMSI
	}
	if ul.VlrCapability.String() != "" {
		j.VlrCapability = &ul.VlrCapability
	}
	return json.Marshal(j)
}

func (ul UpdateLocationArg) GetInvokeID() int8             { return ul.InvokeID }
func (UpdateLocationArg) GetLinkedID() *int8               { return nil }
func (UpdateLocationArg) Code() byte                       { return 2 }
func (UpdateLocationArg) Name() string                     { return "UpdateLocation-Arg" }
func (UpdateLocationArg) DefaultContext() gsmap.AppContext { return NetworkLocUp1 }

func (UpdateLocationArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		UpdateLocationArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.UpdateLocationArg, e
	}
	c := tmp.UpdateLocationArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (ul UpdateLocationArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// imsi, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, ul.IMSI.Bytes())

	// msc-Number, context_specific(80) + primitive(00) + 1(01)
	gsmap.WriteTLV(buf, 0x81, ul.MscNumber.Bytes())

	// vlr-Number, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, ul.VlrNumber.Bytes())

	// lmsi, context_specific(80) + primitive(00) + 10(0a)
	if !ul.LMSI.IsEmpty() {
		gsmap.WriteTLV(buf, 0x8a, ul.LMSI.Bytes())
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// vlr-Capability, context_specific(80) + constructed(20) + 1(01)
	if tmp := ul.VlrCapability.marshal(); len(tmp) != 0 {
		gsmap.WriteTLV(buf, 0xa1, tmp)
	}

	// UpdateLocation-Res, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (UpdateLocationArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// UpdateLocation-Res, universal(00) + constructed(20) + sequence(10)
	ul := UpdateLocationArg{InvokeID: id}
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

	// msc-Number, context_specific(80) + primitive(00) + 1(01)
	if _, v, e := gsmap.ReadTLV(buf, 0x81); e != nil {
		return nil, e
	} else if ul.MscNumber, e = gsmap.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// vlr-Number, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if ul.VlrNumber, e = gsmap.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return ul, nil
	} else if e != nil {
		return nil, e
	}

	// lmsi, context_specific(80) + primitive(00) + 10(0a)
	if t == 0xa1 {
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

	// vlr-Capability, context_specific(80) + constructed(20) + 1(01)
	if t == 0xa1 {
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
UpdateLocationRes operation res.

	UpdateLocationRes ::= SEQUENCE {
		hlr-Number         ISDN-AddressString,
		extensionContainer ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		add-Capability            NULL OPTIONAL,
		pagingArea-Capability [0] NULL OPTIONAL }
*/
type UpdateLocationRes struct {
	InvokeID int8 `json:"id"`

	HlrNumber gsmap.AddressString `json:"hlr-Number"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (ul UpdateLocationRes) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", ul.Name(), ul.InvokeID)
	fmt.Fprintf(buf, "\n%shlr-Number: %s", gsmap.LogPrefix, ul.HlrNumber)
	return buf.String()
}

func (ul UpdateLocationRes) GetInvokeID() int8 { return ul.InvokeID }
func (UpdateLocationRes) Code() byte           { return 2 }
func (UpdateLocationRes) Name() string         { return "UpdateLocation-Res" }

func (UpdateLocationRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		UpdateLocationRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.UpdateLocationRes, e
	}
	c := tmp.UpdateLocationRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (ul UpdateLocationRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// hlr-Number, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, ul.HlrNumber.Bytes())

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// UpdateLocation-Res, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (UpdateLocationRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// UpdateLocation-Res, universal(00) + constructed(20) + sequence(10)
	ul := UpdateLocationRes{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// hlr-Number, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if ul.HlrNumber, e = gsmap.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return ul, nil
	} else if e != nil {
		return nil, e
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = gsmap.UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}

	return ul, nil
}
