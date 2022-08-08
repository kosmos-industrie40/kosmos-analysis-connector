package buffer

import (
	"sync"

	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/connection"
)

// Data contains the machine data in a sync map
type Data interface {
	Insert(machine, sensor string, update connection.SensorData)
	GetValues(machine, sensor string) []connection.SensorData
}

type data struct {
	syncMap map[string]map[string][]connection.SensorData
	mutex   sync.Mutex
}

// NewLocalBuffer initialise the data type
func NewLocalBuffer() Data {
	var d data
	dd := &d
	dd.init()
	return dd
}

func (u *data) init() {
	u.syncMap = make(map[string]map[string][]connection.SensorData)
	u.mutex = sync.Mutex{}
}

// Insert insert new data to a machine sensor combination
func (u *data) Insert(machine, sensor string, update connection.SensorData) {
	u.mutex.Lock()
	_, ok := u.syncMap[machine]
	if !ok {
		mm := make(map[string][]connection.SensorData)
		u.syncMap[machine] = mm
	}
	u.syncMap[machine][sensor] = append(u.syncMap[machine][sensor], update)
	u.mutex.Unlock()
}

// GetValues retuns all sensor data from machine, sensor string
func (u *data) GetValues(machine, sensor string) []connection.SensorData {
	var data []connection.SensorData
	u.mutex.Lock()
	data = u.syncMap[machine][sensor]
	delete(u.syncMap[machine], sensor)
	if len(u.syncMap[machine]) == 0 {
		delete(u.syncMap, machine)
	}
	u.mutex.Unlock()
	return data
}
