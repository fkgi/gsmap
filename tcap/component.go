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
	Component ::= CHOICE {
		invoke              [1] IMPLICIT Invoke,
		returnResultLast    [2] IMPLICIT ReturnResult,
		returnError         [3] IMPLICIT ReturnError,
		reject              [4] IMPLICIT Reject,
		returnResultNotLast [7] IMPLICIT ReturnResult }
*/

func marshalComponents(cs []gsmap.Component) []byte {
	buf := new(bytes.Buffer)
	for _, c := range cs {
		switch c := c.(type) {
		case gsmap.Invoke:
			// Invoke, context_specific(80) + constructed(20) + 1(01)
			gsmap.WriteTLV(buf, 0xa1, marshalInvoke(c))
		case gsmap.ReturnResultLast:
			// ReturnResultLast, context_specific(80) + constructed(20) + 2(02)
			gsmap.WriteTLV(buf, 0xa2, marshalReturnResultLast(c))
		case gsmap.ReturnError:
			// ReturnError, context_specific(80) + constructed(20) + 3(03)
			gsmap.WriteTLV(buf, 0xa3, marshalReturnError(c))
		case Reject:
			// Reject, context_specific(80) + constructed(20) + 4(04)
			gsmap.WriteTLV(buf, 0xa4, marshalReject(c))
		case gsmap.ReturnResult:
			// ReturnResult, context_specific(80) + constructed(20) + 7(07)
			gsmap.WriteTLV(buf, 0xa7, marshalReturnResult(c))
		}
	}
	return buf.Bytes()
}

func unmarshalComponents(data []byte) ([]gsmap.Component, error) {
	buf := bytes.NewBuffer(data)
	cs := make([]gsmap.Component, 0)
	for {
		t, v, e := gsmap.ReadTLV(buf, 0x00)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}

		var c gsmap.Component
		switch t {
		case 0xa1: // invoke, context_specific(80) + constructed(20) + 1(01)
			c, e = unmarshalInvoke(v)
		case 0xa2: // returnResultLast, context_specific(80) + constructed(20) + 2(02)
			c, e = unmarshalReturnResultLast(v)
		case 0xa3: // returnError, context_specific(80) + constructed(20) + 3(03)
			c, e = unmarshalReturnError(v)
		case 0xa4: // reject, context_specific(80) + constructed(20) + 4(04)
			c, e = unmarshalReject(v)
		case 0xa7: // returnResult, context_specific(80) + constructed(20) + 7(07)
			c, e = unmarshalReturnResult(v)
		default:
			e = gsmap.UnexpectedTag([]byte{0xa1, 0xa2, 0xa3, 0xa4, 0xa7}, t)
		}

		if e != nil {
			return nil, e
		}
		cs = append(cs, c)
	}
	return cs, nil
}

/*
	Invoke ::= SEQUENCE {
		invokeID          InvokeIdType,
		linkedID      [0] IMPLICIT InvokeIdType        OPTIONAL,
		operationCode     OPERATION,
		parameter         ANY DEFINED BY operationCode OPTIONAL }

	InvokeIdType ::= INTEGER (–128..127)
	OPERATION ::= CHOICE {
		localValue  INTEGER,
		globalValue OBJECT IDENTIFIER }
*/

func marshalInvoke(c gsmap.Invoke) []byte {
	buf := new(bytes.Buffer)

	// invokeID, universal(00) + primitive(00) + integer(02)
	gsmap.WriteTLV(buf, 0x02, []byte{byte(c.GetInvokeID())})

	// linkedID, context_specific(08) + primitive(00) + 0(00)
	if i := c.GetLinkedID(); i != nil {
		gsmap.WriteTLV(buf, 0x80, []byte{byte(*i)})
	}

	// operationCode, universal(00) + primitive(00) + integer(02)
	gsmap.WriteTLV(buf, 0x02, []byte{c.Code()})

	// parameter
	if param := c.MarshalParam(); param != nil {
		buf.Write(param)
	}
	return buf.Bytes()
}

func unmarshalInvoke(data []byte) (gsmap.Invoke, error) {
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

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e != nil {
		return nil, e
	}

	// linkedID, context_specific(08) + primitive(00) + 0(00)
	var lid *int8
	if t == 0x80 {
		if len(v) != 1 {
			return nil, gsmap.UnexpectedTLV("invalid linkedID value")
		}
		tmp := int8(v[0])
		lid = &tmp

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e != nil {
			return nil, e
		}
	}

	// operationCode, universal(00) + primitive(00) + integer(02)
	if t != 0x02 {
		return nil, gsmap.UnexpectedTag([]byte{0x02}, t)
	} else if len(v) != 1 {
		return nil, gsmap.UnexpectedTLV("invalid operation code")
	} else if op := gsmap.ArgMap[v[0]]; op == nil {
		return nil, gsmap.UnexpectedTLV(fmt.Sprintf(
			"invoke operation code %#x is not supported", v[0]))
	} else {
		// parameter
		return op.Unmarshal(iid, lid, buf)
	}
}

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

