package gsmap

type AppContext uint64

func (a AppContext) Marshal() []byte {
	return []byte{
		byte(a >> 48), byte(a >> 40), byte(a >> 32),
		byte(a >> 24), byte(a >> 16), byte(a >> 8), byte(a)}
}

func (a *AppContext) Unmarshal(b []byte) {
	for _, v := range b {
		*a = (*a << 8) | AppContext(v)
	}
}

func (a AppContext) Application() byte {
	return 0xff & byte(a>>8)
}

func (a AppContext) Version() byte {
	return 0xff & byte(a)
}

/*
const (
	RoamingNumberEnquiry1                   AppContext = 0x0004000001000301
	RoamingNumberEnquiry2                   AppContext = 0x0004000001000302
	RoamingNumberEnquiry3                   AppContext = 0x0004000001000303
	IstAlerting3                            AppContext = 0x0004000001000403
	LocationInfoRetrieval1                  AppContext = 0x0004000001000501
	LocationInfoRetrieval2                  AppContext = 0x0004000001000502
	LocationInfoRetrieval3                  AppContext = 0x0004000001000503
	CallControlTransfer3                    AppContext = 0x0004000001000603
	CallControlTransfer4                    AppContext = 0x0004000001000604
	Reporting3                              AppContext = 0x0004000001000703
	CallCompletion3                         AppContext = 0x0004000001000803
	ServiceTermination3                     AppContext = 0x0004000001000903
	HandoverControl1                        AppContext = 0x0004000001000b01
	HandoverControl2                        AppContext = 0x0004000001000b02
	HandoverControl3                        AppContext = 0x0004000001000b03
	SIWFSAllocationContext3                 AppContext = 0x0004000001000c03
	EquipmentMngt1                          AppContext = 0x0004000001000d01
	EquipmentMngt2                          AppContext = 0x0004000001000d02
	EquipmentMngt3                          AppContext = 0x0004000001000d03
	InfoRetrieval1                          AppContext = 0x0004000001000e01
	InfoRetrieval2                          AppContext = 0x0004000001000e02
	InfoRetrieval3                          AppContext = 0x0004000001000e03
	InterVlrInfoRetrieval2                  AppContext = 0x0004000001000f02
	InterVlrInfoRetrieval3                  AppContext = 0x0004000001000f03
	Tracing1                                AppContext = 0x0004000001001101
	Tracing2                                AppContext = 0x0004000001001102
	Tracing3                                AppContext = 0x0004000001001103
	NetworkFunctionalSs1                    AppContext = 0x0004000001001201
	NetworkFunctionalSs2                    AppContext = 0x0004000001001202
	NetworkUnstructuredSs2                  AppContext = 0x0004000001001302
	SubscriberDataModificationNotification3 AppContext = 0x0004000001001603
	ImsiRetrieval2                          AppContext = 0x0004000001001a02
	SubscriberInfoEnquiry3                  AppContext = 0x0004000001001c03
	AnyTimeInfoEnquiry3                     AppContext = 0x0004000001001d03
	GroupCallControl3                       AppContext = 0x0004000001001f03
	GprsLocationUpdate3                     AppContext = 0x0004000001002003
	GprsLocationInfoRetrieval3              AppContext = 0x0004000001002103
	GprsLocationInfoRetrieval4              AppContext = 0x0004000001002104
	FailureReport3                          AppContext = 0x0004000001002203
	GprsNotify3                             AppContext = 0x0004000001002303
	SsInvocationNotification3               AppContext = 0x0004000001002403
	LocationSvcGateway3                     AppContext = 0x0004000001002503
	LocationSvcEnquiry3                     AppContext = 0x0004000001002603
	AuthenticationFailureReport3            AppContext = 0x0004000001002703
	ShortMsgMTRelayVGCS3                    AppContext = 0x0004000001002903
	MmEventReporting3                       AppContext = 0x0004000001002a03
	AnyTimeInfoHandling3                    AppContext = 0x0004000001002b03
	ResourceManagement3                     AppContext = 0x0004000001002c03
	GroupCallInfoRetrieval3                 AppContext = 0x0004000001002d03
	VcsgLocationUpdate3                     AppContext = 0x0004000001002e03
	VcsgLocationCancellation3               AppContext = 0x0004000001002f03
)
*/
