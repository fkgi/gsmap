package main

import (
	"log"

	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/tcap"
	"github.com/fkgi/gsmap/xua"
)

func init() {
	xua.StateNotify = func(id byte, s xua.Status) {
		log.Printf("[INFO] AS state change: id=0x%2x, state=%s", id, s)
	}
	xua.ErrorNotify = func(id byte, c xua.ErrCode) {
		log.Printf("[INFO] AS error: id=0x%2x, state=%s", id, c)
	}
	xua.SctpNotify = func(id byte, s string) {
		log.Printf("[INFO] SCTP: id=0x%2x, %s", id, s)
	}

	tcap.TraceMessage = func(m tcap.Message, d tcap.Direction, err error) {
		log.Printf("[INFO] %s MAP message handling: error=%v\n%s", d, err, m.String())

		switch msg := m.(type) {
		case *tcap.TcBegin, *tcap.TcContinue, *tcap.TcEnd:
			for _, c := range msg.Components() {
				switch c.(type) {
				case gsmap.Invoke:
					if d == tcap.Rx {
						rxInvoke++
					} else {
						txInvoke++
					}
				case gsmap.ReturnResult:
					if d == tcap.Rx {
						rxResult++
					} else {
						txResult++
					}
				case gsmap.ReturnResultLast:
					if d == tcap.Rx {
						rxResultLast++
					} else {
						txResultLast++
					}
				case gsmap.ReturnError:
					if d == tcap.Rx {
						rxError++
					} else {
						txError++
					}
				}
			}
		case *tcap.TcAbort:
			if d == tcap.Rx {
				rxAbort++
			} else {
				txAbort++
			}
		}
	}

	xua.DunaNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Rx DUNA for PC=%v", pc)
	}
	xua.DavaNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Rx DAVA for PC=%v", pc)
	}
	xua.DaudNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Tx DAUD for PC=%v", pc)
	}
	xua.SconNotify = func(pc []xua.PointCode, con uint32) {
		log.Printf("[INFO] Rx SCON for PC=%v, congestion level=%d", pc, con)
	}
	xua.DupuNotify = func(pc []xua.PointCode, cause uint16) {
		log.Printf("[INFO] Rx DUPU for PC=%v, cause=%d", pc, cause)
	}
	xua.DrstNotify = func(pc []xua.PointCode) {
		log.Printf("[INFO] Rx DRST for PC=%v", pc)
	}

}
