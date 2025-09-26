package gsmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

/*
CallBarred error operation.
callBarringCause must not be used in version 3 and higher.
extensibleCallBarredParam must not be used in version <3.

	callBarred  ERROR ::= {
		PARAMETER
			CallBarredParam -- optional
		CODE	local:13 }

	CallBarredParam ::= CHOICE {
		callBarringCause          CallBarringCause,
		extensibleCallBarredParam ExtensibleCallBarredParam }
	ExtensibleCallBarredParam ::= SEQUENCE {
		callBarringCause	CallBarringCause	OPTIONAL,
		extensionContainer	ExtensionContainer	OPTIONAL,
		... ,
		unauthorisedMessageOriginator	[1] NULL	OPTIONAL,
		-- ^^^^^^ R99 ^^^^^^
		anonymousCallRejection	[2] NULL	OPTIONAL }

unauthorisedMessageOriginator and anonymousCallRejection shall be mutually exclusive.
*/
type CallBarred struct {
	InvokeID int8 `json:"id"`

	NotExtensible bool             `json:"notExtensible,omitempty"`
	Cause         CallBarringCause `json:"callBarringCause,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	UnauthorisedMessageOriginator bool `json:"unauthorisedMessageOriginator,omitempty"`
	// AnonymousCallRejection bool
}

func init() {
	c := CallBarred{}
	ErrMap[c.Code()] = c
	NameMap[c.Name()] = c
}

func (err CallBarred) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", err.Name(), err.InvokeID)
	if err.Cause != 0 {
		fmt.Fprintf(buf, "\n%scallBarringCause: %s", LogPrefix, err.Cause)
	}
	if err.UnauthorisedMessageOriginator {
		fmt.Fprintf(buf, "\n%sunauthorisedMessageOriginator:", LogPrefix)
	}
	return buf.String()
}

func (err CallBarred) GetInvokeID() int8 { return err.InvokeID }
func (CallBarred) Code() byte            { return 13 }
func (CallBarred) Name() string          { return "CallBarred" }

func (CallBarred) NewFromJSON(v []byte, id int8) (Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		CallBarred
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.CallBarred, e
	}
	c := tmp.CallBarred
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (err CallBarred) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// callBarringCause, universal(00) + primitive(00) + enum(0a)
	if tmp := err.Cause.marshal(); tmp != nil {
		WriteTLV(buf, 0x0a, tmp)
	}

	if err.NotExtensible {
		if buf.Len() != 0 {
			return buf.Bytes()
		} else {
			return nil
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	// unauthorisedMessageOriginator, context_specific(80) + primitive(00) + 1(01)
	if err.UnauthorisedMessageOriginator {
		WriteTLV(buf, 0x81, nil)
	}

	// anonymousCallRejection, context_specific(80) + primitive(00) + 2(02)
	// if err.AnonymousCallRejection {
	//	WriteTLV(buf, []byte{0x82}, nil)
	// }

	if buf.Len() != 0 {
		// ExtensibleCallBarredParam, universal(00) + constructed(20) + sequence(10)
		return WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
	}
	return nil
}

func (CallBarred) Unmarshal(id int8, buf *bytes.Buffer) (ReturnError, error) {
	err := CallBarred{InvokeID: id}
	if buf.Len() == 0 {
		return err, nil
	}

	t, v, e := ReadTLV(buf, 0x00)
	if e != nil {
		return nil, e
	}
	// callBarringCause, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Cause.unmarshal(v); e != nil {
			return nil, e
		}
		err.NotExtensible = true
		return err, nil
	}

	// ExtensibleCallBarredParam, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		buf = bytes.NewBuffer(v)
	} else {
		return nil, UnexpectedTag([]byte{0x30}, t)
	}

	// OPTIONAL TLV
	t, v, e = ReadTLV(buf, 0x00)
	if e == io.EOF {
		return err, nil
	} else if e != nil {
		return nil, e
	}

	// callBarringCause, universal(00) + primitive(00) + enum(0a)
	if t == 0x0a {
		if e = err.Cause.unmarshal(v); e != nil {
			return nil, e
		}

		if t, v, e = ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if _, e = UnmarshalExtension(v); e != nil {
			return nil, e
		}

		if t, _, e = ReadTLV(buf, 0x00); e == io.EOF {
			return err, nil
		} else if e != nil {
			return nil, e
		}
	}

	// unauthorisedMessageOriginator, context_specific(80) + primitive(00) + 1(01)
	if t == 0x81 {
		err.UnauthorisedMessageOriginator = true
		/*
			if t, v, e = ReadTLV(buf, nil); e == io.EOF {
				return err, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	// anonymousCallRejection, context_specific(80) + primitive(00) + 2(02)
	// if t[0] == 0x82 {
	//	err.AnonymousCallRejection = true
	//	if t, v, e = ReadTLV(buf, nil); e == io.EOF {
	//		return err, nil
	//	} else if e != nil {
	//		return nil, e
	//	}
	// }

	return err, nil
}

/*
CallBarringCause enum.

	CallBarringCause ::= ENUMERATED {
		barringServiceActive (0),
		operatorBarring      (1)}
*/
type CallBarringCause byte

const (
	_ CallBarringCause = iota
	BarringServiceActive
	OperatorBarring
)

func (c CallBarringCause) String() string {
	switch c {
	case BarringServiceActive:
		return "barringServiceActive"
	case OperatorBarring:
		return "operatorBarring"
	}
	return ""
}

func (c CallBarringCause) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *CallBarringCause) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "barringServiceActive":
		*c = BarringServiceActive
	case "operatorBarring":
		*c = OperatorBarring
	default:
		*c = 0
	}
	return
}

func (c *CallBarringCause) unmarshal(b []byte) (e error) {
	if len(b) != 1 {
		return UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = BarringServiceActive
	case 1:
		*c = OperatorBarring
	default:
		return UnexpectedEnumValue(b)
	}
	return nil
}

func (c CallBarringCause) marshal() []byte {
	switch c {
	case BarringServiceActive:
		return []byte{0x00}
	case OperatorBarring:
		return []byte{0x01}
	}
	return nil
}
