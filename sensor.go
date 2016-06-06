// This file contains per-sensor routines.
package mysensors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// FirstNodeID is the first ID to assign to nodes.
	FirstNodeID = 1
	// GatewayID is the Gateway's ID.
	GatewayID = 0
	// NoChild is the placeholder used for non-sensor node messages.
	NoChild = 255
)

// GaugeMap maps MySensor variables to prometheus variable names.
var GaugeMap = map[SubTypeSetReq]string{
	V_TEMP: "temperature",
	V_HUM: "humidity",
	V_PRESSURE: "pressure",
}

// Gauges contains a mapping from MySensor variables to prometheus gauge objects.
type Gauges struct {
	Gauge	map[SubTypeSetReq]*prometheus.GaugeVec
}

// Set sets the corresponding gauge to the given value.
func (g *Gauges) Set(t SubTypeSetReq, l string, v float64) {
	gs, ok := GaugeMap[t]
	if !ok {
		return
	}
	ga, ok := g.Gauge[t]
	if !ok {
		ga = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: gs,
				Help: fmt.Sprintf("MYSENSORS %s", t),
				ConstLabels: prometheus.Labels{"instance": "192.168.0.10:9001"},
			},
			[]string{"location"},
		)
		prometheus.MustRegister(ga)
		if len(g.Gauge) == 0 {
			g.Gauge = make(map[SubTypeSetReq]*prometheus.GaugeVec)
		}
		g.Gauge[t] = ga
	}
	ga.WithLabelValues(l).Set(v)
}

// Network is a container for all sensor nodes.
type Network struct {
	Nodes map[string]*Node
	gauges *Gauges
}

// NewNetwork initialises a new Network.
func NewNetwork() *Network {
	n := &Network{}
	n.Nodes = make(map[string]*Node, 0)
	n.gauges = &Gauges{}
	return n
}

// HandleMessage handles a MySensors message from the gateway.
func (n *Network) HandleMessage(m *Message) error {
	if m.NodeID == GatewayID {
		return n.handleMessage(m)
	}
	nID := fmt.Sprintf("%d", m.NodeID)
	nd, ok := n.Nodes[nID]
	if !ok {
		nd = NewNode(n)
		n.Nodes[nID] = nd
	}
	return nd.HandleMessage(m)
}

// handleMessage handles messages for/from the gateway.
func (n *Network) handleMessage(m *Message) error {
	log.Printf("GW MSG: %s\n", m)
	return nil
}

// StatusString prints a formatted representation of the network.
func (n *Network) StatusString() string {
	fmt.Printf(">>> status\n")
	for _, node := range n.Nodes {
		fmt.Printf("Node %d:\n", node.ID)
		fmt.Printf("  Location: %s\n", node.Location)
		fmt.Printf("  Battery: %d%%\n", node.Battery)
		fmt.Printf("  Sketch: %s [%s]\n", node.SketchName, node.SketchVersion)
		for _, s := range node.Sensors {
			fmt.Printf("    Sensor %d [%s]\n", s.ID, s.Presentation)
			for t, v := range s.Variables {
				fmt.Printf("      %s: %s\n", t, v.String())
			}
		}
	}
	fmt.Printf("<<< status\n")
	return ""
}
// Load reads State from a file.
func (n *Network) LoadJson(f string) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, n); err != nil {
		return err
	}
	// Re-add parent struct params which arent there after
	// JSON import.
	for _, node := range n.Nodes {
		node.network = n
		for  _, s := range node.Sensors {
			s.node = node
		}
	}
	return nil
}

// SaveJson saves the network to a file in Json format.
func (n *Network) SaveJson(f string) error {
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	json.Indent(&out, data, "", "  ")
	if err = ioutil.WriteFile(f, out.Bytes(), os.ModePerm); err != nil {
		return err
	}
	return nil
}

// NextNodeID allocates and returns a node ID.
func (n *Network) NextNodeID() uint8 {
	nextID := uint8(FirstNodeID)
	for _, node := range n.Nodes {
		if node.ID >= nextID {
			nextID = node.ID + 1
		}
	}
	return nextID
}