/*
Reject component portion struct.

	Reject ::= SEQUENCE {
		invokeID CHOICE {
			derivable     InvokeIdType,
			not-derivable NULL },
		problem CHOICE {
			generalProblem      [0] IMPLICIT GeneralProblem,
			invokeProblem       [1] IMPLICIT InvokeProblem,
			returnResultProblem [2] IMPLICIT ReturnResultProblem,
			returnErrorProblem  [3] IMPLICIT ReturnErrorProblem } }
*/
type Reject struct {
	InvokeID *int8
	Problem  byte
}

const (
	/*
		GeneralProblem ::= INTEGER {
			unrecognizedComponent    (0),
			mistypedComponent        (1),
			badlyStructuredComponent (2) }
	*/
	UnrecognizedComponent    byte = 0x00
	GeneralMistypedComponent byte = 0x01
	BadlyStructuredComponent byte = 0x02

	/*
		InvokeProblem ::= INTEGER {
			duplicateInvokeID         (0),
			unrecognizedOperation     (1),
			mistypedParameter         (2),
			resourceLimitation        (3),
			initiatingRelease         (4),
			unrecognizedLinkedID      (5),
			linkedResponseUnexpected  (6),
			unexpectedLinkedOperation (7) }
	*/
	DuplicateInvokeID         byte = 0x10
	UnrecognizedOperation     byte = 0x11
	InvokeMistypedParameter   byte = 0x12
	ResourceLimitation        byte = 0x13
	InitiatingRelease         byte = 0x14
	UnrecognizedLinkedID      byte = 0x15
	LinkedResponseUnexpected  byte = 0x16
	UnexpectedLinkedOperation byte = 0x17

	/*
		ReturnResultProblem ::= INTEGER {
			unrecognizedInvokeID   (0),
			returnResultUnexpected (1),
			mistypedParameter      (2) }
	*/
	ResultUnrecognizedInvokeID byte = 0x20
	ReturnResultUnexpected     byte = 0x21
	ResultMistypedParameter    byte = 0x22

	/*
		ReturnErrorProblem ::= INTEGER {
			unrecognizedInvokeID  (0),
			returnErrorUnexpected (1),
			unrecognizedError     (2),
			unexpectedError       (3),
			mistypedParameter     (4) }
	*/
	ErrorUnrecognizedInvokeID byte = 0x30
	ReturnErrorUnexpected     byte = 0x31
	UnrecognizedError         byte = 0x32
	UnexpectedError           byte = 0x33
	ErrorMistypedParameter    byte = 0x34
)

