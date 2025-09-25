package gsmap

import (
	"bytes"
)

var (
	ArgMap = map[byte]Invoke{}
	ResMap = map[byte]ReturnResultLast{}
	ErrMap = map[byte]ReturnError{}

	NameMap = map[string]Component{}
)

var LogPrefix = " | | "

/*
func componentString(c Component) string {
	b, e := json.Marshal(c)
	if e != nil {
		return "invalid component: " + e.Error()
	}
	m := map[string]any{}
	json.Unmarshal(b, &m)
	return c.Name() + componentStringSub(LogPrefix, m)
}

func componentStringSub(p string, m map[string]any) string {
	buf := new(strings.Builder)
	for k, v := range m {
		if g, ok := v.(map[string]any); ok {
			v = componentStringSub(p+"| ", g)
		}
		fmt.Fprintf(buf, "\n%s%s: %v", p, k, v)
	}
	return buf.String()
}
*/

/*
Component portion interface.
*/
type Component interface {
	GetInvokeID() int8
	Code() byte
	Name() string
	NewFromJSON([]byte, int8) (Component, error)
	MarshalParam() []byte
}

/*
Invoke component portion interface.
*/
type Invoke interface {
	Component
	GetLinkedID() *int8
	Unmarshal(int8, *int8, *bytes.Buffer) (Invoke, error)
	DefaultContext() AppContext
}

/*
ReturnResultLast component portion interface.
*/
type ReturnResultLast interface {
	Component
	Unmarshal(int8, *bytes.Buffer) (ReturnResultLast, error)
}

/*
ReturnResult component portion interface.
*/
type ReturnResult interface {
	Component
	Unmarshal(int8, *bytes.Buffer) (ReturnResult, error)
}

/*
ReturnError component portion interface.
*/
type ReturnError interface {
	Component
	Unmarshal(int8, *bytes.Buffer) (ReturnError, error)
}
