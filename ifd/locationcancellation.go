package ifd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/common"
)

const (
	LocationCancellation1 gsmap.AppContext = 0x0004000001000201
	LocationCancellation2 gsmap.AppContext = 0x0004000001000202
	LocationCancellation3 gsmap.AppContext = 0x0004000001000203
)

/*
cancelLocation OPERATION ::= {
	ARGUMENT
		CancelLocationArg
	RESULT
		CancelLocationRes	-- optional
	ERRORS {
		dataMissing         |
		unexpectedDataValue }
	CODE	local:3 }
*/

func init() {
	a := CancelLocationArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a

	r := CancelLocationRes{}
	gsmap.ResMap[r.Code()] = r
	gsmap.NameMap[r.Name()] = r
}

/*
CancelLocationArg operation arg.

	CancelLocationArg ::= [3] SEQUENCE {
		identity           Identity,
		cancellationType   CancellationType   OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^^^ R99 ^^^^^^^^
		typeOfUpdate                   [0] TypeOfUpdate       OPTIONAL,
		mtrf-SupportedAndAuthorized    [1] NULL               OPTIONAL,
		mtrf-SupportedAndNotAuthorized [2] NULL               OPTIONAL,
		newMSC-Number                  [3] ISDN-AddressString OPTIONAL,
		newVLR-Number                  [4] ISDN-AddressString OPTIONAL,
		new-lmsi                       [5] LMSI               OPTIONAL,
		reattach-Required              [6] NULL               OPTIONAL
		}
		--mtrf-SupportedAndAuthorized and mtrf-SupportedAndNotAuthorized shall not
		-- both be present
*/
type CancelLocationArg struct {
	InvokeID int8 `json:"id"`

	Identity         Identity         `json:"identity"`
	CancellationType CancellationType `json:"cancellationType,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (cl CancelLocationArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", cl.Name(), cl.InvokeID)
	fmt.Fprintf(buf, "\n%sidentity: %s", gsmap.LogPrefix, cl.Identity)
	if cl.CancellationType != 0 {
		fmt.Fprintf(buf, "\n%scancellationType: %s", gsmap.LogPrefix, cl.CancellationType)
	}
	return buf.String()
}

func (cl CancelLocationArg) GetInvokeID() int8             { return cl.InvokeID }
func (CancelLocationArg) GetLinkedID() *int8               { return nil }
func (CancelLocationArg) Code() byte                       { return 3 }
func (CancelLocationArg) Name() string                     { return "CancelLocation-Arg" }
func (CancelLocationArg) DefaultContext() gsmap.AppContext { return LocationCancellation1 }

func (CancelLocationArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		CancelLocationArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.CancelLocationArg, e
	}
	c := tmp.CancelLocationArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (cl CancelLocationArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// identity, choice
	cl.Identity.marshalTo(buf)

	// cancellationType, universal(00) + primitive(00) + enum(0a)
	if tmp := cl.CancellationType.marshal(); tmp != nil {
		gsmap.WriteTLV(buf, 0x0a, tmp)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// CancelLocation-Arg, context_specific(80) + constructed(20) + 3(03)
	return gsmap.WriteTLV(new(bytes.Buffer), 0xa3, buf.Bytes())
}

func (CancelLocationArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// CancelLocation-Arg, context_specific(80) + constructed(20) + 3(03)
	cl := CancelLocationArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0xa3); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// identity, choice
	if e := cl.Identity.unmarshalFrom(buf); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return cl, nil
	} else if e != nil {
		return nil, e
	}

	// cancellationType, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = cl.CancellationType.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return cl, nil
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

	return cl, nil
}

/*
CancelLocationRes operation res.

	CancelLocationRes ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		...}
*/
type CancelLocationRes struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (cl CancelLocationRes) String() string {
	return fmt.Sprintf("%s (ID=%d)", cl.Name(), cl.InvokeID)
}

func (cl CancelLocationRes) GetInvokeID() int8 { return cl.InvokeID }
func (CancelLocationRes) Code() byte           { return 3 }
func (CancelLocationRes) Name() string         { return "CancelLocation-Res" }

func (CancelLocationRes) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		CancelLocationRes
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.CancelLocationRes, e
	}
	c := tmp.CancelLocationRes
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (cl CancelLocationRes) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// CancelLocationRes, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (CancelLocationRes) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	// CancelLocation-Res, universal(00) + constructed(20) + sequence(10)
	cl := CancelLocationRes{InvokeID: id}
	if buf.Len() == 0 {
		return cl, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return cl, nil
	} else if e != nil {
		return nil, e
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}

	return cl, nil
}
