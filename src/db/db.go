package db

import (
	"database/sql"
	"fmt"
	"time"

	"k8s.io/klog"
)

// ContractExists tests if a contract with a specific id exists or not
func ContractExists(db *sql.DB, contract string) (bool, error) {
	klog.Infof("test if the contract %s already exists", contract)

	existsQuery, err := db.Query("SELECT contract FROM contract WHERE contract = $1", contract)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := existsQuery.Close(); err != nil {
			klog.Errorf("cannot close databse query: %s\n", err)
		}
	}()

	if !existsQuery.Next() {
		return false, nil
	}
	return true, nil
}

// Insert data from a new contract
func Insert(db *sql.DB, machine, sensor, duration, version, contract string) error {

	klog.Infof("insert machine %s sensor %s duration %s version %s and contract %s into db", machine, sensor, duration, version, contract)

	machineSensorIDQuery, err := db.Query("SELECT id FROM machine_sensor WHERE machine = $1 AND sensor = $2", machine, sensor)
	if err != nil {
		return err
	}

	defer func() {
		if err := machineSensorIDQuery.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	var machineSensorID int64
	if machineSensorIDQuery.Next() {
		if err := machineSensorIDQuery.Scan(&machineSensorID); err != nil {
			return err
		}
	} else {
		res, err := db.Query("INSERT INTO machine_sensor (machine, sensor) VALUES ($1, $2) RETURNING id", machine, sensor)
		if err != nil {
			return err
		}

		defer func() {
			if err := res.Close(); err != nil {
				klog.Errorf("cannot close query object %s\n", err)
			}
		}()

		if !res.Next() {
			return fmt.Errorf("the id will not returned")
		}

		if err := res.Scan(&machineSensorID); err != nil {
			return err
		}

		if err != nil {
			return err
		}
	}

	req, err := db.Query("SELECT * from contract WHERE contract = $1", contract)
	if err != nil {
		return err
	}

	defer func() {
		if err := req.Close(); err != nil {
			klog.Errorf("canot close query object: %s", err)
		}
	}()

	if !req.Next() {
		if _, err := db.Exec("INSERT INTO contract (contract, duration, version) VALUES ($1, $2, $3)", contract, duration, version); err != nil {
			return fmt.Errorf("in contract insertion error: %s is occured", err)
		}
	}

	_, err = db.Exec("INSERT INTO contract_machine_sensor (contract, machine_sensor) VALUES ($1, $2)", contract, machineSensorID)

	return err
}

// ContractToMachineSensorExists test if a given machine sensor combination exists in the database or not
func ContractToMachineSensorExists(db *sql.DB, machine, sensor string) (bool, error) {
	query, err := db.Query("SELECT contract FROM contract_machine_sensor JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor WHERE machine = $1 AND sensor = $2", machine, sensor)
	if err != nil {
		return false, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			klog.Errorf("cannot close query object: %s\n", err)
		}
	}()

	if query.Next() {
		return true, nil
	}
	return false, nil
}

// ContractRemove removes a given contract
func ContractRemove(db *sql.DB, contract string) error {
	_, err := db.Exec("DELETE FROM contract WHERE contract = $1", contract)
	return err
}

// MinDuration checks the minimum duration of a machine sensor combination with a defined version
func MinDuration(db *sql.DB, machine, sensor, version string) (time.Duration, error) {
	data, err := db.Query("SELECT duration FROM contract JOIN contract_machine_sensor ON contract_machine_sensor.contract = contract.contract JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor WHERE machine = $1 AND sensor = $2 AND version = $3 ORDER BY duration ASC LIMIT 1", machine, sensor, version)
	if err != nil {
		return time.Minute, err
	}

	var duration time.Duration
	var daString string

	if !data.Next() {
		return time.Minute, fmt.Errorf("no entry in the database found for this machine, sensor, version combination")
	}

	if err := data.Scan(&daString); err != nil {
		return time.Minute, err
	}

	duration, err = time.ParseDuration(daString)
	return duration, err
}