// Node is a node that may contain multiple sensors.
type Node struct {
	// ID is the node ID.
	ID uint8
	// Battery is the battery level percent.
	Battery int64
	// Location per the configuration.
	Location string
	// Version as reported.
	Version string
	// Sketch name.
	SketchName string
	// SketchVersion.
	SketchVersion string
	// Sensors are all sensors attached to the node.
	Sensors map[string]*Sensor
	// network is the parent network.
	network *Network
}

func NewNode(ne *Network) *Node {
	n := &Node{network: ne}
	n.Sensors = make(map[string]*Sensor)
	return n
}

func (n *Node) HandleMessage(m *Message) error {
	sID := fmt.Sprintf("%d", m.ChildSensorID)
	n.ID = m.NodeID
	if m.ChildSensorID == NoChild {
		return n.handleMessage(m)
	}
	cs, ok := n.Sensors[sID]
	if !ok {
		cs = NewSensor(n)
		n.Sensors[sID] = cs
	}
	return cs.HandleMessage(m)
}

func (n *Node) handleMessage(m *Message) error {
	if m.Type != MsgInternal {
		return fmt.Errorf("Unknown message to child id %d", NoChild)
	}
	subType := m.SubType.(SubTypeInternal)
	switch subType {
		case I_BATTERY_LEVEL:
			n.Battery, _ = strconv.ParseInt(string(m.Payload), 10, 32)
		case I_VERSION:
			n.Version = string(m.Payload)
		case I_SKETCH_NAME:
			n.SketchName = string(m.Payload)
		case I_SKETCH_VERSION:
			n.SketchVersion = string(m.Payload)
		default:
			log.Printf("UNKN: %s\n", m.String())
	}
	return nil
}

// Sensor is a child sensor.
type Sensor struct {
	// ID is the sensor ID.
	ID uint8
	// Presentation is the sensor subtype presented.
	Presentation SubTypePresentation
	// Variables are the variables presented by this child sensor.
	Variables map[string] Variable `json:"-"`
	// Node is the parent node.
	node *Node
}

func NewSensor(n *Node) *Sensor {
	s := &Sensor{node: n}
	s.Variables = make(map[string]Variable, 0)
	return s
}

func (s *Sensor) HandleMessage(m *Message) error {
	s.ID = m.ChildSensorID
	switch m.Type {
	case MsgPresentation:
		s.Presentation = m.SubType.(SubTypePresentation)
		log.Printf("PRES: %s\n", m)
	case MsgSet:
		subType := m.SubType.(SubTypeSetReq)
		if len(s.Variables) == 0 {
			s.Variables = make(map[string]Variable, 0)
		}
		if _, ok := s.Variables[subType.String()]; !ok {
			switch subType {
			case V_TEMP, V_HUM, V_PRESSURE:
				s.Variables[subType.String()] = &FloatVariable{}
			default:
				s.Variables[subType.String()] = &StringVariable{}
			}
		}
		s.Variables[subType.String()].Set(string(m.Payload))
		switch v := s.Variables[subType.String()].(type) {
		case *FloatVariable:
			s.node.network.gauges.Set(subType, s.node.Location, v.Value)
		}
		log.Printf("SET: %s\n", m)
	}
	return nil
}

type Variable interface {
	Name() string
	Set(string) error
	String() string
}


type StringVariable struct {
	name string
	Value string
}

func (sv *StringVariable) Name() string {
	return sv.name
}

func (sv *StringVariable) Set(v string) error {
	sv.Value = v
	return nil
}

func (sv *StringVariable) String() string {
	return sv.Value
}

type IntegerVariable struct {
	name string
	Value int64
}

func (iv *IntegerVariable) Name() string {
	return iv.name
}

func (iv *IntegerVariable) Set(v string) error {
	val, err := strconv.ParseInt(v, 10, 32)
	iv.Value = val
	return err
}

func (iv *IntegerVariable) String() string {
	return strconv.FormatInt(iv.Value, 10)
}

type FloatVariable struct {
	name string
	Value float64
	sensor *Sensor
}

func (fv *FloatVariable) Name() string {
	return fv.name
}

func (fv *FloatVariable) Set(v string) error {
	val, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return err
	}
	fv.Value = val
	return nil
}

func (fv *FloatVariable) String() string {
	return strconv.FormatFloat(fv.Value, 'f', 2, 64)
}
