// Package mysensors is a Go API for talking to a MySensors Gateway.
//
// message.go contains definitions for individual messages.
package mysensors

import (
	"fmt"
	"strconv"
	"strings"
)

type AckType uint8

const (
	NoAck AckType = iota
	Ack
)

var ackType = [...]string{
	"noack",
	"ack",
}

func (t AckType) String() string { return ackType[t] }

// MsgType is a MySensors message type.
type MsgType uint8

const (
	MsgPresentation MsgType = iota
	MsgSet
	MsgReq
	MsgInternal
	MsgStream
)

var msgType = [...]string{
	"presentation",
	"set",
	"req",
	"internal",
	"stream",
}

func (t MsgType) String() string { return msgType[t] }

// SubType is an interface for message SubTypes.
type SubType interface {
	// String is the string representation of the type (e.g S_TEMP).
	String() string
	// Value is the numeric value of the SubType.
	Value() uint8
}

// SubTypePresentation are SubTypes for presentation messages.
type SubTypePresentation uint8

const (
	S_DOOR SubTypePresentation = iota
	S_MOTION
	S_SMOKE
	S_LIGHT
	S_DIMMER
	S_COVER
	S_TEMP
	S_HUM
	S_BARO
	S_WIND
	S_RAIN
	S_UV
	S_WEIGHT
	S_POWER
	S_HEATER
	S_DISTANCE
	S_LIGHT_LEVEL
	S_ARDUINO_NODE
	S_ARDUINO_REPEATER_NODE
	S_LOCK
	S_IR
	S_WATER
	S_AIR_QUALITY
	S_CUSTOM
	S_DUST
	S_SCENE_CONTROLLER
	S_RGB_LIGHT
	S_RGBW_LIGHT
	S_COLOR_SENSOR
	S_HVAC
	S_MULTIMETER
	S_SPRINKLER
	S_WATER_LEAK
	S_SOUND
	S_VIBRATION
	S_MOISTURE
	S_BINARY SubTypePresentation = 3
)

var subTypePresentation = [...]string{
	"S_DOOR",
	"S_MOTION",
	"S_SMOKE",
	"S_LIGHT",
	"S_DIMMER",
	"S_COVER",
	"S_TEMP",
	"S_HUM",
	"S_BARO",
	"S_WIND",
	"S_RAIN",
	"S_UV",
	"S_WEIGHT",
	"S_POWER",
	"S_HEATER",
	"S_DISTANCE",
	"S_LIGHT_LEVEL",
	"S_ARDUINO_NODE",
	"S_ARDUINO_REPEATER_NODE",
	"S_LOCK",
	"S_IR",
	"S_WATER",
	"S_AIR_QUALITY",
	"S_CUSTOM",
	"S_DUST",
	"S_SCENE_CONTROLLER",
	"S_RGB_LIGHT",
	"S_RGBW_LIGHT",
	"S_COLOR_SENSOR",
	"S_HVAC",
	"S_MULTIMETER",
	"S_SPRINKLER",
	"S_WATER_LEAK",
	"S_SOUND",
	"S_VIBRATION",
	"S_MOISTURE",
}

// String formats an optionally-present SubTypePresentation for status messages.
func (t SubTypePresentation) String() string {
	return subTypePresentation[t]
}

func (t SubTypePresentation) Value() uint8 { return uint8(t) }

func (t *SubTypePresentation) StatusString() string {
	if t == nil {
		return "UNKNOWN"
	}
	return t.String()
}

// SubTypeSetReq are SubTypes for set and request messages.
type SubTypeSetReq uint8

const (
	V_TEMP SubTypeSetReq = iota
	V_HUM
	V_STATUS
	//V_LIGHT
	V_PERCENTAGE
	//V_DIMMER
	V_PRESSURE
	V_FORECAST
	V_RAIN
	V_RAINRATE
	V_WIND
	V_GUST
	V_DIRECTION
	V_UV
	V_WEIGHT
	V_DISTANCE
	V_IMPEDANCE
	V_ARMED
	V_TRIPPED
	V_WATT
	V_KWH
	V_SCENE_ON
	V_SCENE_OFF
	V_HVAC_FLOW_STATE
	V_HVAC_SPEED
	V_LIGHT_LEVEL
	V_VAR1
	V_VAR2
	V_VAR3
	V_VAR4
	V_VAR5
	V_UP
	V_DOWN
	V_STOP
	V_IR_SEND
	V_IR_RECEIVE
	V_FLOW
	V_VOLUME
	V_LOCK_STATUS
	V_LEVEL
	V_VOLTAGE
	V_CURRENT
	V_RGB
	V_RGBW
	V_ID
	V_UNIT_PREFIX
	V_HVAC_SETPOINT_COOL
	V_HVAC_SETPOINT_HEAT
	V_HVAC_FLOW_MODE
)

