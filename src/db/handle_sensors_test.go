package db

import (
	"fmt"
	"testing"

	dbMock "github.com/DATA-DOG/go-sqlmock"
)

var query string = "SELECT machine, sensor, min(duration) FROM machine_sensor AS ms JOIN contract_machine_sensor AS cms ON ms.id = cms.machine_sensor JOIN contract AS c ON cms.contract = c.contract WHERE version = $1 GROUP BY (machine, sensor)"

func TestQueryHandle_Sensors_Db_Error(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	usedErr := fmt.Errorf("error")
	mock.ExpectQuery(query).WithArgs("version").
		WillReturnError(usedErr)

	_, err = HandleSensors(db, "version")
	if err != usedErr {
		t.Errorf("returned error doesn't match the expected error %s != %s", err, usedErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestQueryHandle_Sensors_Empty_Return(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	mock.ExpectQuery(query).WithArgs("version").
		WillReturnRows(dbMock.NewRows([]string{"machine", "sensor", "duration"}))

	data, err := HandleSensors(db, "version")
	if err != nil {
		t.Errorf("returned error doesn't match the expected error %s != nil", err)
	}

	if len(data) != 0 {
		t.Errorf("unexpected length of the regex: %d", len(data))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestQueryHandle_Sensors_OneResult(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	mock.ExpectQuery(query).WithArgs("version").
		WillReturnRows(dbMock.NewRows([]string{"machine", "sensor", "duration"}).AddRow("machine", "sensor", "duration"))

	data, err := HandleSensors(db, "version")
	if err != nil {
		t.Errorf("returned error doesn't match the expected error %s != nil", err)
	}

	if len(data) != 1 {
		t.Errorf("unexpected length of the regex: %d", len(data))
		return
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}

func TestQueryHandle_Sensors_TwoResult(t *testing.T) {
	db, mock, err := dbMock.New(dbMock.QueryMatcherOption(dbMock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("cannot create mocked db")
	}

	defer db.Close()

	mock.ExpectQuery(query).WithArgs("version").
		WillReturnRows(dbMock.NewRows([]string{"machine", "sensor", "duration"}).AddRow("machine", "sensor", "duration").AddRow("mach1", "sens1", "duration"))

	data, err := HandleSensors(db, "version")
	if err != nil {
		t.Errorf("returned error doesn't match the expected error %s != nil", err)
	}

	if len(data) != 2 {
		t.Errorf("unexpected length of the regex: %d", len(data))
		return
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("not all expectaions were met: %s\n", err)
	}
}
