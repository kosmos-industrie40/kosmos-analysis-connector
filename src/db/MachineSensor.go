package db

import (
	"database/sql"

	"k8s.io/klog"
)

// MachineSensor is the machine sensor definition
type MachineSensor struct {
	Machine string
	Sensor  string
}

// GetMachineSensorFromContract loads all machine sensors based on a contract id
// TODO function needs an unittest
func GetMachineSensorFromContract(db *sql.DB, contract string) ([]MachineSensor, error) {
	res, err := db.Query("SELECT machine, sensor FROM machine_sensor JOIN contract_machine_sensor ON machine_sensor = id WHERE contract = $1", contract)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	var machineSensor []MachineSensor
	for res.Next() {
		var machine, sensor string

		if err := res.Scan(&machine, &sensor); err != nil {
			return nil, err
		}
		machineSensor = append(machineSensor, MachineSensor{Machine: machine, Sensor: sensor})
	}

	return machineSensor, nil
}
