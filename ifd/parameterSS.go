package ifd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/teldata"
)

/*
ssInfoList

	Ext-SS-InfoList ::= SEQUENCE SIZE (1..maxNumOfSS) OF Ext-SS-Info

	Ext-SS-Info ::= CHOICE {
		forwardingInfo	[0] Ext-ForwInfo,
		callBarringInfo	[1] Ext-CallBarInfo,
		cug-Info	    [2] CUG-Info,
		ss-Data	        [3] Ext-SS-Data,
		emlpp-Info      [4] EMLPP-Info}
*/
type ssInfoList []ssInfo

type ssInfo interface {
	GetType() string
	marshal() []byte
}

func (l ssInfoList) String() string {
	buf := new(strings.Builder)
	for _, i := range l {
		switch info := i.(type) {
		case forwInfo:
			fmt.Fprintf(buf, "\n%s| forwardingInfo:\n%s", gsmap.LogPrefix, info)
		case callBarInfo:
			fmt.Fprintf(buf, "\n%s| callBarInfo:\n%s", gsmap.LogPrefix, info)
		case cugInfo:
			fmt.Fprintf(buf, "\n%s| cug-Info:\n%s", gsmap.LogPrefix, info)
		case ssData:
			fmt.Fprintf(buf, "\n%s| ss-Data:\n%s", gsmap.LogPrefix, info)
		case emlppInfo:
			fmt.Fprintf(buf, "\n%s| emlpp-Info:\n%s", gsmap.LogPrefix, info)
		}
	}
	return buf.String()
}

func (l *ssInfoList) UnmarshalJSON(b []byte) (e error) {
	var m [](map[string]json.RawMessage)
	if e = json.Unmarshal(b, &m); e != nil {
		return
	}
	for _, i := range m {
		for k, v := range i {
			switch k {
			case "forwardingInfo":
				info := forwInfo{}
				if e = json.Unmarshal(v, &info); e != nil {
					return
				}
				*l = append(*l, info)
			case "callBarringInfo":
				info := callBarInfo{}
				if e = json.Unmarshal(v, &info); e != nil {
					return
				}
				*l = append(*l, info)
			case "cug-Info":
				info := cugInfo{}
				if e = json.Unmarshal(v, &info); e != nil {
					return
				}
				*l = append(*l, info)
			case "ss-Data":
				info := ssData{}
				if e = json.Unmarshal(v, &info); e != nil {
					return
				}
				*l = append(*l, info)
			case "emlpp-Info":
				info := emlppInfo{}
				if e = json.Unmarshal(v, &info); e != nil {
					return
				}
				*l = append(*l, info)
			}
			break
		}
	}
	return
}

func (l ssInfoList) MarshalJSON() ([]byte, error) {
	var m [](map[string]json.RawMessage)
	for _, i := range l {
		b, e := json.Marshal(i)
		if e != nil {
			return nil, e
		}
		switch i.(type) {
		case forwInfo:
			m = append(m, map[string]json.RawMessage{
				"forwardingInfo": json.RawMessage(b)})
		case callBarInfo:
			m = append(m, map[string]json.RawMessage{
				"callBarringInfo": json.RawMessage(b)})
		case cugInfo:
			m = append(m, map[string]json.RawMessage{
				"cug-Info": json.RawMessage(b)})
		case ssData:
			m = append(m, map[string]json.RawMessage{
				"ss-Data": json.RawMessage(b)})
		case emlppInfo:
			m = append(m, map[string]json.RawMessage{
				"emlpp-Info": json.RawMessage(b)})
		}
	}
	return json.Marshal(m)
}

func (l ssInfoList) marshal() []byte {
	buf := new(bytes.Buffer)
	for _, i := range l {
		switch i.(type) {
		case forwInfo:
			// forwardingInfo, context_specific(80) + constructed(20) + 0(00)
			gsmap.WriteTLV(buf, 0xa0, i.marshal())
		case callBarInfo:
			// callBarringInfo, context_specific(80) + constructed(20) + 1(01)
			gsmap.WriteTLV(buf, 0xa1, i.marshal())
		case cugInfo:
			// cug-Info, context_specific(80) + constructed(20) + 2(02)
			gsmap.WriteTLV(buf, 0xa2, i.marshal())
		case ssData:
			// ss-Data, context_specific(80) + constructed(20) + 3(03)
			gsmap.WriteTLV(buf, 0xa3, i.marshal())
		case emlppInfo:
			// emlpp-Info, context_specific(80) + constructed(20) + 4(04)
			gsmap.WriteTLV(buf, 0xa4, i.marshal())
		}
	}
	return buf.Bytes()
}

