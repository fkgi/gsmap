package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

/*
TeleserviceNotProvisioned error operation.
TeleservNotProvParam must not be used in version <3.

	teleserviceNotProvisioned  ERROR ::= {
		PARAMETER
			TeleservNotProvParam -- optional
		CODE	local:11 }

	TeleservNotProvParam ::= SEQUENCE {
		extensionContainer	ExtensionContainer	OPTIONAL,
		...}
*/
type TeleserviceNotProvisioned struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := TeleserviceNotProvisioned{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err TeleserviceNotProvisioned) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err TeleserviceNotProvisioned) GetInvokeID() int8 { return err.InvokeID }
func (TeleserviceNotProvisioned) Code() byte            { return 11 }
func (TeleserviceNotProvisioned) Name() string          { return "TeleserviceNotProvisioned" }

func (TeleserviceNotProvisioned) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		TeleserviceNotProvisioned
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.TeleserviceNotProvisioned, e
	}
	c := tmp.TeleserviceNotProvisioned
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (TeleserviceNotProvisioned) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// TeleservNotProvParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (TeleserviceNotProvisioned) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// TeleservNotProvParam, universal(00) + constructed(20) + sequence(10)
	err := TeleserviceNotProvisioned{InvokeID: id}
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
