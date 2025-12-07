package tcap

var (
	RxFailureNotify func(error, []byte)
	TraceMessage    func(Message, Direction, error)
)

// Tx or Rx.
type Direction bool

func (v Direction) String() string {
	if v {
		return "Tx"
	}
	return "Rx"
}

const (
	Tx Direction = true
	Rx Direction = false
)