func (ssInfoList) unmarshal(data []byte) (ssInfoList, error) {
	buf := bytes.NewBuffer(data)
	l := ssInfoList{}
	for {
		t, v, e := gsmap.ReadTLV(buf, 0x00)
		if e == io.EOF {
			return l, nil
		} else if e != nil {
			return l, e
		}
		switch t {
		case 0xa0:
			var i forwInfo
			if e = i.unmarshal(v); e == nil {
				l = append(l, i)
			}
		case 0xa1:
			var i callBarInfo
			if e = i.unmarshal(v); e == nil {
				l = append(l, i)
			}
		case 0xa2:
			var i cugInfo
			if e = i.unmarshal(v); e == nil {
				l = append(l, i)
			}
		case 0xa3:
			var i ssData
			if e = i.unmarshal(v); e == nil {
				l = append(l, i)
			}
		case 0xa4:
			var i emlppInfo
			if e = i.unmarshal(v); e == nil {
				l = append(l, i)
			}
		}
		if e != nil {
			return l, e
		}
	}
}

/*
svcCode

	Ext-BasicServiceCode ::= CHOICE {
		ext-BearerService [2] Ext-BearerServiceCode,
		ext-Teleservice   [3] Ext-TeleserviceCode  }
*/
type svcCode struct {
	Code uint8
	Type uint8
}

func (c svcCode) String() string {
	switch c.Type {
	case 2: // BearerServiceCode
		return fmt.Sprintf("ext-BearerService=%d", c.Code)
	case 3: // TeleserviceCode
		return fmt.Sprintf("ext-Teleservice=%d", c.Code)
	}
	return ""
}

func (c svcCode) MarshalJSON() ([]byte, error) {
	switch c.Type {
	case 2: // BearerServiceCode
		return json.Marshal(map[string]uint8{"ext-BearerService": c.Code})
	case 3: // TeleserviceCode
		return json.Marshal(map[string]uint8{"ext-Teleservice": c.Code})
	}
	return nil, fmt.Errorf("unknown type %d", c.Type)
}

func (c *svcCode) UnmarshalJSON(b []byte) (e error) {
	var m map[string]uint8
	if e = json.Unmarshal(b, &m); e != nil {
		return
	}
	for k, v := range m {
		switch k {
		case "ext-BearerService":
			c.Type = 2
		case "ext-Teleservice":
			c.Type = 3
		default:
			e = fmt.Errorf("unknown type %s", k)
		}
		c.Code = v
		break
	}
	return
}

