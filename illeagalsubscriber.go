package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

/*
IllegalSubscriber error operation.
IllegalSubscriberParam must not be used in version <3.

	illegalSubscriber  ERROR ::= {
		PARAMETER
			IllegalSubscriberParam -- optional
		CODE local:9 }
	IllegalSubscriberParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		... }
*/
type IllegalSubscriber struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := IllegalSubscriber{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err IllegalSubscriber) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err IllegalSubscriber) GetInvokeID() int8 { return err.InvokeID }
func (IllegalSubscriber) Code() byte            { return 9 }
func (IllegalSubscriber) Name() string          { return "IllegalSubscriber" }

func (IllegalSubscriber) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		IllegalSubscriber
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.IllegalSubscriber, e
	}
	c := tmp.IllegalSubscriber
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (IllegalSubscriber) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// IllegalSubscriberParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (IllegalSubscriber) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// IllegalSubscriberParam, universal(00) + constructed(20) + sequence(10)
	err := IllegalSubscriber{InvokeID: id}
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
