package xua

import (
	"errors"
	"io"
)

/*
Message of xUA

	 0                   1                   2                   3
	 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|    Version    |   Reserved    | Message Class | Message Type  |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                        Message Length                         |
	+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	|                         Message Data                          |
*/
type message interface {
	// handle this message
	handle(*ASP) error
	// marshal returns Message Class, Message Type and binary Message Data
	marshal() (uint8, uint8, []byte)
	// unmarshal decodes specified Tag/length TLV value from reader
	unmarshal(uint16, uint16, io.ReadSeeker) error

	name() string
	state() string
}

func getMessage(c, t byte) message {
	switch c {
	case 0x00:
		switch t {
		case 0x00:
			return new(ERR)
		case 0x01:
			return new(NTFY)
		}
	case 0x01:
		switch t {
		case 0x01:
			return new(DATA)
		}
	case 0x02:
		switch t {
		case 0x01:
			return new(DUNA)
		case 0x02:
			return new(DAVA)
		case 0x04:
			return new(SCON)
		case 0x05:
			return new(DUPU)
		case 0x06:
			return new(DRST)
		}
	case 0x03:
		switch t {
		case 0x01:
			return new(ASPUP)
		case 0x02:
			return new(ASPDN)
		case 0x03:
			return new(BEAT)
		case 0x04:
			return new(ASPUPAck)
		case 0x05:
			return new(ASPDNAck)
		case 0x06:
			return new(BEATAck)
		}
	case 0x04:
		switch t {
		case 0x01:
			return new(ASPAC)
		case 0x02:
			return new(ASPIA)
		case 0x03:
			return new(ASPACAck)
		case 0x04:
			return new(ASPIAAck)
		}
	}
	return nil
}

type timeout struct {
	srcMessage message
}

func (m *timeout) handle(c *ASP) error {
	if c.state != m.srcMessage {
		return nil
	}

	switch m := m.srcMessage.(type) {
	case *ASPUP:
		if m.retrans++; m.retrans > MaxRetrans {
			m.result <- errors.New("timeout")
		} else {
			c.msgQ <- m
		}
	case *ASPDN:
		if m.retrans++; m.retrans > MaxRetrans {
			m.result <- errors.New("timeout")
		} else {
			c.msgQ <- m
		}
	case *ASPAC:
		if m.retrans++; m.retrans > MaxRetrans {
			m.result <- errors.New("timeout")
		} else {
			c.msgQ <- m
		}
	case *ASPIA:
		if m.retrans++; m.retrans > MaxRetrans {
			m.result <- errors.New("timeout")
		} else {
			c.msgQ <- m
		}
	case *BEAT:
		if m.retrans++; m.retrans > MaxRetrans {
			m.result <- errors.New("timeout")
		} else {
			c.msgQ <- m
		}
	default:
		return errors.New("unexpected message")
	}
	return nil
}

func (*timeout) marshal() (uint8, uint8, []byte)               { return 0, 0, nil }
func (*timeout) unmarshal(uint16, uint16, io.ReadSeeker) error { return nil }
func (*timeout) name() string                                  { return "timeout" }
func (*timeout) state() string                                 { return "" }

type down struct {
	e error
}

func (m *down) handle(c *ASP) error {
	c.state = &down{}
	sockClose(c.sock)
	close(c.msgQ)
	return nil
}
func (*down) marshal() (uint8, uint8, []byte)               { return 0, 0, nil }
func (*down) unmarshal(uint16, uint16, io.ReadSeeker) error { return nil }
func (*down) name() string                                  { return "down" }
func (*down) state() string                                 { return "down" }

type inactive struct{}

func (*inactive) handle(*ASP) error                             { return nil }
func (*inactive) marshal() (uint8, uint8, []byte)               { return 0, 0, nil }
func (*inactive) unmarshal(uint16, uint16, io.ReadSeeker) error { return nil }
func (*inactive) name() string                                  { return "" }
func (*inactive) state() string                                 { return "inactive" }

type active struct{}

func (*active) handle(*ASP) error                             { return nil }
func (*active) marshal() (uint8, uint8, []byte)               { return 0, 0, nil }
func (*active) unmarshal(uint16, uint16, io.ReadSeeker) error { return nil }
func (*active) name() string                                  { return "" }
func (*active) state() string                                 { return "active" }

type closing struct{}

func (*closing) handle(*ASP) error                             { return nil }
func (*closing) marshal() (uint8, uint8, []byte)               { return 0, 0, nil }
func (*closing) unmarshal(uint16, uint16, io.ReadSeeker) error { return nil }
func (*closing) name() string                                  { return "" }
func (*closing) state() string                                 { return "closing" }