/*
forwInfo

	Ext-ForwInfo ::= SEQUENCE {
		ss-Code                   SS-Code,
		forwardingFeatureList     Ext-ForwFeatureList,
		extensionContainer    [0] ExtensionContainer  OPTIONAL,
		...}
	SS-Code ::= OCTET STRING (SIZE (1))
	Ext-ForwFeatureList ::= SEQUENCE SIZE (1..maxNumOfExt-BasicServiceGroups) OF
		Ext-ForwFeature
*/
type forwInfo struct {
	SsCode   byte          `json:"ss-Code"`
	ForwList []forwFeature `json:"forwardingFeatureList"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (forwInfo) GetType() string { return "forwardingInfo" }

func (i forwInfo) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s| | SS-Code: %x", gsmap.LogPrefix, i.SsCode)
	for n, ff := range i.ForwList {
		fmt.Fprintf(buf, "\n%s| | forwardingFeature[%d]: %s", gsmap.LogPrefix, n, ff)
	}
	return buf.String()
}

func (i forwInfo) marshal() []byte {
	buf := new(bytes.Buffer)

	// ss-Code, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, []byte{i.SsCode})

	// forwardingFeatureList, universal(00) + constructed(20) + sequence(10)
	buf2 := new(bytes.Buffer)
	for _, ff := range i.ForwList {
		// Ext-ForwFeature, universal(00) + constructed(20) + sequence(10)
		gsmap.WriteTLV(buf2, 0x30, ff.marshal())
	}
	gsmap.WriteTLV(buf, 0x30, buf2.Bytes())

	// extensionContainer, context_specific(80) + constructed(20) + 0(00)

	return buf.Bytes()
}

func (i *forwInfo) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	// ss-Code, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return e
	} else if len(v) == 0 {
		return gsmap.UnexpectedTLV("length must >0")
	} else {
		i.SsCode = v[0]
	}

	// forwardingFeatureList, universal(00) + constructed(20) + sequence(10)
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return e
	} else if len(v) == 0 {
		return gsmap.UnexpectedTLV("length must >0")
	} else {
		i.ForwList = []forwFeature{}
		buf2 := bytes.NewBuffer(v)
		for {
			ff := forwFeature{}
			// Ext-ForwFeature, universal(00) + constructed(20) + sequence(10)
			if _, v, e := gsmap.ReadTLV(buf2, 0x30); e == io.EOF {
				break
			} else if e != nil {
				return e
			} else if e = ff.unmarshal(v); e != nil {
				return e
			} else {
				i.ForwList = append(i.ForwList, ff)
			}
		}
	}
	return nil
}

/*
forwFeature

	Ext-ForwFeature ::= SEQUENCE {
		basicService               Ext-BasicServiceCode  OPTIONAL,
		ss-Status             [4]  Ext-SS-Status,
		forwardedToNumber     [5]  ISDN-AddressString    OPTIONAL,
		-- When this data type is sent from an HLR which supports CAMEL Phase 2
		-- to a VLR that supports CAMEL Phase 2 the VLR shall not check the
		-- format of the number
		forwardedToSubaddress [8]  ISDN-SubaddressString OPTIONAL,
		forwardingOptions     [6]  Ext-ForwOptions       OPTIONAL,
		noReplyConditionTime  [7]  Ext-NoRepCondTime     OPTIONAL,
		extensionContainer    [9]  ExtensionContainer    OPTIONAL,
		...,
		longForwardedToNumber [10] FTN-AddressString     OPTIONAL }
	Ext-SS-Status ::= OCTET STRING (SIZE (1..5))
		-- OCTETS 2-5: reserved for future use.
	Ext-ForwOptions ::= OCTET STRING (SIZE (1..5))
		-- OCTETS 2-5: reserved for future use.
	Ext-NoRepCondTime ::= INTEGER (1..100)
		-- Only values 5-30 are used.
*/
type forwFeature struct {
	BasicService          svcCode             `json:"basicService,omitempty"`
	SsStatus              byte                `json:"ss-Status"`
	ForwardedToNumber     teldata.GlobalTitle `json:"forwardedToNumber,omitempty"`
	ForwardedToSubaddress []byte              `json:"forwardedToSubaddress,omitempty"`
	ForwardingOptions     byte                `json:"forwardingOptions,omitempty"`
	NoReplyConditionTime  uint8               `json:"noReplyConditionTime,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	LongForwardedToNumber []byte `json:"longForwardedToNumber,omitempty"`
}

func (f forwFeature) String() string {
	buf := new(strings.Builder)
	if svc := f.BasicService.String(); svc != "" {
		fmt.Fprintf(buf, "\n%s| | | basicService: %s", gsmap.LogPrefix, svc)
	}
	fmt.Fprintf(buf, "\n%s| | | ss-Status: %x", gsmap.LogPrefix, f.SsStatus)
	if !f.ForwardedToNumber.IsEmpty() {
		fmt.Fprintf(buf, "\n%s| | | forwardedToNumber: %s", gsmap.LogPrefix, f.ForwardedToNumber)
	}
	if len(f.ForwardedToSubaddress) != 0 {
		fmt.Fprintf(buf, "\n%s| | | forwardedToSubaddress: %d", gsmap.LogPrefix, f.ForwardedToSubaddress)
	}
	if f.ForwardingOptions != 0 {
		fmt.Fprintf(buf, "\n%s| | | forwardingOptions: %x", gsmap.LogPrefix, f.ForwardingOptions)
	}
	if f.NoReplyConditionTime != 0 {
		fmt.Fprintf(buf, "\n%s| | | noReplyConditionTime: %d", gsmap.LogPrefix, f.NoReplyConditionTime)
	}
	if len(f.LongForwardedToNumber) != 0 {
		fmt.Fprintf(buf, "\n%s| | | longForwardedToNumber: %x", gsmap.LogPrefix, f.LongForwardedToNumber)
	}
	return buf.String()
}

