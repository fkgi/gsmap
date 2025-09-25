package ife

import (
	"encoding/json"

	"github.com/fkgi/gsmap"
)

/*
DeliveryFailureCause

	SM-EnumeratedDeliveryFailureCause ::= ENUMERATED {
		memoryCapacityExceeded     (0),
		equipmentProtocolError     (1),
		equipmentNotSM-Equipped    (2),
		unknownServiceCentre       (3),
		sc-Congestion              (4),
		invalidSME-Address         (5),
		subscriberNotSC-Subscriber (6) }
*/
type DeliveryFailureCause byte

const (
	_ DeliveryFailureCause = iota
	MemoryCapacityExceeded
	EquipmentProtocolError
	EquipmentNotSM_Equipped
	UnknownServiceCentre
	SC_Congestion
	InvalidSME_Address
	SubscriberNotSC_Subscriber
)

func (c DeliveryFailureCause) String() string {
	switch c {
	case MemoryCapacityExceeded:
		return "memoryCapacityExceeded"
	case EquipmentProtocolError:
		return "equipmentProtocolError"
	case EquipmentNotSM_Equipped:
		return "equipmentNotSM-Equipped"
	case UnknownServiceCentre:
		return "unknownServiceCentre"
	case SC_Congestion:
		return "sc-Congestion"
	case InvalidSME_Address:
		return "invalidSME-Address"
	case SubscriberNotSC_Subscriber:
		return "subscriberNotSC-Subscriber"
	}
	return ""
}

func (c *DeliveryFailureCause) UnmarshalJSON(b []byte) (e error) {
	var s string
	e = json.Unmarshal(b, &s)
	switch s {
	case "memoryCapacityExceeded":
		*c = MemoryCapacityExceeded
	case "equipmentProtocolError":
		*c = EquipmentProtocolError
	case "equipmentNotSM-Equipped":
		*c = EquipmentNotSM_Equipped
	case "unknownServiceCentre":
		*c = UnknownServiceCentre
	case "sc-Congestion":
		*c = SC_Congestion
	case "invalidSME-Address":
		*c = InvalidSME_Address
	case "subscriberNotSC-Subscriber":
		*c = SubscriberNotSC_Subscriber
	default:
		*c = 0
	}
	return
}

func (c DeliveryFailureCause) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *DeliveryFailureCause) unmarshal(b []byte) error {
	if len(b) != 1 {
		return gsmap.UnexpectedEnumValue(b)
	}
	switch b[0] {
	case 0:
		*c = MemoryCapacityExceeded
	case 1:
		*c = EquipmentProtocolError
	case 2:
		*c = EquipmentNotSM_Equipped
	case 3:
		*c = UnknownServiceCentre
	case 4:
		*c = SC_Congestion
	case 5:
		*c = InvalidSME_Address
	case 6:
		*c = SubscriberNotSC_Subscriber
	default:
		return gsmap.UnexpectedEnumValue(b)
	}
	return nil
}

func (c DeliveryFailureCause) marshal() []byte {
	switch c {
	case MemoryCapacityExceeded:
		return []byte{0x00}
	case EquipmentProtocolError:
		return []byte{0x01}
	case EquipmentNotSM_Equipped:
		return []byte{0x02}
	case UnknownServiceCentre:
		return []byte{0x03}
	case SC_Congestion:
		return []byte{0x04}
	case InvalidSME_Address:
		return []byte{0x05}
	case SubscriberNotSC_Subscriber:
		return []byte{0x06}
	}
	return nil
}
