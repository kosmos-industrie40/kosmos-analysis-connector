package mapper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/connection"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/db"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/mqtt"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/uploader"
)

// Contract contains the logic to handle a contract message
type Contract struct {
	connector *connection.Connection
	db        *sql.DB
	version   string
	uploader  *uploader.Sensor
}

// NewContractMapper initialise the contract struct
func NewContractMapper(mClient mqtt.Mqtt, connector *connection.Connection, version string, db *sql.DB, upload *uploader.Sensor) Contract {
	var c Contract
	c.connector = connector
	c.version = version
	c.db = db
	c.uploader = upload
	klog.Infof("subscribe to contracts create")
	if err := mClient.Subscribe("kosmos/contracts/create", c.createMessageHandler); err != nil {
		klog.Errorf("cannot subscribe to kosmos/contracts/create: %s\n", err)
	}

	klog.Infof("subscribe to contracts/all")
	if err := mClient.Subscribe("kosmos/contracts/all", c.allMessageHandler); err != nil {
		klog.Errorf("cannot subscribe to kosmos/contracts/all: %s", err)
	}

	klog.Infof("subscribe to contracts delete")
	if err := mClient.Subscribe("kosmos/contracts/delete", c.deleteMessageHandler); err != nil {
		klog.Errorf("cannot subscribe to kosmos/contracts/delete: %s\n", err)
	}

	return c
}

func (c Contract) deleteMessageHandler(client MQTT.Client, m MQTT.Message) {
	klog.Infof("handle contract delete message")
	var dCon struct {
		Body struct {
			Contract string `json:"contract"`
		} `json:"body"`
	}

	if err := json.Unmarshal(m.Payload(), &dCon); err != nil {
		klog.Errorf("cannot unmarshal contract deletion message")
		return
	}

	req, err := c.connector.Request("DELETE", fmt.Sprintf("contract/%s", dCon.Body.Contract), nil, strings.NewReader(""))
	if err != nil {
		klog.Errorf("cannot successful delete contract: %s\n", err)
	}

	if req.StatusCode != 204 {
		klog.Errorf("cannot successful delete contract; transmitted status code is: %d", req.StatusCode)
	}

	machineSensor, err := db.GetMachineSensorFromContract(c.db, dCon.Body.Contract)
	if err != nil {
		klog.Errorf("cannot get machineSensor from a contract %s err: %s", dCon.Body.Contract, err)
	}
	if err := db.ContractRemove(c.db, dCon.Body.Contract); err != nil {
		klog.Errorf("cannot remove contract from db %s", err)
	}

	for _, v := range machineSensor {
		exists, err := db.ContractToMachineSensorExists(c.db, v.Machine, v.Sensor)
		if err != nil {
			klog.Errorf("cannot query existence of machine sensor: %s\n", err)
		}

		if !exists {
			return
		}

		duration, err := db.MinDuration(c.db, v.Machine, v.Sensor, c.version)
		if err != nil {
			klog.Errorf("cannot receive duration: %s", err)
		}

		c.uploader.ChangeInterval(v.Machine, v.Sensor, duration)
	}
}

