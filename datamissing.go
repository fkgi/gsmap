package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

/*
DataMissing error operation.
DataMissing must not be used in version 1.
DataMissingParam must not be used in version <3.

	dataMissing  ERROR ::= {
		PARAMETER
			DataMissingParam -- optional
		CODE local:35 }

	DataMissingParam ::= SEQUENCE {
		extensionContainer ExtensionContainer OPTIONAL,
		... }
*/
type DataMissing struct {
	InvokeID int8 `json:"id"`

	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func init() {
	c := DataMissing{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err DataMissing) String() string {
	return fmt.Sprintf("%s (ID=%d)", err.Name(), err.InvokeID)
}

func (err DataMissing) GetInvokeID() int8 { return err.InvokeID }
func (DataMissing) Code() byte            { return 35 }
func (DataMissing) Name() string          { return "DataMissing" }

func (DataMissing) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		DataMissing
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.DataMissing, e
	}
	c := tmp.DataMissing
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (DataMissing) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	if buf.Len() != 0 {
		// DataMissingParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (DataMissing) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	// DataMissingParam, universal(00) + constructed(20) + sequence(10)
	err := DataMissing{InvokeID: id}
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