func (c Reject) String() string {
	buf := new(strings.Builder)
	if c.InvokeID == nil {
		fmt.Fprintf(buf, "%s (ID=NULL)", c.Name())
	} else {
		fmt.Fprintf(buf, "%s (ID=%d)", c.Name(), *c.InvokeID)
	}
	switch c.Problem {
	case UnrecognizedComponent:
		fmt.Fprintf(buf, "\n%sgeneralProblem: unrecognizedComponent", gsmap.LogPrefix)
	case GeneralMistypedComponent:
		fmt.Fprintf(buf, "\n%sgeneralProblem: mistypedComponent", gsmap.LogPrefix)
	case BadlyStructuredComponent:
		fmt.Fprintf(buf, "\n%sgeneralProblem: badlyStructuredComponent", gsmap.LogPrefix)
	case DuplicateInvokeID:
		fmt.Fprintf(buf, "\n%sinvokeProblem: duplicateInvokeID", gsmap.LogPrefix)
	case UnrecognizedOperation:
		fmt.Fprintf(buf, "\n%sinvokeProblem: unrecognizedOperation", gsmap.LogPrefix)
	case InvokeMistypedParameter:
		fmt.Fprintf(buf, "\n%sinvokeProblem: mistypedParameter", gsmap.LogPrefix)
	case ResourceLimitation:
		fmt.Fprintf(buf, "\n%sinvokeProblem: resourceLimitation", gsmap.LogPrefix)
	case InitiatingRelease:
		fmt.Fprintf(buf, "\n%sinvokeProblem: initiatingRelease", gsmap.LogPrefix)
	case UnrecognizedLinkedID:
		fmt.Fprintf(buf, "\n%sinvokeProblem: unrecognizedLinkedID", gsmap.LogPrefix)
	case LinkedResponseUnexpected:
		fmt.Fprintf(buf, "\n%sinvokeProblem: linkedResponseUnexpected", gsmap.LogPrefix)
	case UnexpectedLinkedOperation:
		fmt.Fprintf(buf, "\n%sinvokeProblem: unexpectedLinkedOperation", gsmap.LogPrefix)
	case ResultUnrecognizedInvokeID:
		fmt.Fprintf(buf, "\n%sreturnResultProblem: unrecognizedInvokeID", gsmap.LogPrefix)
	case ReturnResultUnexpected:
		fmt.Fprintf(buf, "\n%sreturnResultProblem: returnResultUnexpected", gsmap.LogPrefix)
	case ResultMistypedParameter:
		fmt.Fprintf(buf, "\n%sreturnResultProblem: mistypedParameter", gsmap.LogPrefix)
	case ErrorUnrecognizedInvokeID:
		fmt.Fprintf(buf, "\n%sreturnErrorProblem: unrecognizedInvokeID", gsmap.LogPrefix)
	case ReturnErrorUnexpected:
		fmt.Fprintf(buf, "\n%sreturnErrorProblem: returnErrorUnexpected", gsmap.LogPrefix)
	case UnrecognizedError:
		fmt.Fprintf(buf, "\n%sreturnErrorProblem: unrecognizedError", gsmap.LogPrefix)
	case UnexpectedError:
		fmt.Fprintf(buf, "\n%sreturnErrorProblem: unexpectedError", gsmap.LogPrefix)
	case ErrorMistypedParameter:
		fmt.Fprintf(buf, "\n%sreturnErrorProblem: mistypedParameter", gsmap.LogPrefix)
	default:
		fmt.Fprintf(buf, "\n%sproblem: unknown", gsmap.LogPrefix)
	}
	return buf.String()
}

func (c Reject) GetInvokeID() int8 {
	if c.InvokeID != nil {
		return *c.InvokeID
	}
	return 0
}
func (Reject) Code() byte           { return 0 }
func (Reject) Name() string         { return "Reject" }
func (Reject) MarshalParam() []byte { return nil }

func (Reject) NewFromJSON(v []byte, i int8) (gsmap.Component, error) {
	c := Reject{}
	e := json.Unmarshal(v, &c)
	return c, e
}

func marshalReject(c Reject) []byte {
	buf := new(bytes.Buffer)

	// invokeID, CHOICE
	if c.InvokeID != nil {
		// invokeID, universal(00) + primitive(00) + integer(02)
		gsmap.WriteTLV(buf, 0x02, []byte{byte(*c.InvokeID)})
	} else {
		// invokeID, universal(00) + primitive(00) + null(05)
		gsmap.WriteTLV(buf, 0x05, nil)
	}

	// problem CHOICE, context_specific(80) + primitive(00) + 0-3(00-03)
	return gsmap.WriteTLV(buf, 0x80|(c.Problem>>4), []byte{c.Problem & 0x0f})
}

func unmarshalReject(data []byte) (Reject, error) {
	buf := bytes.NewBuffer(data)
	c := Reject{}

	// invokeID, CHOICE
	if t, v, e := gsmap.ReadTLV(buf, 0x00); e != nil {
		return c, e
	} else if t == 0x02 { // invokeID, universal(00) + primitive(00) + integer(02)
		if len(v) != 1 {
			return c, gsmap.UnexpectedTLV("invalid invokeID value")
		}
		tmp := int8(v[0])
		c.InvokeID = &tmp
	} else if t == 0x05 { // invokeID, universal(00) + primitive(00) + null(05)
		c.InvokeID = nil
	} else {
		return c, gsmap.UnexpectedTag([]byte{0x02, 0x05}, t)
	}

	// problem CHOICE, context_specific(80) + primitive(00) + 0-3(00-03)
	if t, v, e := gsmap.ReadTLV(buf, 0x00); e != nil {
		return c, e
	} else if len(v) != 1 || v[0]&0xf0 != 0x00 {
		return c, gsmap.UnexpectedTLV("invalid parameter value")
	} else if t == 0x80 { // generalProblem
		c.Problem = v[0] & 0x0f
	} else if t == 0x81 { // invokeProblem
		c.Problem = v[0]&0x0f | 0x10
	} else if t == 0x82 { // returnResultProblem
		c.Problem = v[0]&0x0f | 0x20
	} else if t == 0x83 { // returnErrorProblem
		c.Problem = v[0]&0x0f | 0x30
	} else {
		return c, gsmap.UnexpectedTag([]byte{0x80, 0x81, 0x82, 0x83}, t)
	}
	return c, nil
}
