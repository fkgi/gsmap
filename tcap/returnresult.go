package tcap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
)

/*
	ReturnResult ::= SEQUENCE {
		invokeID InvokeIdType,
		result   SEQUENCE {
			operationCode OPERATION,
			parameter     ANY DEFINED BY operationCode } OPTIONAL }
*/

func marshalReturnResultLast(c gsmap.ReturnResultLast) []byte {
	buf := new(bytes.Buffer)

	// invokeID, universal(00) + primitive(00) + integer(02)
	gsmap.WriteTLV(buf, 0x02, []byte{byte(c.GetInvokeID())})

	// result
	res := c.MarshalParam()
	if res == nil {
		return buf.Bytes()
	}

	buf2 := new(bytes.Buffer)
	// operationCode, universal(00) + primitive(00) + integer(02)
	gsmap.WriteTLV(buf2, 0x02, []byte{c.Code()})
	// parameter
	buf2.Write(res)

	// result, universal(00) +  constructed(20) + sequence(10)
	return gsmap.WriteTLV(buf, 0x30, buf2.Bytes())
}

func unmarshalReturnResultLast(data []byte) (gsmap.ReturnResultLast, error) {
	buf := bytes.NewBuffer(data)

	// invokeID, universal(00) + primitive(00) + integer(02)
	var iid int8
	if _, v, e := gsmap.ReadTLV(buf, 0x02); e != nil {
		return nil, e
	} else if len(v) != 1 {
		return nil, gsmap.UnexpectedTLV("invalid invokeID value")
	} else {
		iid = int8(v[0])
	}

	// result, universal(00) +  constructed(20) + sequence(10)
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e == io.EOF {
		return EmptyResult{InvokeID: iid}, nil
	} else if e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// operationCode, universal(00) + primitive(00) + integer(02)
	if _, v, e := gsmap.ReadTLV(buf, 0x02); e != nil {
		return nil, e
	} else if len(v) != 1 {
		return nil, gsmap.UnexpectedTLV("invalid operation code")
	} else if op := gsmap.ResMap[v[0]]; op == nil {
		return nil, gsmap.UnexpectedTLV(fmt.Sprintf(
			"response operation code %#x is not supported", v[0]))
	} else {
		// parameter
		return op.Unmarshal(iid, buf)
	}
}

/*
	ReturnError ::= SEQUENCE {
		invokeID  InvokeIdType,
		errorCode ERROR,
		parameter ANY DEFINED BY errorCode OPTIONAL }
*/

func marshalReturnError(c gsmap.ReturnError) []byte {
	buf := new(bytes.Buffer)

	// invokeID, universal(00) + primitive(00) + integer(02)
	gsmap.WriteTLV(buf, 0x02, []byte{byte(c.GetInvokeID())})

	// errorCode, universal(00) + primitive(00) + integer(02)
	gsmap.WriteTLV(buf, 0x02, []byte{c.Code()})

	// parameter
	if param := c.MarshalParam(); param != nil {
		buf.Write(param)
	}
	return buf.Bytes()
}

func unmarshalReturnError(data []byte) (gsmap.ReturnError, error) {
	buf := bytes.NewBuffer(data)

	// invokeID, universal(00) + primitive(00) + integer(02)
	var iid int8
	if _, v, e := gsmap.ReadTLV(buf, 0x02); e != nil {
		return nil, e
	} else if len(v) != 1 {
		return nil, gsmap.UnexpectedTLV("invalid invokeID value")
	} else {
		iid = int8(v[0])
	}

	// ErrorCode, universal(00) + primitive(00) + integer(02)
	if _, v, e := gsmap.ReadTLV(buf, 0x02); e != nil {
		return nil, e
	} else if len(v) != 1 {
		return nil, gsmap.UnexpectedTLV("invalid operation code")
	} else if op := gsmap.ErrMap[v[0]]; op == nil {
		return nil, gsmap.UnexpectedTLV(fmt.Sprintf(
			"error operation code %#x is not supported", v[0]))
	} else {
		// parameter
		return op.Unmarshal(iid, buf)
	}
}

// EmptyResult is ReturnResult witout result parameter.
type EmptyResult struct {
	InvokeID int8 `json:"id"`
}

func (c EmptyResult) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", c.Name(), c.InvokeID)
	return buf.String()
}

func (c EmptyResult) GetInvokeID() int8  { return c.InvokeID }
func (EmptyResult) Code() byte           { return 0 }
func (EmptyResult) Name() string         { return "EmptyResult" }
func (EmptyResult) MarshalParam() []byte { return nil }

func (EmptyResult) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		EmptyResult
	}{}
	e := json.Unmarshal(v, &tmp)
	c := tmp.EmptyResult

	if e != nil {
	} else if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, e
}

func (EmptyResult) Unmarshal(id int8, buf *bytes.Buffer) (gsmap.ReturnResultLast, error) {
	return EmptyResult{InvokeID: id}, nil
}

func marshalReturnResult(gsmap.ReturnResult) []byte {
	return []byte{}
}
func unmarshalReturnResult([]byte) (gsmap.ReturnResult, error) {
	return nil, nil
}
