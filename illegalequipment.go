package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

/*
IllegalEquipment error operation.
IllegalEquipment must not be used in version 1.
IllegalEquipmentParam must not be used in version <3

	illegalEquipment  ERROR ::= {
		PARAMETER
			IllegalEquipmentParam -- optional
		CODE local:12 }
	IllegalEquipmentParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		... }
*/
type IllegalEquipment struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := IllegalEquipment{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err IllegalEquipment) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err IllegalEquipment) GetInvokeID() int8 { return err.InvokeID }
func (IllegalEquipment) Code() byte            { return 12 }
func (IllegalEquipment) Name() string          { return "IllegalEquipment" }

func (IllegalEquipment) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		IllegalEquipment
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.IllegalEquipment, e
	}
	c := tmp.IllegalEquipment
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (IllegalEquipment) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// IllegalEquipmentParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (IllegalEquipment) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// IllegalEquipmentParam, universal(00) + constructed(20) + sequence(10)
	err := IllegalEquipment{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t, v, e := ReadTLV(buf, 0x00); e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	} else if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}
	return err, nil
}
