package xua

var (
	AspUpNotify func(la, pa *SCTPAddr, e error)
	AsUpNotify  func(ctx uint32, e error)
	DunaNotify  func([]PointCode)
	DavaNotify  func([]PointCode)
	DaudNotify  func([]PointCode)
	SconNotify  func([]PointCode, uint32)
	DupuNotify  func([]PointCode, uint16)
	DrstNotify  func([]PointCode)

	TxFailureNotify func(error, []byte) = nil
	RxFailureNotify func(error, []byte) = nil
)