func (c Contract) allMessageHandler(client MQTT.Client, m MQTT.Message) {
	klog.Infof("receive mqtt message to handler all contracts")
	klog.V(2).Infof("qos: %d, duplication: %t, messageID: %d", m.Qos(), m.Duplicate(), m.MessageID())

	var contracts []mqtt.Contract
	if err := json.Unmarshal(m.Payload(), &contracts); err != nil {
		klog.Errorf("cannot unmarshal contract message: %s\n", err)
		return
	}

	var mcCon []connection.Contract
	for _, con := range contracts {
		cCon, analysisCloud, found := c.convertContract(con)
		if found {
			continue
		}

		for _, v := range cCon.Body.Sensors {
			if err := db.Insert(c.db, cCon.Body.Machine, v.Name, analysisCloud.Connection.Interval, c.version, cCon.Body.Contract.ID); err != nil {
				klog.Errorf("cannot insert new contract into database: %s", err)
				continue
			}

			duration, err := time.ParseDuration(analysisCloud.Connection.Interval)
			if err != nil {
				klog.Errorf("duration parsing uploading interval failed: %s", err)
				continue
			}

			c.uploader.StartHandler(cCon.Body.Machine, v.Name, duration)

			dura, err := db.MinDuration(c.db, cCon.Body.Machine, v.Name, c.version)
			if err != nil {
				klog.Errorf("cannot receive min duration: %s\n", err)
				continue
			}

			klog.Infof("change interval")

			c.uploader.ChangeInterval(cCon.Body.Machine, v.Name, dura)
		}

		mcCon = append(mcCon, cCon)

	}

	byteData, err := json.Marshal(mcCon)
	if err != nil {
		klog.Errorf("cannot marshal connector contract: %s\n", err)
	}

	klog.Infof("start to make the http request; with data\n%s", string(byteData))
	req, err := c.connector.Request("POST", "contract/", nil, strings.NewReader(string(byteData)))
	if err != nil {
		klog.Errorf("cannot upload contract to analysis cloud %s\n", err)
	}

	if req.StatusCode != 201 {
		klog.Errorf("status code of post contract has not the expected value with %d", req.StatusCode)
	}

}

func (c Contract) convertContract(mCon mqtt.Contract) (connection.Contract, connection.ContractAnalysisSystem, bool) {
	var analysisCloud connection.ContractAnalysisSystem

	found := false
	for _, v := range mCon.Body.Analysis.Systems {
		if v.System == "cloud" {
			analysisCloud = connection.ContractAnalysisSystem{
				System: v.System,
				Enable: v.Enable,
				Connection: connection.AnalysisConnection{
					Container: connection.Container{
						URL:         v.Connection.Container.URL,
						Tag:         v.Connection.Container.Tag,
						Arguments:   v.Connection.Container.Arguments,
						Environment: v.Connection.Container.Environment,
					},
					Interval: v.Connection.Interval,
					URL:      v.Connection.URL,
					UserMgmt: v.Connection.UserMgmt,
				},
			}
			for _, y := range v.Pipelines {
				pipe := connection.Pipelines{
					Sensor: y.Sensor,
					MlTrigger: connection.PipelinesMlTrigger{
						Type: y.MlTrigger.Type,
						Definition: connection.PipelinesMlTriggerDefinition{
							After: y.MlTrigger.Definition.After,
						},
					},
				}

				for _, x := range y.Pipeline {
					pip := connection.PipelinesPipeline{
						Container: connection.Container{
							URL:         x.Container.URL,
							Tag:         x.Container.Tag,
							Arguments:   x.Container.Arguments,
							Environment: x.Container.Environment,
						},
						PersistOutput: x.PersistOutput,

						From: (*connection.Model)(x.From),
						To:   (*connection.Model)(x.To),
					}
					pipe.Pipeline = append(pipe.Pipeline, pip)
				}
				analysisCloud.Pipelines = append(analysisCloud.Pipelines, pipe)
			}
			found = true
			break
		}
	}

	if !found {
		return connection.Contract{}, connection.ContractAnalysisSystem{}, false
	}

	if !analysisCloud.Enable || analysisCloud.System == "" {
		klog.Infof("analysis cloud is not enabled in contract: %s", mCon.Body.Contract.ID)
		return connection.Contract{}, connection.ContractAnalysisSystem{}, false
	}

	var cCon connection.Contract
	cCon.Body.Analysis.Enable = mCon.Body.Analysis.Enable
	cCon.Body.Analysis.Systems = append(cCon.Body.Analysis.Systems, analysisCloud)

	cCon.Body.Contract = connection.ContractInfos{
		Valid: connection.ContractInfosValid{
			Start: mCon.Body.Contract.Valid.Start,
			End:   mCon.Body.Contract.Valid.End,
		},
		ID:           mCon.Body.Contract.ID,
		CreationTime: mCon.Body.Contract.CreationTime,
		Partners:     mCon.Body.Contract.Partners,
		Permissions: connection.Permissions{
			Read:  mCon.Body.Contract.Permissions.Read,
			Write: mCon.Body.Contract.Permissions.Write,
		},
		Version:        mCon.Body.Contract.Version,
		ParentContract: mCon.Body.Contract.ParentContract,
	}

	cCon.Body.Machine = mCon.Body.Machine
	cCon.Body.KosmosLocalSystems = mCon.Body.KosmosLocalSystems

	klog.Infof("count of sensors in the contract: %d", len(mCon.Body.Sensors))
	for _, v := range mCon.Body.Sensors {
		var sensor connection.ContractSensor
		sensor.Name = v.Name
		sensor.Meta = v.Meta
		for _, v := range v.StorageDuration {
			var storage connection.ContractSensorDuration
			storage.Duration = v.Duration
			storage.Meta = v.Meta
			storage.SystemName = v.SystemName
			sensor.StorageDuration = append(sensor.StorageDuration, storage)
		}
		cCon.Body.Sensors = append(cCon.Body.Sensors, sensor)
	}
	return cCon, analysisCloud, true
}

