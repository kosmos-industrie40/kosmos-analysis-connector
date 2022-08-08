package db

import (
	"fmt"
	"testing"
	"time"

	dbMock "github.com/DATA-DOG/go-sqlmock"
)

func TestContractExists(t *testing.T) {
	testTable := []struct {
		description string
		err         error
		contract    string
		rows        *dbMock.Rows
		exists      bool
	}{
		{
			description: "sucessfull returned true",
			err:         nil,
			contract:    "contract",
			rows:        dbMock.NewRows([]string{"contract"}).AddRow("contract"),
			exists:      true,
		},
		{
			description: "sucessfull returned false",
			err:         nil,
			contract:    "contract",
			rows:        dbMock.NewRows([]string{"contract"}),
			exists:      false,
		},
	}

	for _, test := range testTable {
		t.Run(test.description, func(t *testing.T) {
			db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("cannot open datbase mock: %s", err)
			}

			defer db.Close()

			que := mock.ExpectQuery("SELECT contract FROM contract WHERE contract = $1")

			if err != nil {
				que.WithArgs(test.contract)
				que.WillReturnError(test.err)
			} else {
				que.WithArgs(test.contract)
				que.WillReturnRows(test.rows)
			}

			exists, err := ContractExists(db, test.contract)
			t.Logf("contract existens returnes: %t, %s", exists, err)
			if err != nil && test.err != nil {
				if err.Error() != test.err.Error() {
					t.Errorf("returned err != expected error\n\t%s != %s", err, test.err)
				}
			} else if err != nil || test.err != nil {
				t.Errorf("returned err != expected error\n\t%s != %s", err, test.err)
			}

			if exists != test.exists {
				t.Errorf("returned value != not expected value\n\t%t != %t", exists, test.exists)
			}
		})
	}
}

