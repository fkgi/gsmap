package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
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
	asp     = xua.ASP{}
	backend string
	verbose *bool
)

func main() {
	l := flag.String("l", "", "local address")
	p := flag.String("p", "", "peer address")
	rc := flag.Uint("r", 0, "routing context")
	ppc := flag.Uint("c", 0, "peer point code")
	lpc := flag.Uint("d", 0, "local point code")
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
	pa, e := xua.ParseSCTPAddr(*p)
	if e != nil {
		log.Fatalln("[ERROR]", "invalid peer address:", e)
	}

	xua.LocalAddr.GlobalTitle.NatureOfAddress = teldata.International
	xua.LocalAddr.GlobalTitle.NumberingPlan = teldata.ISDNTelephony
	if xua.LocalAddr.GlobalTitle.Digits, e = teldata.ParseTBCD(*gt); e != nil {
		log.Fatalln("[ERROR]", "invalid global title address:", e)
	}
	switch *ssn {
	case "msc":
		xua.LocalAddr.SubsystemNumber = teldata.SsnMSC
	case "hlr":
		xua.LocalAddr.SubsystemNumber = teldata.SsnHLR
	case "vlr":
		xua.LocalAddr.SubsystemNumber = teldata.SsnVLR
	case "":
	default:
		log.Fatalln("[ERROR]", "invalid subsystem number")
	}

	xua.LocalPC = uint32(*lpc)
	asp.PointCode = uint32(*ppc)
	asp.Context = uint32(*rc)

	if asp.PointCode != 0 && xua.LocalPC == 0 {
		log.Fatalln("[ERROR]", "local point code is not specified")
	}
	if asp.Context == 0 {
		log.Fatalln("[ERROR]", "routing context is not specified")
	}

	tcap.SelectASP = func() *xua.ASP {
		return &asp
	}

	tcap.Tw = time.Duration(*to) * time.Second

	backend = "http://" + *be
	_, e = url.Parse(backend)
	if e != nil || len(*be) == 0 {
		log.Println("[ERROR]", "invalid HTTP backend host, MAP request will be rejected")
		backend = ""
	} else {
		log.Println("[INFO]", "HTTP backend is", backend)
		var dt *http.Transport
		if t, ok := http.DefaultTransport.(*http.Transport); ok {
			dt = t.Clone()
		} else {
			dt = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				ForceAttemptHTTP2:     false,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}
		dt.MaxIdleConns = 0
		dt.MaxIdleConnsPerHost = 1000
		client = http.Client{
			Transport: dt,
			Timeout:   tcap.Tw}
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
		return &tcap.AARE{
			Context:   q.Context,
			Result:    tcap.RejectPermanent,
			ResultSrc: tcap.SrcUsrNull,
		}
	}

	/*
		for k := range gsmap.NameMap {
			log.Println("[INFO]", "available component", k)
		}
	*/

	http.HandleFunc("/mapmsg/v1/", handleOutgoingDialog)
	http.HandleFunc("/dialog/", handleContinueDialog)
	http.HandleFunc("/mapstate/v1/connection", conStateHandler)
	http.HandleFunc("/mapstate/v1/statistics", statsHandler)
	go func() {
		log.Fatalln(http.ListenAndServe(*api, nil))
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sigc
		if asp.State() == xua.Down {
			os.Exit(0)
		} else {
			asp.Close()
		}
	}()

	log.Println("[INFO]", "Connecting ASP")
	log.Println("[INFO]", "closed, error=", asp.DialAndServe(la, pa))
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
			e = json.Unmarshal(v, cgpa)
		default:
			if c, ok := gsmap.NameMap[k]; !ok {
				e = errors.New("unknown component: " + k)
			} else if c, e = c.NewFromJSON(v, defaultID); e == nil {
				cpnt = append(cpnt, c)
			}
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
