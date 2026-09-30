package tcap

var (
	RxFailureNotify func(error, []byte)
	TraceTxMessage  func(Message, error)
	TraceRxMessage  func(Message, error)
)
