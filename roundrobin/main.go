package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/tcap"
	"github.com/fkgi/gsmap/xua"
	"github.com/fkgi/teldata"
)

var (
	backend string
)

func main() {
	log.Println("[INFO]", "booting Round-Robin debugger for MAP...")

	if v := os.Getenv("VERBOSE"); v != "yes" {
		if v != "no" {
			log.Println("[INFO]", "parameter VERBOSE is empty or invalid, set to default")
		}
		xua.TraceEvent = func(old, new, event string, err error) {
			if err != nil {
				log.Printf("[INFO] event %s handling failed: %v", event, err)
			}
			if old != "inactive" && new == "inactive" {
				log.Println("[INFO]", "ASP state update: inactive")
			} else if old != "active" && new == "active" {
				log.Println("[INFO]", "ASP state update: active")
			}
		}
		tcap.TraceTxMessage = func(m tcap.Message, _ error) { count(m, false) }
		tcap.TraceRxMessage = func(m tcap.Message, _ error) { count(m, true) }
		traceTxDialog = func(_, _ string) {}
		traceRxDialog = func(_, _ string) {}
		traceTxHttpRequest = func(_ string, _ []byte, _ int, _ []byte, err error) {
			if err != nil {
				log.Printf("[INFO] Tx HTTP request handling failed: %v", err)
			}
		}
		traceRxHttpRequest = func(_ string, _ []byte, _ int, _ []byte, err error) {
			if err != nil {
				log.Printf("[INFO] Rx HTTP request handling failed: %v", err)
			}
		}
	}

	if to := os.Getenv("TIMEOUT"); to == "" {
	} else if t, e := strconv.Atoi(to); e != nil {
		log.Printf("[INFO] parameter TIMEOUT is invalid, set to default %fs",
			tcap.Tw.Seconds())
	} else {
		tcap.Tw = time.Second * time.Duration(t)
	}

	la, e := xua.ParseSCTPAddr(os.Getenv("LOCAL_ADDR"))
	if e != nil {
		log.Fatalln("[ERROR]", "invalid local address:", e)
	}

	pa := make([]*xua.SCTPAddr, 0)
	for i := range 10 {
		if a := os.Getenv(fmt.Sprintf("PEER_ADDR%d", i)); a == "" {
			continue
		} else if p, e := xua.ParseSCTPAddr(a); e != nil {
			log.Fatalln("[ERROR]", "invalid peer address of", a, ":", e)
		} else if len(p.IP) == 0 || p.IP[0].To4() == nil {
			log.Fatalln("[ERROR]", "invalid peer address of", a, ": not IPv4")
		} else {
			pa = append(pa, p)
		}
	}

	if len(pa) == 0 {
		tcap.EndPoint, e = xua.NewSignalingTransferPoint(la)
	} else {
		tcap.EndPoint, e = xua.NewSignalingEndPoint(la)
	}
	if e != nil {
		log.Fatalln("[ERROR]", "failed to bind:", e)
	}
	if i, e := strconv.Atoi(os.Getenv("LOCAL_POINT_CODE")); e != nil {
		log.Fatalln("[ERROR]", "invalid local point code")
	} else {
		tcap.EndPoint.LocalPointCode = uint32(i)
	}
	if i, e := strconv.Atoi(os.Getenv("GATEWAY_POINT_CODE")); e != nil {
		log.Fatalln("[ERROR]", "invalid gateway point code")
	} else {
		tcap.EndPoint.GwPointCode = uint32(i)
	}
	if i, e := strconv.Atoi(os.Getenv("ROUTING_CONTEXT")); e != nil {
		tcap.EndPoint.Context = 0
	} else {
		tcap.EndPoint.Context = uint32(i)
	}
	ni := os.Getenv("NETWORK_INDICATOR")
	switch ni {
	case "international":
		tcap.EndPoint.NetIndicator = 0
	case "spare":
		tcap.EndPoint.NetIndicator = 1
	case "national":
		tcap.EndPoint.NetIndicator = 2
	case "reserved":
		tcap.EndPoint.NetIndicator = 3
	default:
		ni = "international"
		tcap.EndPoint.NetIndicator = 0
	}
	if i, e := strconv.Atoi(os.Getenv("NETWORK_APPEARANCE")); e != nil {
		tcap.EndPoint.NetAppearance = 0
	} else {
		tcap.EndPoint.NetAppearance = uint32(i)
	}
	xua.PayloadHandler = tcap.HandlePayload

	tcap.LocalGT.GlobalTitle.NatureOfAddress = teldata.International
	tcap.LocalGT.GlobalTitle.NumberingPlan = teldata.ISDNTelephony
	if tcap.LocalGT.GlobalTitle.Digits, e = teldata.ParseTBCD(os.Getenv("GLOBAL_TITLE")); e != nil {
		log.Fatalln("[ERROR]", "invalid global title address:", e)
	}
	switch os.Getenv("SUBSYSTEM_NUMBER") {
	case "msc":
		tcap.LocalGT.SubsystemNumber = teldata.SsnMSC
	case "hlr":
		tcap.LocalGT.SubsystemNumber = teldata.SsnHLR
	case "vlr":
		tcap.LocalGT.SubsystemNumber = teldata.SsnVLR
	default:
		tcap.LocalGT.SubsystemNumber = teldata.SsnMSC
	}

	log.Printf("[INFO] ASP local information"+
		"\n | address:                     %s"+
		"\n | point code(routing context): %d(%d)"+
		"\n | network indicator:           %s"+
		"\n | network appearance:          %d"+
		"\n | global title/ssn:            %s / %s",
		la,
		tcap.EndPoint.LocalPointCode, tcap.EndPoint.Context,
		ni,
		tcap.EndPoint.NetAppearance,
		tcap.LocalGT.GlobalTitle, tcap.LocalGT.SubsystemNumber)

	backend = "http://" + os.Getenv("BACKENDAPI_ADDR")
	if u, e := url.Parse(backend); e != nil || u.Host == "" {
		log.Println("[WARN]", "invalid HTTP backend host, Rx request will be rejected")
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

	/*
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
	*/

	http.HandleFunc("POST /mapmsg/v1/{ac}/{ver}", handleOutgoingDialog)
	http.HandleFunc("POST /dialog/{id}/continue", handleContinueDialogContinue)
	http.HandleFunc("POST /dialog/{id}/end", handleContinueDialogEnd)
	http.HandleFunc("POST /dialog/{id}/abort", handleContinueDialogAbort)
	// http.HandleFunc("POST /dialog/{id}/{action}", handleContinueDialogPost)
	http.HandleFunc("POST /dialog/{id}", handleContinueDialog)
	http.HandleFunc("DELETE /dialog/{id}", handleContinueDialogDelete)
	http.HandleFunc("GET /metrics", metricsHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		traceRxHttpRequest(r.URL.Path, nil, http.StatusNotFound, nil, nil)
	})

	frontend := os.Getenv("LOCALAPI_ADDR")
	log.Println("[INFO]", "listening HTTP...\n | local port:", frontend)
	go func() {
		err := http.ListenAndServe(frontend, nil)
		if err != nil {
			log.Println("[WARN]", "failed to listen HTTP, Tx request is not available:", err)
		}
	}()

	log.Println("[INFO]", "Connecting ASP...")
	if len(pa) == 0 {
		go func() {
			e := tcap.EndPoint.ListenAndServe()
			log.Println("[WARN]", "transport listener closed:", e)
		}()
	} else {
		for _, a := range pa {
			go tcap.EndPoint.ConnectTo(a)
		}
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigc
	log.Println("[INFO]", "interrupted, closing connections")
	time.AfterFunc(time.Second*30, func() {
		log.Fatalln("[ERROR]", "closing timeout, forcefully stopped")
	})
	tcap.EndPoint.Close()
	log.Println("[INFO]", "server stopped")
}

func readFromJSON(d []byte, defaultID int8) (cdpa, cgpa *xua.SCCPAddr, cpnt []gsmap.Component, e error) {
	data := map[string]json.RawMessage{}
	if e = json.Unmarshal(d, &data); e != nil {
		return
	}

	cpnt = []gsmap.Component{}
	for k, v := range data {
		switch k {
		case "cdpa":
			cdpa = &xua.SCCPAddr{}
			e = json.Unmarshal(v, cdpa)
		case "cgpa":
			cgpa = &xua.SCCPAddr{}
			e = json.Unmarshal(v, cgpa)
		case "EmptyResult":
			var c gsmap.Component
			c, e = tcap.EmptyResult{}.NewFromJSON(v, defaultID)
			if e == nil {
				cpnt = append([]gsmap.Component{c}, cpnt...)
			}
			defaultID++
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
		if data["cdpa"], e = json.Marshal(cdpa); e != nil {
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
