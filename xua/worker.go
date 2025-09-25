package xua

import (
	"time"
)

const (
	minWorkers = 128
	maxWorkers = 65535 - minWorkers
)

var (
	sharedQ        = make(chan userData, maxWorkers)
	PayloadHandler func(SCCPAddr, SCCPAddr, []byte)
)

type userData struct {
	cgpa SCCPAddr
	cdpa SCCPAddr
	data []byte
}

func init() {
	var activeWorkers = make(chan int, 1)
	activeWorkers <- 0

	worker := func() {
		for c := 0; c < 500; {
			if len(sharedQ) < minWorkers {
				time.Sleep(time.Millisecond * 10)
				c++
				continue
			}
			if req, ok := <-sharedQ; !ok {
				break
			} else {
				PayloadHandler(req.cgpa, req.cdpa, req.data)
				c = 0
			}
		}
		activeWorkers <- (<-activeWorkers - 1)
	}

	for i := 0; i < minWorkers; i++ {
		go func() {
			for req, ok := <-sharedQ; ok; req, ok = <-sharedQ {
				a := <-activeWorkers
				activeWorkers <- a
				if len(sharedQ) > minWorkers && a < maxWorkers {
					activeWorkers <- (<-activeWorkers + 1)
					go worker()
				}
				PayloadHandler(req.cgpa, req.cdpa, req.data)
			}
		}()
	}
}
