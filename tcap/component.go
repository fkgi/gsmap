package tcap

import (
	"bytes"
	"io"

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
