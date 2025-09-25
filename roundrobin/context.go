package main

import (
	"github.com/fkgi/gsmap"
	"github.com/fkgi/gsmap/ifc"
	"github.com/fkgi/gsmap/ifd"
	"github.com/fkgi/gsmap/ife"
)

func getContext(n, v string) gsmap.AppContext {
	switch n {
	case "NetworkLocUp":
		switch v {
		case "v1":
			return ifd.NetworkLocUp1
		case "v2":
			return ifd.NetworkLocUp2
		case "v3":
			return ifd.NetworkLocUp3
		}
	case "LocationCancellation":
		switch v {
		case "v1":
			return ifd.LocationCancellation1
		case "v2":
			return ifd.LocationCancellation2
		case "v3":
			return ifd.LocationCancellation3
		}
	case "SubscriberDataMngt":
		switch v {
		case "v1":
			return ifd.SubscriberDataMngt1
		case "v2":
			return ifd.SubscriberDataMngt2
		case "v3":
			return ifd.SubscriberDataMngt3
		}
	case "MwdMngt":
		switch v {
		case "v1":
			return ifd.MwdMngt1
		case "v2":
			return ifd.MwdMngt2
		case "v3":
			return ifd.MwdMngt3
		}
	case "ShortMsgGateway":
		switch v {
		case "v1":
			return ifc.ShortMsgGateway1
		case "v2":
			return ifc.ShortMsgGateway2
		case "v3":
			return ifc.ShortMsgGateway3
		}
	case "ShortMsgAlert":
		switch v {
		case "v1":
			return ifc.ShortMsgAlert1
		case "v2":
			return ifc.ShortMsgAlert2
		}
	case "ShortMsgRelay":
		switch v {
		case "v1":
			return ife.ShortMsgRelay1
		}
	case "ShortMsgMORelay":
		switch v {
		case "v2":
			return ife.ShortMsgMORelay2
		case "v3":
			return ife.ShortMsgMORelay3
		}
	case "ShortMsgMTRelay":
		switch v {
		case "v2":
			return ife.ShortMsgMTRelay2
		case "v3":
			return ife.ShortMsgMTRelay3
		}
	}
	return 0
}

func getContextName(c gsmap.AppContext) (string, string) {
	switch c {
	case ifd.NetworkLocUp1:
		return "NetworkLocUp", "v1"
	case ifd.NetworkLocUp2:
		return "NetworkLocUp", "v2"
	case ifd.NetworkLocUp3:
		return "NetworkLocUp", "v3"
	case ifd.LocationCancellation1:
		return "LocationCancellation", "v1"
	case ifd.LocationCancellation2:
		return "LocationCancellation", "v2"
	case ifd.LocationCancellation3:
		return "LocationCancellation", "v3"
	case ifd.SubscriberDataMngt1:
		return "SubscriberDataMngt", "v1"
	case ifd.SubscriberDataMngt2:
		return "SubscriberDataMngt", "v2"
	case ifd.SubscriberDataMngt3:
		return "SubscriberDataMngt", "v3"
	case ifd.MwdMngt1:
		return "MwdMngt", "v1"
	case ifd.MwdMngt2:
		return "MwdMngt", "v2"
	case ifd.MwdMngt3:
		return "MwdMngt", "v3"
	case ifc.ShortMsgGateway1:
		return "ShortMsgGateway", "v1"
	case ifc.ShortMsgGateway2:
		return "ShortMsgGateway", "v2"
	case ifc.ShortMsgGateway3:
		return "ShortMsgGateway", "v3"
	case ifc.ShortMsgAlert1:
		return "ShortMsgAlert", "v1"
	case ifc.ShortMsgAlert2:
		return "ShortMsgAlert", "v2"
	case ife.ShortMsgRelay1:
		return "ShortMsgRelay", "v1"
	case ife.ShortMsgMORelay2:
		return "ShortMsgMORelay", "v2"
	case ife.ShortMsgMORelay3:
		return "ShortMsgMORelay", "v3"
	case ife.ShortMsgMTRelay2:
		return "ShortMsgMTRelay", "v2"
	case ife.ShortMsgMTRelay3:
		return "ShortMsgMTRelay", "v3"
	}
	return "", ""
}