func (f forwFeature) MarshalJSON() ([]byte, error) {
	tmp := struct {
		BasicService          *svcCode             `json:"basicService,omitempty"`
		SsStatus              byte                 `json:"ss-Status"`
		ForwardedToNumber     *teldata.GlobalTitle `json:"forwardedToNumber,omitempty"`
		ForwardedToSubaddress []byte               `json:"forwardedToSubaddress,omitempty"`
		ForwardingOptions     *byte                `json:"forwardingOptions,omitempty"`
		NoReplyConditionTime  *uint8               `json:"noReplyConditionTime,omitempty"`
		// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
		LongForwardedToNumber []byte `json:"longForwardedToNumber,omitempty"`
	}{
		SsStatus:              f.SsStatus,
		ForwardedToSubaddress: f.ForwardedToSubaddress,
		LongForwardedToNumber: f.LongForwardedToNumber,
	}
	if f.BasicService.Type == 2 || f.BasicService.Type == 3 {
		tmp.BasicService = &f.BasicService
	}
	if !f.ForwardedToNumber.IsEmpty() {
		tmp.ForwardedToNumber = &f.ForwardedToNumber
	}
	if f.ForwardingOptions != 0 {
		tmp.ForwardingOptions = &f.ForwardingOptions
	}
	if f.NoReplyConditionTime != 0 {
		tmp.NoReplyConditionTime = &f.NoReplyConditionTime
	}

	return json.Marshal(tmp)
}

func (f *forwFeature) UnmarshalJSON(b []byte) (e error) {
	tmp := struct {
		BasicService          *svcCode             `json:"basicService,omitempty"`
		SsStatus              byte                 `json:"ss-Status"`
		ForwardedToNumber     *teldata.GlobalTitle `json:"forwardedToNumber,omitempty"`
		ForwardedToSubaddress []byte               `json:"forwardedToSubaddress,omitempty"`
		ForwardingOptions     *byte                `json:"forwardingOptions,omitempty"`
		NoReplyConditionTime  *uint8               `json:"noReplyConditionTime,omitempty"`
		// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
		LongForwardedToNumber []byte `json:"longForwardedToNumber,omitempty"`
	}{}
	if e = json.Unmarshal(b, &tmp); e != nil {
		return
	}
	if tmp.BasicService != nil {
		f.BasicService = *tmp.BasicService
	}
	f.SsStatus = tmp.SsStatus
	if tmp.ForwardedToNumber != nil {
		f.ForwardedToNumber = *tmp.ForwardedToNumber
	}
	if tmp.ForwardedToSubaddress != nil {
		f.ForwardedToSubaddress = tmp.ForwardedToSubaddress
	}
	if tmp.ForwardingOptions != nil {
		f.ForwardingOptions = *tmp.ForwardingOptions
	}
	if tmp.NoReplyConditionTime != nil {
		f.NoReplyConditionTime = *tmp.NoReplyConditionTime
	}
	if tmp.LongForwardedToNumber != nil {
		f.LongForwardedToNumber = tmp.LongForwardedToNumber
	}
	return
}

func (f *forwFeature) marshal() []byte {
	buf := new(bytes.Buffer)

	switch f.BasicService.Type {
	case 2: // basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
		gsmap.WriteTLV(buf, 0x82, []byte{f.BasicService.Code})
	case 3: // basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
		gsmap.WriteTLV(buf, 0x83, []byte{f.BasicService.Code})
	}

	// ss-Status, context_specific(80) + primitive(00) + 4(04)
	gsmap.WriteTLV(buf, 0x84, []byte{f.SsStatus})

	// forwardedToNumber, context_specific(80) + primitive(00) + 5(05)
	// forwardedToSubaddress, context_specific(80) + primitive(00) + 8(08)
	// forwardingOptions, context_specific(80) + primitive(00) + 6(06)
	// noReplyConditionTime, context_specific(80) + primitive(00) + 7(07)
	// extensionContainer, context_specific(80) + constructed(20) + 9(09)
	// longForwardedToNumber, context_specific(80) + primitive(00) + 10(0a)

	return buf.Bytes()
}

