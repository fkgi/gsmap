package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

/*
FacilityNotSupported error operation.
FacilityNotSupParam must not be used in version <3.

	facilityNotSupported  ERROR ::= {
		PARAMETER
			FacilityNotSupParam -- optional
		CODE local:21 }

	FacilityNotSupParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		shapeOfLocationEstimateNotSupported          [0] NULL OPTIONAL,
		neededLcsCapabilityNotSupportedInServingNode [1] NULL OPTIONAL }
*/
type FacilityNotSupported struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := FacilityNotSupported{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err FacilityNotSupported) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err FacilityNotSupported) GetInvokeID() int8 { return err.InvokeID }
func (FacilityNotSupported) Code() byte            { return 21 }
func (FacilityNotSupported) Name() string          { return "FacilityNotSupported" }

func (FacilityNotSupported) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		FacilityNotSupported
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.FacilityNotSupported, e
	}
	c := tmp.FacilityNotSupported
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (FacilityNotSupported) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// FacilityNotSupParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (FacilityNotSupported) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// FacilityNotSupParam, universal(00) + constructed(20) + sequence(10)
	err := FacilityNotSupported{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// OPTIONAL TLV
	t, v, e := ReadTLV(buf, 0x00)
	if e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
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
