package mqtt

import (
	"crypto/tls"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"
)

// Mqtt contains the mqtt functionality
type Mqtt struct {
	client mqtt.Client
}

// NewMqtt returns a mqtt.Mqtt object from a paho MQTT.Client
func NewMqtt(client mqtt.Client) Mqtt {
	var mq Mqtt
	mq.client = client
	return mq
}

// Connect create a connection to a mqtt broker
func (m *Mqtt) Connect(url, user, password, clientID string, port int, tlsConfig *tls.Config) error {
	options := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("%s:%d", url, port))
	options.SetAutoReconnect(true)
	options.SetTLSConfig(tlsConfig)
	options.SetClientID(clientID)
	options.SetCleanSession(false)
	// options.SetKeepAlive(10)  // Results in error
	if user != "" {
		options.SetUsername(user)
		if password != "" {
			options.SetPassword(password)
		}
	}

	m.client = mqtt.NewClient(options)
	token := m.client.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Subscribe subscribe to a specific mqtt topic
func (m *Mqtt) Subscribe(topic string, callbackFunc mqtt.MessageHandler) error {
	klog.Infof("subscribe to topic: %s", topic)
	token := m.client.Subscribe(topic, 2, callbackFunc)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Unsubscribe unsubscribe of a given mqtt topic
func (m *Mqtt) Unsubscribe(topic ...string) error {
	token := m.client.Unsubscribe(topic...)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// Send a mqtt message to a specifc topic
func (m *Mqtt) Send(topic string, msg []byte) error {
	token := m.client.Publish(topic, 2, false, msg)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

// Disconnect closes the mqtt connection
func (m *Mqtt) Disconnect() {
	m.client.Disconnect(10)
}
