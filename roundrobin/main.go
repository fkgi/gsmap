package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/tcap"
	"github.com/fkgi/gsmap/xua"
	"github.com/fkgi/teldata"
)

var (
	backend string
	verbose *bool
)

type list []string

func (a *list) String() string {
	return fmt.Sprintf("%v", *a)
}
func (a *list) Set(v string) error {
	*a = append(*a, v)
	return nil
}

func main() {
	var peerAddrs list
	l := flag.String("l", "", "local address")
	flag.Var(&peerAddrs, "p", "peer address")
	rc := flag.Uint("r", 0, "routing context")
	ppc := flag.Uint("c", 0, "peer point code")
	lpc := flag.Uint("d", 0, "local point code")
	ni := flag.Uint("i", 0, "network indicator")
	na := flag.Int("n", -1, "network appearance")
	gt := flag.String("g", "", "global title address")
	ssn := flag.String("s", "", "subsystem number msc|hlr|vlr")
	api := flag.String("a", ":8080", "local API port")
	be := flag.String("b", "localhost:80", "backend API port")
	to := flag.Int("t", int(tcap.Tw/time.Second), "Message timeout timer [s]")
	verbose = flag.Bool("v", false, "Verbose log output")
	flag.Parse()

	log.Println("[INFO]", "booting Round-Robin debugger for MAP...")

	la, e := xua.ParseSCTPAddr(*l)
	if e != nil {
		log.Fatalln("[ERROR]", "invalid local address:", e)
	}
	pa := make([]*xua.SCTPAddr, 0, len(peerAddrs))
	for _, a := range peerAddrs {
		p, e := xua.ParseSCTPAddr(a)
		if e != nil {
			log.Fatalln("[ERROR]", "invalid peer address:", e)
		}
		pa = append(pa, p)
	}

	if *lpc == 0 || *ppc == 0 {
		log.Fatalln("[ERROR]", "point code is not specified")
	}
	if *rc == 0 {
		log.Fatalln("[ERROR]", "routing context is not specified")
	}

	tcap.EndPoint, e = xua.NewSignalingEndpoint(la)
	if e != nil {
		log.Fatalln("[ERROR]", "failed to bind:", e)
	}
	tcap.EndPoint.ReturnOnError = true
	if *ni > 0 && *ni < 4 {
		tcap.EndPoint.NetIndicator = uint8(*ni)
	}
	if *na >= 0 {
		tmp := uint32(*na)
		tcap.EndPoint.NetAppearance = &tmp
	}
	tcap.EndPoint.PayloadHandler = tcap.HandlePayload

	tcap.EndPoint.GlobalTitle.NatureOfAddress = teldata.International
	tcap.EndPoint.GlobalTitle.NumberingPlan = teldata.ISDNTelephony
	if tcap.EndPoint.GlobalTitle.Digits, e = teldata.ParseTBCD(*gt); e != nil {
		log.Fatalln("[ERROR]", "invalid global title address:", e)
	}
	switch *ssn {
	case "msc":
		tcap.EndPoint.SubsystemNumber = teldata.SsnMSC
	case "hlr":
		tcap.EndPoint.SubsystemNumber = teldata.SsnHLR
	case "vlr":
		tcap.EndPoint.SubsystemNumber = teldata.SsnVLR
	case "":
	default:
		log.Fatalln("[ERROR]", "invalid subsystem number")
	}

	tcap.EndPoint.PointCode = uint32(*lpc)
	tcap.EndPoint.Context = uint32(*rc)
	tcap.PeerPointCode = uint32(*ppc)

	tcap.Tw = time.Duration(*to) * time.Second

	backend = "http://" + *be
	_, e = url.Parse(backend)
	if e != nil || len(*be) == 0 {
		log.Println("[ERROR]", "invalid HTTP backend host, MAP request will be rejected")
		backend = ""
	} else {
		log.Println("[INFO]", "HTTP backend is", backend)
		t, _ := http.DefaultTransport.(*http.Transport)
		dt := t.Clone()
		dt.MaxIdleConns = 0
		dt.MaxIdleConnsPerHost = 1000
		client = http.Client{
			Transport: dt,
			Timeout:   tcap.Tw,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			}}
		tcap.NewInvoke = handleIncomingDialog
	}

	if !*verbose {
		tcap.TraceMessage = nil
	}

	tcap.DialogueHandler = func(q tcap.AARQ) tcap.Dialogue {
		n, v := getContextName(q.Context)
		if n != "" && v != "" {
			return &tcap.AARE{
				Context:   q.Context,
				Result:    tcap.Accept,
				ResultSrc: tcap.SrcUsrNull}
		}
		log.Println("[INFO]", "unsupported application context is required: ", q.Context)
		return &tcap.ABRT{Source: tcap.SvcUser}
	}

	/*
		for k := range gsmap.NameMap {
			log.Println("[INFO]", "available component", k)
		}
	*/

	http.HandleFunc("POST /mapmsg/v1/{ac}/{ver}", handleOutgoingDialog)
	http.HandleFunc("POST /dialog/{id}", handleContinueDialog)
	http.HandleFunc("DELETE /dialog/{id}", handleContinueDialog)
	http.HandleFunc("GET /mapstate/v1/connection", conStateHandler)
	http.HandleFunc("GET /mapstate/v1/statistics", statsHandler)
	go func() {
		log.Fatalln(http.ListenAndServe(*api, nil))
	}()

	log.Println("[INFO]", "Connecting ASP")
	for _, a := range pa {
		if e = tcap.EndPoint.ConnectTo(a); e != nil {
			log.Fatalln("ERROR", "failed to connect ASP:", e)
		}
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigc
	tcap.EndPoint.Close()
}

func readFromJSON(d []byte, defaultID int8) (cdpa xua.SCCPAddr, cgpa *xua.SCCPAddr, cpnt []gsmap.Component, e error) {
	data := map[string]json.RawMessage{}
	if e = json.Unmarshal(d, &data); e != nil {
		return
	}

	cpnt = []gsmap.Component{}
	for k, v := range data {
		switch k {
		case "cdpa":
			e = json.Unmarshal(v, &cdpa)
		case "cgpa":
			cgpa = &xua.SCCPAddr{}
			e = json.Unmarshal(v, cgpa)
		default:
			if c, ok := gsmap.NameMap[k]; !ok {
				e = errors.New("unknown component: " + k)
			} else if c, e = c.NewFromJSON(v, defaultID); e != nil {
			} else if _, ok := c.(gsmap.Invoke); ok {
				cpnt = append(cpnt, c)
			} else {
				cpnt = append([]gsmap.Component{c}, cpnt...)
			}
			defaultID++
		}
		if e != nil {
			return
		}
	}
	return
}

func writeToJSON(cdpa, cgpa *xua.SCCPAddr, cpnt []gsmap.Component) ([]byte, error) {
	var e error
	data := map[string]json.RawMessage{}
	if cdpa != nil {
		if data["cdpa"], e = json.Marshal(*cdpa); e != nil {
			return nil, e
		}
	}
	if cgpa != nil {
		if data["cgpa"], e = json.Marshal(cgpa); e != nil {
			return nil, e
		}
	}
	for _, r := range cpnt {
		if data[r.Name()], e = json.Marshal(r); e != nil {
			return nil, e
		}
	}
	return json.Marshal(data)
}