func TestInsertMachineSensorExists(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot open database mock: %s", err)
	}

	defer db.Close()

	mock.ExpectQuery("SELECT id FROM machine_sensor WHERE machine = $1 AND sensor = $2").
		WithArgs("machine", "sensor").
		WillReturnRows(dbMock.NewRows([]string{"id"}).AddRow(4))
	mock.ExpectQuery("SELECT * from contract WHERE contract = $1").
		WithArgs("contract").
		WillReturnRows(dbMock.NewRows([]string{"contract"}))
	mock.ExpectExec("INSERT INTO contract (contract, duration, version) VALUES ($1, $2, $3)").
		WithArgs("contract", "duration", "version").
		WillReturnResult(dbMock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO contract_machine_sensor (contract, machine_sensor) VALUES ($1, $2)").
		WithArgs("contract", 4).
		WillReturnResult(dbMock.NewResult(1, 1))

	if err := Insert(db, "machine", "sensor", "duration", "version", "contract"); err != nil {
		t.Errorf("cannot insert into database %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestInsertMachineSensorNotExists(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot open database mock: %s", err)
	}

	defer db.Close()

	mock.ExpectQuery("SELECT id FROM machine_sensor WHERE machine = $1 AND sensor = $2").
		WithArgs("machine", "sensor").
		WillReturnRows(dbMock.NewRows([]string{"id"}))
	mock.ExpectQuery("INSERT INTO machine_sensor (machine, sensor) VALUES ($1, $2) RETURNING id").
		WithArgs("machine", "sensor").
		WillReturnRows(dbMock.NewRows([]string{"id"}).AddRow(4))
	mock.ExpectQuery("SELECT * from contract WHERE contract = $1").
		WithArgs("contract").
		WillReturnRows(dbMock.NewRows([]string{"contract"}))
	mock.ExpectExec("INSERT INTO contract (contract, duration, version) VALUES ($1, $2, $3)").
		WithArgs("contract", "duration", "version").
		WillReturnResult(dbMock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO contract_machine_sensor (contract, machine_sensor) VALUES ($1, $2)").
		WithArgs("contract", 4).
		WillReturnResult(dbMock.NewResult(1, 1))

	if err := Insert(db, "machine", "sensor", "duration", "version", "contract"); err != nil {
		t.Errorf("cannot insert into database %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestContractRemove(t *testing.T) {
	db, mock, err := dbMock.New()
	if err != nil {
		t.Fatalf("cannot oben mock db")
	}

	defer db.Close()

	mock.ExpectExec("DELETE FROM contract").WithArgs("contract").WillReturnResult(dbMock.NewResult(1, 1))

	if err := ContractRemove(db, "contract"); err != nil {
		t.Errorf("cannot remove contract: %s\n", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestContractToMachineSensorExists_true(t *testing.T) {
	db, mock, err := dbMock.New()
	if err != nil {
		t.Fatalf("cannot open mocked db: %s\n", err)
	}

	defer db.Close()

	mock.ExpectQuery("SELECT contract FROM contract_machine_sensor JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor").
		WithArgs("machine", "sensor").
		WillReturnRows(dbMock.NewRows([]string{"contract"}).AddRow("contract"))

	ret, err := ContractToMachineSensorExists(db, "machine", "sensor")
	if err != nil {
		t.Errorf("unexpected returned error %s", err)
	}

	if !ret {
		t.Errorf("unexpected return value with false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestContractToMachineSensorExists_false(t *testing.T) {
	db, mock, err := dbMock.New()
	if err != nil {
		t.Fatalf("cannot open mocked db: %s\n", err)
	}

	defer db.Close()

	mock.ExpectQuery("SELECT contract FROM contract_machine_sensor JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor").
		WithArgs("machine", "sensor").
		WillReturnRows(dbMock.NewRows([]string{"contract"}))

	ret, err := ContractToMachineSensorExists(db, "machine", "sensor")
	if err != nil {
		t.Errorf("unexpected returned error %s", err)
	}

	if ret {
		t.Errorf("unexpected return value with true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestQueryIntervalErrorDb(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	usedErr := fmt.Errorf("error")
	mock.ExpectQuery("SELECT duration FROM contract JOIN contract_machine_sensor ON contract_machine_sensor.contract = contract.contract JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor WHERE machine = $1 AND sensor = $2 AND version = $3 ORDER BY duration ASC LIMIT 1").
		WithArgs("machine", "sensor", "version").
		WillReturnError(usedErr)

	_, err = MinDuration(db, "machine", "sensor", "version")
	if err != usedErr {
		t.Errorf("returned error doesn't match the expected error %s != %s", err, usedErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestQueryIntervalNoEntry(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	mock.ExpectQuery("SELECT duration FROM contract JOIN contract_machine_sensor ON contract_machine_sensor.contract = contract.contract JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor WHERE machine = $1 AND sensor = $2 AND version = $3 ORDER BY duration ASC LIMIT 1").
		WithArgs("machine", "sensor", "version").
		WillReturnRows(dbMock.NewRows([]string{"inverval"}))

	_, err = MinDuration(db, "machine", "sensor", "version")
	if err.Error() != "no entry in the database found for this machine, sensor, version combination" {
		t.Errorf("returned error doesn't match the expected error %s != %s", err, "no entry in the database found for this machine, sensor, version combination")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestQueryIntervalNotParseDuration(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	mock.ExpectQuery("SELECT duration FROM contract JOIN contract_machine_sensor ON contract_machine_sensor.contract = contract.contract JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor WHERE machine = $1 AND sensor = $2 AND version = $3 ORDER BY duration ASC LIMIT 1").
		WithArgs("machine", "sensor", "version").
		WillReturnRows(dbMock.NewRows([]string{"inverval"}).AddRow("nein"))

	_, err = MinDuration(db, "machine", "sensor", "version")
	if err.Error() != "time: invalid duration \"nein\"" {
		t.Errorf("returned error doesn't match the expected error\n\t%s != %s", err, "time: invalid duration \"nein\"")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestQueryInterval(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	mock.ExpectQuery("SELECT duration FROM contract JOIN contract_machine_sensor ON contract_machine_sensor.contract = contract.contract JOIN machine_sensor ON machine_sensor.id = contract_machine_sensor.machine_sensor WHERE machine = $1 AND sensor = $2 AND version = $3 ORDER BY duration ASC LIMIT 1").
		WithArgs("machine", "sensor", "version").
		WillReturnRows(dbMock.NewRows([]string{"inverval"}).AddRow("5m"))

	duration, err := MinDuration(db, "machine", "sensor", "version")
	if err != nil {
		t.Errorf("returned error doesn't match the expected error %s != nil", err)
	}

	dura := 5 * time.Minute
	if duration != dura {
		t.Errorf("unexpected duration: %v != %v", duration, dura)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}
