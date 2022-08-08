package buffer

import (
	"encoding/json"
	"testing"

	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/connection"
)

func TestInsert(t *testing.T) {
	timeTable := []struct {
		description string
		machine     string
		sensor      string
		data        []connection.SensorData
	}{
		{
			"insert one element",
			"machine",
			"sensor",
			[]connection.SensorData{
				{
					Signature: "signature",
				},
			},
		},

		{
			"insert two elements",
			"machine",
			"sensor",
			[]connection.SensorData{
				{
					Signature: "signature1",
				},
				{
					Signature: "signature2",
				},
			},
		},
	}

	for _, v := range timeTable {
		var data data
		t.Run(v.description, func(t *testing.T) {
			data.init()

			for _, y := range v.data {
				data.Insert(v.machine, v.sensor, y)
			}

			var found bool = false
			for _, x := range v.data {
				xEncoded, err := json.Marshal(x)
				if err != nil {
					t.Fatalf("cannot encode connection.MachineData to json: %s", err)
				}
				sensor, ok := data.syncMap[v.machine]
				if !ok {
					t.Fatalf("no entry of machine %s found", v.machine)
				}

				mData, ok := sensor[v.sensor]
				if !ok {
					t.Fatalf("no entry of sensor %s found", v.sensor)
				}

				for _, y := range mData {
					yEncoded, err := json.Marshal(y)
					if err != nil {
						t.Fatalf("cannot encode connection.MachineData to json: %s", err)
					}
					if string(yEncoded) == string(xEncoded) {
						found = true
					}
				}

				if !found {
					t.Errorf("cannot found the expected data")
				}
			}
		})
		t.Run("get "+v.description, func(t *testing.T) {
			ret := data.GetValues(v.machine, v.sensor)

			_, ok := data.syncMap[v.machine]
			if ok {
				t.Errorf("sensor does exists: %s\n", v.sensor)
			}

			bRet, err := json.Marshal(ret)
			if err != nil {
				t.Errorf("cannot marshal: %s", err)
			}

			vData, err := json.Marshal(v.data)
			if err != nil {
				t.Errorf("cannot marshal: %s", err)
			}

			if string(vData) != string(bRet) {
				t.Error("the returned value and the expected value are not equal")
			}
		})
	}
}