func (f *forwFeature) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	if t == 0x82 {
		// basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
		if len(v) == 0 {
			return gsmap.UnexpectedTLV("length must >0")
		}
		f.BasicService.Type = 2
		f.BasicService.Code = v[0]

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	} else if t == 0x83 {
		// basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
		if len(v) == 0 {
			return gsmap.UnexpectedTLV("length must >0")
		}
		f.BasicService.Type = 3
		f.BasicService.Code = v[0]

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// ss-Status, context_specific(80) + primitive(00) + 4(04)
	if t == 0x84 {
		if len(v) == 0 {
			return gsmap.UnexpectedTLV("length must >0")
		}
		f.SsStatus = v[0]
		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return nil
			} else if e != nil {
				return e
			}
		*/
	}

	// forwardedToNumber, context_specific(80) + primitive(00) + 5(05)
	// forwardedToSubaddress, context_specific(80) + primitive(00) + 8(08)
	// forwardingOptions, context_specific(80) + primitive(00) + 6(06)
	// noReplyConditionTime, context_specific(80) + primitive(00) + 7(07)
	// extensionContainer, context_specific(80) + constructed(20) + 9(09)
	// longForwardedToNumber, context_specific(80) + primitive(00) + 10(0a)

	return nil
}

/*
callBarInfo

	Ext-CallBarInfo ::= SEQUENCE {
		ss-Code                SS-Code,
		callBarringFeatureList Ext-CallBarFeatureList,
		extensionContainer     ExtensionContainer     OPTIONAL,
		...}
	Ext-CallBarFeatureList ::= SEQUENCE SIZE (1..maxNumOfExt-BasicServiceGroups) OF
		Ext-CallBarringFeature
*/
type callBarInfo struct {
	SsCode      byte             `json:"ss-Code"`
	CallBarList []callBarFeature `json:"callBarringFeatureList"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (callBarInfo) GetType() string { return "callBarringInfo" }

func (i callBarInfo) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s| | SS-Code: %x", gsmap.LogPrefix, i.SsCode)
	for n, ff := range i.CallBarList {
		fmt.Fprintf(buf, "\n%s| | callBarringFeature[%d]: %s", gsmap.LogPrefix, n, ff)
	}
	return buf.String()
}

func (i callBarInfo) marshal() []byte {
	buf := new(bytes.Buffer)

	// ss-Code, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, []byte{i.SsCode})

	// callBarringFeatureList, universal(00) + constructed(20) + sequence(10)
	buf2 := new(bytes.Buffer)
	for _, ff := range i.CallBarList {
		// Ext-CallBarringFeature, universal(00) + constructed(20) + sequence(10)
		gsmap.WriteTLV(buf2, 0x30, ff.marshal())
	}
	gsmap.WriteTLV(buf, 0x30, buf2.Bytes())

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return buf.Bytes()
}

func (i *callBarInfo) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	// ss-Code, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return e
	} else if len(v) == 0 {
		return gsmap.UnexpectedTLV("length must >0")
	} else {
		i.SsCode = v[0]
	}

	// callBarringFeatureList, universal(00) + constructed(20) + sequence(10)
	if _, v, e := gsmap.ReadTLV(buf, 0x30); e != nil {
		return e
	} else if len(v) == 0 {
		return gsmap.UnexpectedTLV("length must >0")
	} else {
		i.CallBarList = []callBarFeature{}
		buf2 := bytes.NewBuffer(v)
		for {
			ff := callBarFeature{}
			// Ext-CallBarringFeature, universal(00) + constructed(20) + sequence(10)
			if _, v, e := gsmap.ReadTLV(buf2, 0x30); e == io.EOF {
				break
			} else if e != nil {
				return e
			} else if e = ff.unmarshal(v); e != nil {
				return e
			} else {
				i.CallBarList = append(i.CallBarList, ff)
			}
		}
	}
	return nil
}

/*
callBarFeature

	Ext-CallBarringFeature ::= SEQUENCE {
		basicService           Ext-BasicServiceCode OPTIONAL,
		ss-Status          [4] Ext-SS-Status,
		extensionContainer     ExtensionContainer   OPTIONAL,
		...}
*/
type callBarFeature struct {
	BasicService svcCode `json:"basicService,omitempty"`
	SsStatus     byte    `json:"ss-Status"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (f callBarFeature) String() string {
	buf := new(strings.Builder)
	if svc := f.BasicService.String(); svc != "" {
		fmt.Fprintf(buf, "\n%s| | | basicService: %s", gsmap.LogPrefix, svc)
	}
	fmt.Fprintf(buf, "\n%s| | | ss-Status: %x", gsmap.LogPrefix, f.SsStatus)
	return buf.String()
}

func (f callBarFeature) MarshalJSON() ([]byte, error) {
	tmp := struct {
		BasicService *svcCode `json:"basicService,omitempty"`
		SsStatus     byte     `json:"ss-Status"`
		// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	}{
		SsStatus: f.SsStatus,
	}
	if f.BasicService.Type == 2 || f.BasicService.Type == 3 {
		tmp.BasicService = &f.BasicService
	}

	return json.Marshal(tmp)
}

func (f *callBarFeature) UnmarshalJSON(b []byte) (e error) {
	tmp := struct {
		BasicService *svcCode `json:"basicService,omitempty"`
		SsStatus     byte     `json:"ss-Status"`
		// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
	}{}
	if e = json.Unmarshal(b, &tmp); e != nil {
		return
	}
	if tmp.BasicService != nil {
		f.BasicService = *tmp.BasicService
	}
	f.SsStatus = tmp.SsStatus
	return
}

func (f *callBarFeature) marshal() []byte {
	buf := new(bytes.Buffer)

	switch f.BasicService.Type {
	case 2: // basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
		gsmap.WriteTLV(buf, 0x82, []byte{f.BasicService.Code})
	case 3: // basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
		gsmap.WriteTLV(buf, 0x83, []byte{f.BasicService.Code})
	}

	// ss-Status, context_specific(80) + primitive(00) + 4(04)
	gsmap.WriteTLV(buf, 0x84, []byte{f.SsStatus})

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return buf.Bytes()
}

func (f *callBarFeature) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	if t == 0x82 {
		// basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
		if len(v) == 0 {
			return gsmap.UnexpectedTLV("length must >0")
		}
		f.BasicService.Type = 2
		f.BasicService.Code = v[0]

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	} else if t == 0x83 {
		// basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
		if len(v) == 0 {
			return gsmap.UnexpectedTLV("length must >0")
		}
		f.BasicService.Type = 3
		f.BasicService.Code = v[0]

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// ss-Status, context_specific(80) + primitive(00) + 4(04)
	if t == 0x84 {
		if len(v) == 0 {
			return gsmap.UnexpectedTLV("length must >0")
		}
		f.SsStatus = v[0]
		/*
			if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
				return nil
			} else if e != nil {
				return e
			}
		*/
	}

	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return nil
}

/*
cugInfo

	CUG-Info ::= SEQUENCE {
		cug-SubscriptionList     CUG-SubscriptionList,
		cug-FeatureList          CUG-FeatureList      OPTIONAL,
		extensionContainer   [0] ExtensionContainer   OPTIONAL,
		...}
	CUG-SubscriptionList ::= SEQUENCE SIZE (0..maxNumOfCUG) OF
		CUG-Subscription
	CUG-FeatureList ::= SEQUENCE SIZE (1..maxNumOfExt-BasicServiceGroups) OF
		CUG-Feature
*/
type cugInfo struct {
	SubscriptionList []cugSubscription `json:"cug-SubscriptionList"`
	FeatureList      []cugFeature      `json:"cug-FeatureList,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (cugInfo) GetType() string { return "cug-Info" }

func (i cugInfo) String() string {
	buf := new(strings.Builder)
	for n, ff := range i.SubscriptionList {
		fmt.Fprintf(buf, "\n%s| | cug-SubscriptionList[%d]: %v", gsmap.LogPrefix, n, ff)
	}
	for n, ff := range i.FeatureList {
		fmt.Fprintf(buf, "\n%s| | cug-FeatureList[%d]: %v", gsmap.LogPrefix, n, ff)
	}
	return buf.String()
}

func (i cugInfo) marshal() []byte {
	buf := new(bytes.Buffer)
	// cug-SubscriptionList, universal(00) + constructed(20) + sequence(10)
	// cug-FeatureList, universal(00) + constructed(20) + sequence(10)
	// extensionContainer, context_specific(80) + constructed(20) + 0(00)

	return buf.Bytes()
}

func (i *cugInfo) unmarshal(_ []byte) error {
	// buf := bytes.NewBuffer(data)

	// cug-SubscriptionList, universal(00) + constructed(20) + sequence(10)
	// cug-FeatureList, universal(00) + constructed(20) + sequence(10)
	// extensionContainer, context_specific(80) + constructed(20) + 0(00)

	return nil
}

/*
cugSubscription

	CUG-Subscription ::= SEQUENCE {
		cug-Index                 CUG-Index,
		cug-Interlock             CUG-Interlock,
		intraCUG-Options          IntraCUG-Options,
		basicServiceGroupList     Ext-BasicServiceGroupList OPTIONAL,
		extensionContainer    [0] ExtensionContainer        OPTIONAL,
		...}
	CUG-Index ::= INTEGER (0..32767)
	CUG-Interlock ::= OCTET STRING (SIZE (4))
	IntraCUG-Options ::= ENUMERATED {
		noCUG-Restrictions (0),
		cugIC-CallBarred   (1),
		cugOG-CallBarred   (2)}
	Ext-BasicServiceGroupList ::= SEQUENCE SIZE (1..maxNumOfExt-BasicServiceGroups) OF
		Ext-BasicServiceCode
*/
type cugSubscription struct {
	Index            int16     `json:"cug-Index"`
	Interlock        [4]byte   `json:"cug-Interlock"`
	IntraOption      byte      `json:"intraCUG-Options"`
	BasicServiceList []svcCode `json:"basicServiceGroupList,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

/*
cugFeature

	CUG-Feature ::= SEQUENCE {
		basicService              Ext-BasicServiceCode  OPTIONAL,
		preferentialCUG-Indicator CUG-Index             OPTIONAL,
		interCUG-Restrictions     InterCUG-Restrictions,
		extensionContainer        ExtensionContainer    OPTIONAL,
		...}
	InterCUG-Restrictions ::= OCTET STRING (SIZE (1))
*/
type cugFeature struct {
	BasicService          svcCode `json:"basicService,omitempty"`
	PreferentialIndicator int16   `json:"preferentialCUG-Indicator,omitempty"`
	IntraRestriction      byte    `json:"interCUG-Restrictions"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

/*
ssData

	Ext-SS-Data ::= SEQUENCE {
		ss-Code                   SS-Code,
		ss-Status             [4] Ext-SS-Status,
		ss-SubscriptionOption     SS-SubscriptionOption     OPTIONAL,
		basicServiceGroupList     Ext-BasicServiceGroupList OPTIONAL,
		extensionContainer    [5] ExtensionContainer        OPTIONAL,
		...}
	SS-SubscriptionOption ::= CHOICE {
		cliRestrictionOption [2] CliRestrictionOption,
		overrideCategory     [1] OverrideCategory}
	CliRestrictionOption ::= ENUMERATED {
		permanent                  (0),
		temporaryDefaultRestricted (1),
		temporaryDefaultAllowed    (2)}
	OverrideCategory ::= ENUMERATED {
		overrideEnabled  (0),
		overrideDisabled (1)}
	BasicServiceGroupList ::= SEQUENCE SIZE (1..maxNumOfBasicServiceGroups) OF
		BasicServiceCode
*/
type ssData struct {
	SsCode            byte      `json:"ss-Code"`
	SsStatus          byte      `json:"ss-Status"`
	SsSubscriptionOpt []byte    `json:"ss-SubscriptionOption,omitempty"`
	BasicServiceList  []svcCode `json:"basicServiceGroupList,omitempty"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (ssData) GetType() string { return "ss-Data" }

func (i ssData) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s| | SS-Code: %x", gsmap.LogPrefix, i.SsCode)
	fmt.Fprintf(buf, "\n%s| | SS-Status: %x", gsmap.LogPrefix, i.SsStatus)
	if len(i.SsSubscriptionOpt) != 0 {
		fmt.Fprintf(buf, "\n%s| | ss-SubscriptionOption: %x", gsmap.LogPrefix, i.SsSubscriptionOpt)
	}
	for _, s := range i.BasicServiceList {
		fmt.Fprintf(buf, "\n%s| | basicServiceGroup: %x", gsmap.LogPrefix, s)
	}
	return buf.String()
}

func (i ssData) marshal() []byte {
	buf := new(bytes.Buffer)

	// ss-Code, universal(00) + primitive(00) + octet_string(04)
	gsmap.WriteTLV(buf, 0x04, []byte{i.SsCode})

	// ss-Status, context_specific(80) + primitive(00) + 4(04)
	gsmap.WriteTLV(buf, 0x84, []byte{i.SsStatus})

	// ss-SubscriptionOption, universal(00) + constructed(20) + octet_string(04)
	if len(i.SsSubscriptionOpt) != 0 {
		gsmap.WriteTLV(buf, 0x24, i.SsSubscriptionOpt)
	}

	// basicServiceGroupList, universal(00) + constructed(20) + sequence(10)
	if len(i.BasicServiceList) != 0 {
		buf2 := new(bytes.Buffer)
		for _, sc := range i.BasicServiceList {
			switch sc.Type {
			case 2: // basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
				gsmap.WriteTLV(buf2, 0x82, []byte{sc.Code})
			case 3: // basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
				gsmap.WriteTLV(buf2, 0x83, []byte{sc.Code})
			}
		}
		gsmap.WriteTLV(buf, 0x30, buf2.Bytes())
	}

	// extensionContainer, context_specific(80) + constructed(20) + 5(05)

	return buf.Bytes()
}

func (i *ssData) unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	// ss-Code, universal(00) + primitive(00) + octet_string(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x04); e != nil {
		return e
	} else if len(v) == 0 {
		return gsmap.UnexpectedTLV("length must >0")
	} else {
		i.SsCode = v[0]
	}

	// ss-Status, context_specific(80) + primitive(00) + 4(04)
	if _, v, e := gsmap.ReadTLV(buf, 0x84); e != nil {
		return e
	} else if len(v) == 0 {
		return gsmap.UnexpectedTLV("length must >0")
	} else {
		i.SsStatus = v[0]
	}

	t, v, e := gsmap.ReadTLV(buf, 0x00)
	if e == io.EOF {
		return nil
	} else if e != nil {
		return e
	}

	// ss-SubscriptionOption, universal(00) + constructed(20) + octet_string(04)
	if t == 0x24 {
		i.SsSubscriptionOpt = v

		if t, v, e = gsmap.ReadTLV(buf, 0x00); e == io.EOF {
			return nil
		} else if e != nil {
			return e
		}
	}

	// basicServiceGroupList, universal(00) + constructed(20) + sequence(10)
	if t == 0x30 {
		i.BasicServiceList = []svcCode{}
		buf2 := bytes.NewBuffer(v)
		for {
			t2, v2, e := gsmap.ReadTLV(buf2, 0x00)
			if e == io.EOF {
				break
			} else if e != nil {
				return e
			}

			if t2 == 0x82 {
				// basicService ext-BearerService, context_specific(80) + primitive(00) + 2(02)
				if len(v2) == 0 {
					return gsmap.UnexpectedTLV("length must >0")
				}
				i.BasicServiceList = append(i.BasicServiceList,
					svcCode{Type: 2, Code: v2[0]})
			} else if t2 == 0x83 {
				// basicService ext-Teleservice, context_specific(80) + primitive(00) + 3(03)
				if len(v2) == 0 {
					return gsmap.UnexpectedTLV("length must >0")
				}
				i.BasicServiceList = append(i.BasicServiceList,
					svcCode{Type: 3, Code: v2[0]})
			}
		}
	}
	return nil
}

