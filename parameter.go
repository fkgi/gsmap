package gsmap

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/fkgi/teldata"
)

/*
ExtensionContainer parameter.

	ExtensionContainer ::= SEQUENCE {
		privateExtensionList [0] PrivateExtensionList OPTIONAL,
		pcs-Extensions       [1] PCS-Extensions       OPTIONAL,
		...}

	PrivateExtensionList ::= SEQUENCE SIZE (1..maxNumOfPrivateExtensions) OF PrivateExtension

	PrivateExtension ::= SEQUENCE {
		extId   MAP-EXTENSION.&extensionId   ({ExtensionSet}),
		extType MAP-EXTENSION.&ExtensionType ({ExtensionSet}{@extId}) OPTIONAL }

	maxNumOfPrivateExtensions  INTEGER ::= 10

	ExtensionSet MAP-EXTENSION ::=
		{...
		-- ExtensionSet is the set of all defined private extensions
		}
		-- Unsupported private extensions shall be discarded if received.

	PCS-Extensions ::= SEQUENCE {
		...}
*/
type ExtensionContainer struct{}

func MarshalExtension(ExtensionContainer) []byte {
	return []byte{}
}

func UnmarshalExtension([]byte) (ExtensionContainer, error) {
	return ExtensionContainer{}, nil
}

/*
OctetString
*/
type OctetString []byte

func (o OctetString) String() string {
	return fmt.Sprintf("%x", []byte(o))
}

func (o *OctetString) UnmarshalJSON(b []byte) (e error) {
	var s string
	if e = json.Unmarshal(b, &s); e != nil {
		return
	}
	var tmp []byte
	if tmp, e = hex.DecodeString(s); e != nil {
		return
	}
	*o = OctetString(tmp)
	return
}

func (o OctetString) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(o))
}

/*
AddressString

	AddressString ::= OCTET STRING (SIZE (1..maxAddressLength))
		-- This type is used to represent a number for addressing purposes.
		-- a) The first octet includes a one bit extension indicator, a 3 bits nature of address indicator and a 4 bits numbering plan indicator, encoded as follows:
			-- bit 8: 1 (no extension)
			-- bits 765: nature of address indicator
			-- bits 4321: numbering plan indicator
		-- b) The following octets representing digits of an address encoded as a TBCD-STRING.
	maxAddressLength INTEGER ::= 20
	ISDN-AddressString ::=
		AddressString (SIZE (1..maxISDN-AddressLength))
*/
type AddressString struct {
	NatureOfAddress teldata.NatureOfAddress `json:"na"`
	NumberingPlan   teldata.NumberingPlan   `json:"np"`
	Digits          teldata.TBCD            `json:"digits"`
}

func (a AddressString) String() string {
	return fmt.Sprintf("%s (NA=%s, NP=%s)",
		a.Digits, a.NatureOfAddress, a.NumberingPlan)
}

func (a AddressString) Bytes() []byte {
	return append(
		[]byte{0x80 | byte(a.NatureOfAddress<<4) | byte(a.NumberingPlan)},
		a.Digits.Bytes()...)
}

func (a AddressString) IsEmpty() bool {
	return a.Digits.Length() == 0
}

func DecodeAddressString(data []byte) (a AddressString, e error) {
	if len(data) == 0 {
		e = UnexpectedTLV("invalid empty data")
	} else {
		a.NatureOfAddress = teldata.NatureOfAddress(data[0]&0x70) >> 4
		a.NumberingPlan = teldata.NumberingPlan(data[0] & 0x0f)
		a.Digits = make(teldata.TBCD, len(data)-1)
		copy(a.Digits, data[1:])
	}
	return
}
