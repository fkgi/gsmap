package ifc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/common"
)

/*
MessageWaitingListFull error operation.

	messageWaitingListFull  ERROR ::= {
		PARAMETER
			MessageWaitListFullParam	-- optional
		CODE	local:33 }

	MessageWaitListFullParam ::= SEQUENCE {
		extensionContainer	ExtensionContainer	OPTIONAL,
		...}
*/
type MessageWaitingListFull struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := MessageWaitingListFull{}
	gsmap.ErrMap[c.Code()] = c
	gsmap.NameMap[c.Name()] = c
}

func (err MessageWaitingListFull) String() string {
	return fmt.Sprint(err.Name(), "(ID", err.InvokeID, ")")
}

func (err MessageWaitingListFull) GetInvokeID() int8 { return err.InvokeID }
func (MessageWaitingListFull) Code() byte            { return 33 }
func (MessageWaitingListFull) Name() string          { return "MessageWaitingListFull" }

func (MessageWaitingListFull) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		MessageWaitingListFull
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.MessageWaitingListFull, e
	}
	c := tmp.MessageWaitingListFull
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (MessageWaitingListFull) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// MessageWaitingListFull, universal(00) + constructed(20) + sequence(10)
		return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (MessageWaitingListFull) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnError, error) {
	// MessageWaitingListFull, universal(00) + constructed(20) + sequence(10)
	err := MessageWaitingListFull{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	} else if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t, v, e := gsmap.ReadTLV(buf, 0x00); e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	} else if t == 0x30 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}
	}
	return err, nil
}
