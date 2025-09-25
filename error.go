package gsmap

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type UnexpectedTLVError struct {
	s string

	funcName string
	fileName string
	line     int
}

func UnexpectedTag(exp []byte, act byte) UnexpectedTLVError {
	buf := new(strings.Builder)
	for _, b := range exp {
		fmt.Fprintf(buf, "%#x,", b)
	}
	s := buf.String()
	if len(s) != 0 {
		s = s[:len(s)-1]
	}
	return unexpectedTLV(
		fmt.Sprintf("expected tags are [%s] but %#x", s, act), 2)
}

func UnexpectedEnumValue(data []byte) UnexpectedTLVError {
	i, _ := binary.Varint(data)
	return unexpectedTLV(fmt.Sprintf(
		"unexpected enum value %0"+strconv.Itoa(len(data)*2)+"x", i), 2)
}

func UnexpectedTLV(s string) UnexpectedTLVError {
	return unexpectedTLV(s, 2)
}

func unexpectedTLV(s string, skip int) (e UnexpectedTLVError) {
	if p, f, l, ok := runtime.Caller(skip); ok {
		e.fileName = filepath.Base(f)
		e.line = l
		e.funcName = runtime.FuncForPC(p).Name()
	} else {
		e.fileName = "unknown"
		e.line = 0
		e.funcName = "unknown"
	}
	e.s = s
	return
}

func (e UnexpectedTLVError) Error() string {
	return fmt.Sprintf("unexpected TLV in %s(%s %d), %s",
		e.funcName, e.fileName, e.line, e.s)
}
