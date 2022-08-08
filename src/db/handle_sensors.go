package db

import (
	"database/sql"

	"k8s.io/klog"
)

// HandleSensor contains the machine, sensor and the minimal duration
type HandleSensor struct {
	Machine  string
	Sensor   string
	Duration string
}

// HandleSensors returns all sensors with the minimal upload duration
func HandleSensors(db *sql.DB, version string) ([]HandleSensor, error) {
	res, err := db.Query("SELECT machine, sensor, min(duration) FROM machine_sensor AS ms JOIN contract_machine_sensor AS cms ON ms.id = cms.machine_sensor JOIN contract AS c ON cms.contract = c.contract WHERE version = $1 GROUP BY (machine, sensor)", version)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			klog.Errorf("cannot close result type error is: %s\n", err)
		}
	}()

	var retHandleSensors []HandleSensor
	for res.Next() {
		var machine, sensor, duration string
		if err := res.Scan(&machine, &sensor, &duration); err != nil {
			return nil, err
		}

		retHandleSensors = append(retHandleSensors, HandleSensor{Machine: machine, Sensor: sensor, Duration: duration})
	}

	return retHandleSensors, nil
}
