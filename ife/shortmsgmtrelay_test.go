package ife_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/ife"
	"github.com/fkgi/teldata"
)

func TestMarshalJSONMT(t *testing.T) {
	imsi, _ := teldata.ParseIMSI("1234567")
	gt := gsmap.AddressString{
		NatureOfAddress: teldata.International,
		NumberingPlan:   teldata.ISDNTelephony}
	gt.Digits, _ = teldata.ParseTBCD("0987")
	mt := ife.MTForwardSMArg{
		SMRPDA: ife.RpIMSI{IMSI: imsi},
		SMRPOA: ife.RpSCAddress{SCAddr: gt},
	}
	b, e := json.MarshalIndent(mt, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	fmt.Println(string(b))
	fmt.Println(mt)

	mt = ife.MTForwardSMArg{}
	e = json.Unmarshal(b, &mt)
	if e != nil {
		t.Fatal(e)
	}
	fmt.Println(mt)
}

func TestMarshalJSONMO(t *testing.T) {
	msisdn := gsmap.AddressString{
		NatureOfAddress: teldata.International,
		NumberingPlan:   teldata.ISDNTelephony}
	msisdn.Digits, _ = teldata.ParseTBCD("819011110001")
	gt := gsmap.AddressString{
		NatureOfAddress: teldata.International,
		NumberingPlan:   teldata.ISDNTelephony}
	gt.Digits, _ = teldata.ParseTBCD("0987")
	mt := ife.MOForwardSMArg{
		SMRPDA: ife.RpSCAddress{SCAddr: gt},
		SMRPOA: ife.RpMSISDN{MSISDN: msisdn},
	}
	b, e := json.MarshalIndent(mt, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	fmt.Println(string(b))
	fmt.Println(mt)

	mt = ife.MOForwardSMArg{}
	e = json.Unmarshal(b, &mt)
	if e != nil {
		t.Fatal(e)
	}
	fmt.Println(mt)
}