/*
emlppInfo

	EMLPP-Info ::= SEQUENCE {
		maximumentitledPriority EMLPP-Priority,
		defaultPriority         EMLPP-Priority,
		extensionContainer      ExtensionContainer OPTIONAL,
		...}
	EMLPP-Priority ::= INTEGER (0..15)
*/
type emlppInfo struct {
	MaxPriority uint8 `json:"maximumentitledPriority"`
	Priority    uint8 `json:"defaultPriority"`
	// Extension ExtensionContainer `json:"extensionContainer,omitempty"`
}

func (emlppInfo) GetType() string { return "emlpp-Info" }

func (i emlppInfo) String() string {
	buf := new(strings.Builder)
	fmt.Fprintf(buf, "%s| | maximumentitledPriority: %d", gsmap.LogPrefix, i.MaxPriority)
	fmt.Fprintf(buf, "\n%s| | defaultPriority: %d", gsmap.LogPrefix, i.Priority)
	return buf.String()
}

func (i emlppInfo) marshal() []byte {
	buf := new(bytes.Buffer)
	// maximumentitledPriority, universal(00) + primitive(00) + integer(02)
	// defaultPriority, universal(00) + primitive(00) + integer(02)
	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return buf.Bytes()
}

func (i *emlppInfo) unmarshal(_ []byte) error {
	// buf := bytes.NewBuffer(data)

	// maximumentitledPriority, universal(00) + primitive(00) + integer(02)
	// defaultPriority, universal(00) + primitive(00) + integer(02)
	// extensionContainer, universal(00) + constructed(20) + sequence(10)

	return nil
}
