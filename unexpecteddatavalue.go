package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

/*
UnexpectedDataValue error operation.
UnexpectedDataParam must not be used in version <3.

	unexpectedDataValue  ERROR ::= {
		PARAMETER
			UnexpectedDataParam -- optional
		CODE local:36 }

	UnexpectedDataParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^ R99 ^^^^^^
		unexpectedSubscriber [0] NULL OPTIONAL }

the unexpectedSubscriber indication in the unexpectedDataValue error shall not be used
for operations that allow the unidentifiedSubscriber error.
*/
type UnexpectedDataValue struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := UnexpectedDataValue{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err UnexpectedDataValue) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err UnexpectedDataValue) GetInvokeID() int8 { return err.InvokeID }
func (UnexpectedDataValue) Code() byte            { return 36 }
func (UnexpectedDataValue) Name() string          { return "UnexpectedDataValue" }

func (UnexpectedDataValue) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		UnexpectedDataValue
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.UnexpectedDataValue, e
	}
	c := tmp.UnexpectedDataValue
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (UnexpectedDataValue) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// UnexpectedDataParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (UnexpectedDataValue) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// UnexpectedDataParam, universal(00) + constructed(20) + sequence(10)
	err := UnexpectedDataValue{InvokeID: id}
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
			if t, _, e = ReadTLV(buf, nil); e == io.EOF {
				return err, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}
	return err, nil
}
