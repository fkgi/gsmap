package ifd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/common"
	"github.com/fkgi/teldata"
)

const (
	Reset1 gsmap.AppContext = 0x0004000001000a01
	Reset2 gsmap.AppContext = 0x0004000001000a02
	Reset3 gsmap.AppContext = 0x0004000001000a03
)

/*
reset  OPERATION ::= {	--Timer m
	ARGUMENT
		ResetArg
	CODE	local:37 }
*/

func init() {
	a := ResetArg{}
	gsmap.ArgMap[a.Code()] = a
	gsmap.NameMap[a.Name()] = a
}

/*
ResetArg

	ResetArg ::= SEQUENCE {
		sendingNodenumber      SendingNode-Number,
		hlr-List               HLR-List           OPTIONAL,
		-- The hlr-List parameter shall only be applicable for a restart of the HSS/HLR.
		extensionContainer [0] ExtensionContainer OPTIONAL,
		...,
		-- ^^^^^^^^ R99 ^^^^^^^^
		reset-Id-List            [1] Reset-Id-List           OPTIONAL,
		subscriptionData         [2] InsertSubscriberDataArg OPTIONAL,
		subscriptionDataDeletion [3] DeleteSubscriberDataArg OPTIONAL}

	SendingNode-Number ::= CHOICE {
		hlr-Number     ISDN-AddressString,
		css-Number [1] ISDN-AddressString }
	HLR-List ::= SEQUENCE SIZE (1..maxNumOfHLR-Id) OF HLR-Id
	HLR-Id ::= IMSI
*/
type ResetArg struct {
	InvokeID int8 `json:"id"`

	HlrNumber common.AddressString `json:"hlr-Number"`
	HlrList   []teldata.IMSI       `json:"hlr-List,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (re ResetArg) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s (ID=%d)", re.Name(), re.InvokeID)
	fmt.Fprintf(buf, "\n%shlr-Number: %s", gsmap.LogPrefix, re.HlrNumber)
	if len(re.HlrList) != 0 {
		fmt.Fprintf(buf, "\n%scancellationType: %v", gsmap.LogPrefix, re.HlrList)
	}
	return buf.String()
}

func (re ResetArg) GetInvokeID() int8             { return re.InvokeID }
func (ResetArg) GetLinkedID() *int8               { return nil }
func (ResetArg) Code() byte                       { return 37 }
func (ResetArg) Name() string                     { return "Reset-Arg" }
func (ResetArg) DefaultContext() gsmap.AppContext { return Reset1 }

func (ResetArg) NewFromJSON(v []byte, id int8) (gsmap.Component, error) {
	tmp := struct {
		InvokeID *int8 `json:"id"`
		ResetArg
	}{}
	if e := json.Unmarshal(v, &tmp); e != nil {
		return tmp.ResetArg, e
	}
	c := tmp.ResetArg
	if tmp.InvokeID == nil {
		c.InvokeID = id
	} else {
		c.InvokeID = *tmp.InvokeID
	}
	return c, nil
}

func (re ResetArg) MarshalParam() []byte {
	buf := new(bytes.Buffer)

	// hlr-Number, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, re.HlrNumber.Bytes())

	// hlr-List, universal(00) + constructed(20) + sequence(10)
	if len(re.HlrList) != 0 {
		gsmap.WriteTLV(buf, 0x30, marshalImsiList(re.HlrList))
	}

	// extensionContainer, context_specific(80) + constructed(20) + 0(00)

	// Reset-Arg, universal(00) + constructed(20) + sequence(10)
	return gsmap.WriteTLV(new(bytes.Buffer), 0x30, buf.Bytes())
}

func (ResetArg) Unmarshal(id int8, _ *int8, buf *bytes.Buffer) (gsmap.Invoke, error) {
	// Reset-Arg, universal(00) + constructed(20) + sequence(10)
	re := ResetArg{InvokeID: id}
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return nil, e
	} else {
		buf = bytes.NewBuffer(v)
	}

	// hlr-Number, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return nil, e
	} else if re.HlrNumber, e = common.DecodeAddressString(v); e != nil {
		return nil, e
	}

	// OPTIONAL TLV
	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return re, nil
	} else if e != nil {
		return nil, e
	}

	// hlr-List, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		if re.HlrList, e = unmarshalImsiList(v); e != nil {
			return nil, e
		}

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return re, nil
		} else if e != nil {
			return nil, e
		}
	}

	// extensionContainer, context_specific(80) + constructed(20) + 0(00)
	if t == 0xa0 {
		if _, e = common.UnmarshalExtension(v); e != nil {
			return nil, e
		}
		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return re, nil
			} else if e != nil {
				return nil, e
			}
		*/
	}

	return re, nil
}

func marshalImsiList(l []teldata.IMSI) []byte {
	buf := new(bytes.Buffer)
	for _, c := range l {
		// universal(00) + primitive(00) + octet_string(04)
		gsmap.WriteTLV(buf, 0x04, c.Bytes())
	}
	return buf.Bytes()
}

func unmarshalImsiList(b []byte) ([]teldata.IMSI, error) {
	l := []teldata.IMSI{}
	buf := bytes.NewBuffer(b)
	for {
		// universal(00) + primitive(00) + octet_string(04)
		if _, v, e := gsmap.ReadTLV(buf, 0x04); e == io.EOF {
			break
		} else if e != nil {
			return nil, e
		} else if c, e := teldata.DecodeIMSI(v); e != nil {
			return nil, e
		} else {
			l = append(l, c)
		}
	}
	return l, nil
}
