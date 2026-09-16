package xua

var (
	TraceEvent func(old, new, event string, e error)

	DunaNotify func([]PointCode)
	DavaNotify func([]PointCode)
	DaudNotify func([]PointCode)
	SconNotify func([]PointCode, uint32)
	DupuNotify func([]PointCode, uint16)
	DrstNotify func([]PointCode)

	AsStateNotify func(string)

	TxFailureNotify func(error, []byte) = nil
	RxFailureNotify func(error, []byte) = nil
)
