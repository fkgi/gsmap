package ipsp

var (
	ErrorNotify func(id byte, c ErrCode)
	StateNotify func(id byte, s Status)
	SctpNotify  func(id byte, s string)

	DunaNotify func([]PointCode)
	DavaNotify func([]PointCode)
	DaudNotify func([]PointCode)
	SconNotify func([]PointCode, uint32)
	DupuNotify func([]PointCode, uint16)
	DrstNotify func([]PointCode)

	TxFailureNotify func(error, []byte) = nil
	RxFailureNotify func(error, []byte) = nil
)
