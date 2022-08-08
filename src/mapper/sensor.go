package mapper

import (
	"encoding/json"
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/buffer"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/connection"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/mqtt"
)

// SensorData is the logic to handle sensor update data
type SensorData struct {
	machine string
	sensor  string
	buffer  buffer.Data
	//exitChan chan bool TODO enable an unsubscribe mechanism
}

// Init initialize the SensorData handler
func (s SensorData) Init(mClient mqtt.Mqtt, buf buffer.Data, machine, sensor string) error {
	s.machine = machine
	s.sensor = sensor
	s.buffer = buf

	if err := mClient.Subscribe(fmt.Sprintf("kosmos/machine-data/%s/sensor/%s/update", s.machine, s.sensor), s.handler); err != nil {
		return err
	}

	return nil
}

func (s SensorData) handler(client MQTT.Client, m MQTT.Message) {
	klog.Infof("a sensor handle message received for machine %s sensor %s and topic:\n\t%s", s.machine, s.sensor, m.Topic())
	var (
		mData mqtt.SensorData
		cData connection.SensorData
	)

	if err := json.Unmarshal(m.Payload(), &mData); err != nil {
		klog.Errorf("cannot unmarshal sensor upload data: %s", err)
	}

	for _, column := range mData.Body.Columns {
		cData.Body.Columns = append(cData.Body.Columns, connection.Column{
			Name: column.Name,
			Type: column.Type,
		})
	}

	cData.Body.Data = mData.Body.Data
	cData.Body.Machine = s.machine
	cData.Body.Sensor = s.sensor
	cData.Body.Timestamp = mData.Body.Timestamp
	cData.Body.From = "analysis edge test"

	for _, v := range mData.Body.Meta {
		cData.Body.Meta = append(cData.Body.Meta, connection.Meta{
			Name:        v.Name,
			Description: v.Description,
			Type:        v.Type,
			Value:       v.Value,
		})
	}

	s.buffer.Insert(s.machine, s.sensor, cData)
	klog.Infof("added message to buffer")
}