var subTypeSetReq = [...]string{
	"V_TEMP",
	"V_HUM",
	"V_STATUS",
	//"V_LIGHT",
	"V_PERCENTAGE",
	//"V_DIMMER",
	"V_PRESSURE",
	"V_FORECAST",
	"V_RAIN",
	"V_RAINRATE",
	"V_WIND",
	"V_GUST",
	"V_DIRECTION",
	"V_UV",
	"V_WEIGHT",
	"V_DISTANCE",
	"V_IMPEDANCE",
	"V_ARMED",
	"V_TRIPPED",
	"V_WATT",
	"V_KWH",
	"V_SCENE_ON",
	"V_SCENE_OFF",
	"V_HVAC_FLOW_STATE",
	"V_HVAC_SPEED",
	"V_LIGHT_LEVEL",
	"V_VAR1",
	"V_VAR2",
	"V_VAR3",
	"V_VAR4",
	"V_VAR5",
	"V_UP",
	"V_DOWN",
	"V_STOP",
	"V_IR_SEND",
	"V_IR_RECEIVE",
	"V_FLOW",
	"V_VOLUME",
	"V_LOCK_STATUS",
	"V_LEVEL",
	"V_VOLTAGE",
	"V_CURRENT",
	"V_RGB",
	"V_RGBW",
	"V_ID",
	"V_UNIT_PREFIX",
	"V_HVAC_SETPOINT_COOL",
	"V_HVAC_SETPOINT_HEAT",
	"V_HVAC_FLOW_MODE",
}

func (t SubTypeSetReq) String() string { return subTypeSetReq[t] }

func (t SubTypeSetReq) Value() uint8 { return uint8(t) }

// SubTypeInternal are SubTypes for internal messages.

type SubTypeInternal uint8

const (
	I_BATTERY_LEVEL SubTypeInternal = iota
	I_TIME
	I_VERSION
	I_ID_REQUEST
	I_ID_RESPONSE
	I_INCLUSION_MODE
	I_CONFIG
	I_FIND_PARENT
	I_FIND_PARENT_RESPONSE
	I_LOG_MESSAGE
	I_CHILDREN
	I_SKETCH_NAME
	I_SKETCH_VERSION
	I_REBOOT
	I_GATEWAY_READY
	I_REQUEST_SIGNING
	I_GET_NONCE
	I_GET_NONCE_RESPONSE
)

var subTypeInternal = [...]string{
	"I_BATTERY_LEVEL",
	"I_TIME",
	"I_VERSION",
	"I_ID_REQUEST",
	"I_ID_RESPONSE",
	"I_INCLUSION_MODE",
	"I_CONFIG",
	"I_FIND_PARENT",
	"I_FIND_PARENT_RESPONSE",
	"I_LOG_MESSAGE",
	"I_CHILDREN",
	"I_SKETCH_NAME",
	"I_SKETCH_VERSION",
	"I_REBOOT",
	"I_GATEWAY_READY",
	"I_REQUEST_SIGNING",
	"I_GET_NONCE",
	"I_GET_NONCE_RESPONSE",
}

func (t SubTypeInternal) String() string { return subTypeInternal[t] }

func (t SubTypeInternal) Value() uint8 { return uint8(t) }

// Message is a complete MySensors message.

type Message struct {
	// NodeID is the node id.
	NodeID uint8
	// NodeID is the child sensor id.
	ChildSensorID uint8
	// Type is the message type.
	Type MsgType
	// Ack indicates the ACK value.
	Ack AckType
	// SubType is the subtype of the message.
	SubType SubType
	// Payload it the payload of the message.
	Payload []byte
}

// String returns a string representation of the message.
func (m *Message) String() string {
	return fmt.Sprintf("%d:%d:%s:%s:%s:%s",
		m.NodeID, m.ChildSensorID, m.Type, m.Ack, m.SubType, string(m.Payload))
}

// Copy returns a copy of the message.
func (m *Message) Copy() *Message {
	b := m.Marshal()
	n := &Message{}
	if err := n.Unmarshal(b); err != nil {
		panic(err)
	}
	return n
}

// Marshal marshals the message into a byte slice.
func (m *Message) Marshal() []byte {
	return []byte(fmt.Sprintf("%d;%d;%d;%d;%d;%s\n", m.NodeID, m.ChildSensorID, m.Type, m.Ack, m.SubType, m.Payload))
}

// Unmarshal reads the given wire bytes into the Message.
func (m *Message) Unmarshal(b []byte) error {
	s := strings.TrimSuffix(string(b), "\x0a")
	parts := strings.SplitN(s, ";", 6)
	if len(parts) != 6 {
		return fmt.Errorf("invalid format, only %d parts", len(parts))
	}
	if nodeID, err := strconv.Atoi(parts[0]); err != nil {
		return err
	} else {
		m.NodeID = uint8(nodeID)
	}

	if childSensorID, err := strconv.Atoi(parts[1]); err != nil {
		return err
	} else {
		m.ChildSensorID = uint8(childSensorID)
	}

	if mType, err := strconv.Atoi(parts[2]); err != nil {
		return err
	} else {
		m.Type = MsgType(mType)
	}

	if ack, err := strconv.Atoi(parts[3]); err != nil {
		return err
	} else {
		m.Ack = AckType(ack)
	}

	if subType, err := strconv.Atoi(parts[4]); err != nil {
		return err
	} else {
		switch m.Type {
		case MsgPresentation:
			m.SubType = SubTypePresentation(subType)
		case MsgSet, MsgReq:
			m.SubType = SubTypeSetReq(subType)
		case MsgInternal:
			m.SubType = SubTypeInternal(subType)
		}
	}

	m.Payload = []byte(parts[5])
	return nil
}
