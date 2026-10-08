package tcap

var (
	TraceTxMessage func(Message, error)
	TraceRxMessage func(Message, error)
)
