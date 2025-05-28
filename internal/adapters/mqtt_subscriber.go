package adapters

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/vibin/frigate_alerter/internal/domain"
)

// MQTTSubscriber implements the EventSubscriber interface
type MQTTSubscriber struct {
	client     mqtt.Client
	topic      string
	connected  bool
}

// NewMQTTSubscriber creates a new MQTT subscriber
func NewMQTTSubscriber(brokerURL string) (*MQTTSubscriber, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(fmt.Sprintf("frigate-alerter-%d", time.Now().Unix())).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			slog.Error("MQTT connection lost", "error", err)
		}).
		SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
			slog.Info("MQTT attempting to reconnect")
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &MQTTSubscriber{
		client:    client,
		topic:     "frigate/events",
		connected: true,
	}, nil
}

// Subscribe starts listening for events on the MQTT topic
func (m *MQTTSubscriber) Subscribe(handler func(event *domain.FrigateEvent)) error {
	if !m.connected {
		return fmt.Errorf("MQTT client not connected")
	}

	token := m.client.Subscribe(m.topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		var event domain.FrigateEvent
		if err := json.Unmarshal(msg.Payload(), &event); err != nil {
			slog.Error("Error unmarshalling MQTT message", "error", err, "payload", string(msg.Payload()))
			return
		}
		
		handler(&event)
	})

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	slog.Info("Subscribed to MQTT topic", "topic", m.topic)
	return nil
}

// Close disconnects from the MQTT broker
func (m *MQTTSubscriber) Close() error {
	if m.connected {
		m.client.Disconnect(250) // wait 250ms for the disconnect to complete
		m.connected = false
	}
	return nil
}