// createMessageHandler is the function that is called everytime a contract is sent to the MQTT-Topic
// kosmos/contracts/create. The Contract is then parsed and written to the database as well as send
// to the cloud.
func (c Contract) createMessageHandler(client MQTT.Client, m MQTT.Message) {
	klog.Info("receive mqtt message to handle a contract")
	klog.Infof("qos: %d, duplication: %t, messageID: %d", m.Qos(), m.Duplicate(), m.MessageID())
	// Unmarshal received Contract into a Golang object
	var mCon mqtt.Contract
	if err := json.Unmarshal(m.Payload(), &mCon); err != nil {
		klog.Errorf("can not unmarshal contract message: %s\n", err)
		return
	}
	// Convert contract into parts which are relevant to the cloud e.g. pipelines
	cCon, analysisCloud, found := c.convertContract(mCon)
	if !found {
		return
	}
	// For every sensor in the contract...
	for _, v := range cCon.Body.Sensors {

		// ... store the sensor in the database...
		if err := db.Insert(c.db, mCon.Body.Machine, v.Name, analysisCloud.Connection.Interval, c.version, mCon.Body.Contract.ID); err != nil {
			klog.Errorf("Can not insert new contract into database: %s", err)
			return
		}

		// ...parse the frequency with which data is sent to the cloud...
		duration, err := time.ParseDuration(analysisCloud.Connection.Interval)
		if err != nil {
			klog.Errorf("Duration parsing uploading interval failed: %s", err)
			continue
		}

		//... start a mqtt-handler which subscribes to the necessary data topics...
		klog.Infof("start handle machine %s sensor %s and duration %s", mCon.Body.Machine, v.Name, duration)
		sensor_mapper := SensorData{}
		buf := c.uploader.GetBuffer()
		mClient := mqtt.NewMqtt(client)
		if err := sensor_mapper.Init(mClient, *buf, mCon.Body.Machine, v.Name); err != nil {
			klog.Errorf("cannot create mapper on machine %s sensor %s and duration %v", mCon.Body.Machine, v.Name, duration)
			return
		}

		// ...and start an upload-handler which sends the data to the cloud...
		c.uploader.StartHandler(mCon.Body.Machine, v.Name, duration)
		//...parse the minimal frequency per sensor so only that frequency is used...
		dura, err := db.MinDuration(c.db, mCon.Body.Machine, v.Name, c.version)
		if err != nil {
			klog.Errorf("Can not receive minimal duration: %s\n", err)
			return
		}

		// ...change the interval for the upload handler to the minimal frequency.
		klog.Infof("Change interval...")
		c.uploader.ChangeInterval(mCon.Body.Machine, v.Name, dura)
	}

	klog.Infof("Marshal JSON of new contract...")
	byteData, err := json.Marshal(cCon)
	if err != nil {
		klog.Errorf("Cannot marshal connector contract: %s\n", err)
	}

	klog.Infof("Start to make the http request; with data\n%s", string(byteData))
	req, err := c.connector.Request("POST", "contract/", nil, strings.NewReader(string(byteData)))
	if err != nil {
		klog.Errorf("Can not upload contract to analyses cloud: %s\n", err)
		return
	}

	if req.StatusCode != 201 {
		klog.Errorf("status code of post contract has not the expected value with %d", req.StatusCode)
	}
}
