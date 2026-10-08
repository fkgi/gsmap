package tcap

import (
	"bytes"
	"fmt"

	"github.com/fkgi/gsmap"
)

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
