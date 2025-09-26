package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

/*
SystemFailure error operation.
networkResource must not be used in version 3.
extensibleSystemFailureParam must not be used in version <3.

	systemFailure  ERROR ::= {
		PARAMETER
			SystemFailureParam -- optional
		CODE local:34 }

	SystemFailureParam ::= CHOICE {
		networkResource              NetworkResource,
		extensibleSystemFailureParam ExtensibleSystemFailureParam }

	ExtensibleSystemFailureParam ::= SEQUENCE {
		networkResource    NetworkResource    OPTIONAL,
		extensionContainer ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		additionalNetworkResource [0] AdditionalNetworkResource OPTIONAL,
		failureCauseParam         [1] FailureCauseParam         OPTIONAL }
*/
type SystemFailure struct {
	InvokeID int8 `json:"id"`

	NotExtensible bool            `json:"notExtensible,omitempty"`
	Resource      NetworkResource `json:"networkResource,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := SystemFailure{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err SystemFailure) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Resource != 0 {
		fmt.Fprintf(buf, "\n%snetworkResource: %s", LogPrefix, err.Resource)
	}
	return buf.String()
}

func (err SystemFailure) GetInvokeID() int8 { return err.InvokeID }
func (SystemFailure) Code() byte            { return 34 }
func (SystemFailure) Name() string          { return "SystemFailure" }

func (SystemFailure) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		SystemFailure
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.SystemFailure, e
	}
	c := tmp.SystemFailure
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err SystemFailure) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	if err.NotExtensible {
		// networkResource, universal(00) + primitive(00) + enum(0a)
		if tmp := err.Resource.marshal(); tmp != nil {
			WriteTLV(buf, 0x0a, tmp)
		} else {
			WriteTLV(buf, 0x0a, []byte{0x00})
		}
		return buf.Bytes()
	}

	// networkResource, universal(00) + primitive(00) + enum(0a)
	if tmp := err.Resource.marshal(); tmp != nil {
		WriteTLV(buf, 0x0a, tmp)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// ExtensibleSystemFailureParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (SystemFailure) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	err := SystemFailure{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	}

	t, v, e := ReadTLV(buf, 0x00)
	if e != nil {
		return nil, e
	}
	// networkResource, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Resource.unmarshal(v); e != nil {
			return nil, e
		}
		err.NotExtensible = true
		return err, nil
	}

	// ExtensibleSystemFailureParam, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		buf = bytes.NewBuffer(v)
	} else {
		return nil, UnexpectedTag([]byte{0x30}, t)
	}

	// OPTIONAL TLV
	t, v, e = ReadTLV(buf, 0x00)
	if e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	}

	// networkResource, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Resource.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}
		/*
			if t, v, e = ReadTLV(buf, nil); e == io.EOF {
				return err, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return err, nil
}

/*
NetworkResource enum.

	NetworkResource ::= ENUMERATED {
		plmn           (0),
		hlr            (1),
		vlr            (2),
		pvlr           (3),
		controllingMSC (4),
		vmsc           (5),
		eir            (6),
		rss            (7)}
*/
type NetworkResource byte

const (
	_ NetworkResource = iota
	ResourcePlmn
	ResourceHlr
	ResourceVlr
	ResourcePvlr
	ResourceControllingMSC
	ResourceVmsc
	ResourceEir
	ResourceRss
)

func (d NetworkResource) String() string {
	switch d {
	case ResourcePlmn:
		return "plmn"
	case ResourceHlr:
		return "hlr"
	case ResourceVlr:
		return "vlr"
	case ResourcePvlr:
		return "pvlr"
	case ResourceControllingMSC:
		return "controllingMSC"
	case ResourceVmsc:
		return "vmsc"
	case ResourceEir:
		return "eir"
	case ResourceRss:
		return "rss"
	}
	return ""
}

func (d NetworkResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *NetworkResource) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "plmn":
		*d = ResourcePlmn
	case "hlr":
		*d = ResourceHlr
	case "vlr":
		*d = ResourceVlr
	case "pvlr":
		*d = ResourcePvlr
	case "controllingMSC":
		*d = ResourceControllingMSC
	case "vmsc":
		*d = ResourceVmsc
	case "eir":
		*d = ResourceEir
	case "rss":
		*d = ResourceRss
	default:
		*d = 0
	}
	return
}

func (d *NetworkResource) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*d = ResourcePlmn
	case 1:
		*d = ResourceHlr
	case 2:
		*d = ResourceVlr
	case 3:
		*d = ResourcePvlr
	case 4:
		*d = ResourceControllingMSC
	case 5:
		*d = ResourceVmsc
	case 6:
		*d = ResourceEir
	case 7:
		*d = ResourceRss
	default:
		return UnexpectedEnumValue(b)
	}
	return nil
}

func (d NetworkResource) marshal() []byte {
	switch d {
	case ResourcePlmn:
		return []byte{0x00}
	case ResourceHlr:
		return []byte{0x01}
	case ResourceVlr:
		return []byte{0x02}
	case ResourcePvlr:
		return []byte{0x03}
	case ResourceControllingMSC:
		return []byte{0x04}
	case ResourceVmsc:
		return []byte{0x05}
	case ResourceEir:
		return []byte{0x06}
	case ResourceRss:
		return []byte{0x07}
	}
	return nil
}
