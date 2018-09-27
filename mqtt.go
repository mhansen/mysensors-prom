package mysensors

import (
	"flag"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	broker       = flag.String("broker", "tcp://192.168.0.100:1883", "MQTT broker address")
	topicPrefix  = flag.String("topic_prefix", "mysensors", "Prefix for MQTT topic")
	clientPrefix = flag.String("client_prefix", "mysensors-", "Prefix for MQTT client name")
)

var clientID = 0

type MQTTClient struct {
	client  mqtt.Client
	options *mqtt.ClientOptions
	msgChan chan *Message
}

func (m *MQTTClient) Start(ch chan *Message) error {
	m.options = mqtt.NewClientOptions().AddBroker(*broker)
	m.options.SetClientID(*clientPrefix)
	m.options.SetConnectionLostHandler(m.connLostHandler)
	m.options.SetAutoReconnect(false)

	m.msgChan = ch

	err := m.startClient()
	go m.messageListener()
	return err
}

func (m *MQTTClient) startClient() error {
	m.client = mqtt.NewClient(m.options)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *MQTTClient) messageListener() {
	for msg := range m.msgChan {
		topic := fmt.Sprintf("%s/%d/%d/%s", *topicPrefix, msg.NodeID, msg.ChildSensorID, msg.SubType)
		if token := m.client.Publish(topic, 0, true, msg.Payload); token.Wait() && token.Error() != nil {
			log.Printf("MQTT publish error: %v\n", token.Error())
		}
	}
}

func (m *MQTTClient) connLostHandler(client mqtt.Client, reason error) {
	log.Printf("MQTT connection lost: %v", reason)
	clientID++
	m.options.SetClientID(fmt.Sprintf("%s%d", *clientPrefix, clientID))
	// TODO: Handle persistent failure.
	m.startClient()
}
