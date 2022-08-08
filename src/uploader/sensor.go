package uploader

import (
	"encoding/json"
	"strings"
	"time"

	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/buffer"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/connection"
)

// Sensor contains the logic to upload data to analysis cloud
type Sensor struct {
	buf          buffer.Data
	con          *connection.Connection
	quitChannels map[string]map[string]chan bool
}

// Init initialise the connection to the analysis cloud
func (u *Sensor) Init(buf buffer.Data, con *connection.Connection) {
	//var quitChanel map[string]map[string]chan bool
	//quitChanel = map[string]map[string]chan bool{}
	//u := Sensor{buf: buf, con: con, quitChannels: quitChanel}
	u.buf = buf
	u.con = con
	u.quitChannels = make(map[string]map[string]chan bool)
}

// GetBuffer returns a pointer to the buffer of the uploader
func (u *Sensor) GetBuffer() *buffer.Data {
	return &u.buf
}

// StartHandler starts a handler for a given machine sensor combination
func (u Sensor) StartHandler(machine, sensor string, interval time.Duration) {
	_, ok := u.quitChannels[machine]
	if !ok {
		u.quitChannels[machine] = make(map[string]chan bool)
	}

	if _, ok := u.quitChannels[machine][sensor]; ok {
		return
	}

	channel := make(chan bool, 1)
	u.quitChannels[machine][sensor] = channel
	go u.handler(interval, machine, sensor, channel)
}

// Stop stops an specific handler defined by machine and sensor
func (u *Sensor) Stop(machine, sensor string) {
	u.stop(machine, sensor)
	delete(u.quitChannels[machine], sensor)
	if len(u.quitChannels[machine]) == 0 {
		delete(u.quitChannels, machine)
	}
}

// ChangeInterval change the interval, which uploads a specific
func (u *Sensor) ChangeInterval(machine, sensor string, interval time.Duration) {
	klog.Infof("stop requested machine %s sensor %s", machine, sensor)
	channel := u.stop(machine, sensor)
	klog.Infof("stopped %s, %s sensor", machine, sensor)
	if channel == nil {
		u.StartHandler(machine, sensor, interval)
		return
	}

	go u.handler(interval, machine, sensor, channel)
}

func (u *Sensor) stop(machine, sensor string) chan bool {
	if _, ok := u.quitChannels[machine]; !ok {
		return nil
	}
	channel, ok := u.quitChannels[machine][sensor]
	if !ok {
		return nil
	}

	channel <- true
	return channel
}

func (u *Sensor) handler(interval time.Duration, machine, sensor string, quit chan bool) {
	klog.Infof("handler of machine %s sensor %s with interval %v has been started", machine, sensor, interval)
	var nextIteration bool = true
	go func() {
		<-quit
		nextIteration = false
	}()
	for {
		klog.Infof("next iteration of machine %s and upload %s sensor has begun", machine, sensor)
		if !nextIteration {
			return
		}
		time.Sleep(interval)
		klog.Infof("sleeping interval has been finished of machine %s and upload %s", machine, sensor)
		data := u.buf.GetValues(machine, sensor)

		klog.Infof("handling machine %s sensor %s with the length of data %d", machine, sensor, len(data))

		// do not upload empty data
		if len(data) == 0 {
			continue
		}

		encodedData, err := json.Marshal(data)
		if err != nil {
			klog.Errorf("cannot marshal data: %s", err)
		}

		klog.Infof("upload data to analysis cloud of machine %s and sensor %s", machine, sensor)
		req, err := u.con.Request("POST", "machine-data", nil, strings.NewReader(string(encodedData)))
		if err != nil {
			klog.Errorf("cannot upload data, this data is not stores by this program: %s", err)
		}
		klog.Infof("data upload has been finished of machine %s and sensor %s", machine, sensor)

		if req.StatusCode != 201 && req.StatusCode != 200 {
			klog.Errorf("cannot upload data, status code %d is been returned", req.StatusCode)
		}

	}
}
