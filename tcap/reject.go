package tcap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fkgi/gsmap"
)

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
