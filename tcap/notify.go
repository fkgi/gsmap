package tcap

var (
	RxFailureNotify func(error, []byte)
	TraceMessage    func(Message, Direction, error)
)
