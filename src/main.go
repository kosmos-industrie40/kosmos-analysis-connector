package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/auth"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/buffer"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/connection"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/constants"
	database "github.com/kosmos-industrie40/kosmos-analyse-connector/src/db"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/mapper"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/mqtt"
	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/uploader"
)

var (
	cli struct {
		ConfigFile  string
		Monitoring  string
		WriteConfig bool
		ConfigDir   string
	}
	vi      *viper.Viper
	version = "0"
)

type status struct {
	Body struct {
		From   string `json:"connector"`
		Status string `json:"status"`
	} `json:"body"`
}

func init() {
	klog.InitFlags(nil)

	flag.StringVar(&cli.ConfigFile, "config", "exampleConfiguration.yaml", "is the name to the configuration file")
	flag.StringVar(&cli.Monitoring, "address", ":8081", "The address to listen for the http requests of prometheus.")
	flag.BoolVar(&cli.WriteConfig, "configDefault", false, "generates a default configuration and exit the program")

	flag.Parse()

	vi = viper.New()
	vi.AutomaticEnv()
	vi.SetEnvPrefix("cc")
	vi.SetConfigFile(cli.ConfigFile)

	// set default values in the configuration

	// edge
	// database
	vi.SetDefault(constants.EdgeDatabaseURL, "localhost")
	vi.SetDefault(constants.EdgeDatabasePort, 5432)
	vi.SetDefault(constants.EdgeDatabaseUser, "kosmos")
	vi.SetDefault(constants.EdgeDatabasePassword, "")
	vi.SetDefault(constants.EdgeDatabaseDatabase, "edge")

	// mqtt
	vi.SetDefault(constants.EdgeMqttURL, "localhost")
	vi.SetDefault(constants.EdgeMqttPort, 1883)
	vi.SetDefault(constants.EdgeMqttUser, "")
	vi.SetDefault(constants.EdgeMqttPassword, "")

	// analysis cloud
	// connector
	vi.SetDefault(constants.AnalysisCloudConnectorURL, "localhost")
	vi.SetDefault(constants.AnalysisCloudConnectorPort, 80)

	// userMgmt
	vi.SetDefault(constants.AnalysisCloudUserMgmtSchema, "https")
	vi.SetDefault(constants.AnalysisCloudUserMgmtPath, "auth")
	vi.SetDefault(constants.AnalysisCloudUserMgmtPort, 443)
	vi.SetDefault(constants.AnalysisCloudUserMgmtUser, "test user")
	vi.SetDefault(constants.AnalysisCloudUserMgmtPassword, "")

	// read in configuration
	err := vi.ReadInConfig()
	if err != nil {
		fmt.Printf("cannot read configuration:\n")
		fmt.Println(err)
		os.Exit(1)
	}

	if cli.WriteConfig {
		if err := vi.WriteConfig(); err != nil {
			klog.Errorf("cannot write down config: %s", err)
			os.Exit(1)
		}
		fmt.Printf("The default config is written to: %s/%s\n", cli.ConfigDir, cli.ConfigFile)
		os.Exit(0)
	}
}

func sendStatus(mqtt mqtt.Mqtt) {
	for {
		var stat status
		stat.Body.From = "analysis"
		stat.Body.Status = "alive"
		time.Sleep(1 * time.Minute)
		dat, err := json.Marshal(stat)
		if err != nil {
			klog.Errorf("cannot marshal status: %s", err)
		}

		if err := mqtt.Send("kosmos/status", dat); err != nil {
			klog.Errorf("cannot publish status: %s", err)
		}
	}
}

func main() {

	var mqttClient mqtt.Mqtt
	var err error
	for i := 0; i < 10; i++ {
		err = mqttClient.Connect(
			vi.GetString(constants.EdgeMqttURL),
			vi.GetString(constants.EdgeMqttUser),
			vi.GetString(constants.EdgeMqttPassword),
			"static",
			vi.GetInt(constants.EdgeMqttPort),
			&tls.Config{},
		)
		if err != nil {
			klog.Infof("MQTT connection retry: %d/10\n", i+1)
			time.Sleep(15 * time.Second)
		} else {
			klog.Info("Connected to MQTT-Broker!")
			i = 10
		}
	}
	if err != nil {
		klog.Errorf("cannot connect to the mqtt broker: %s\n", err)
		os.Exit(1)
	}

	go sendStatus(mqttClient)

	tokenChan := make(chan auth.Token, 2)
	auth := auth.NewOidcAuth(
		tokenChan,
		vi.GetString(constants.AnalysisCloudUserMgmtSchema),
		vi.GetString(constants.AnalysisCloudUserMgmtURL),
		vi.GetString(constants.AnalysisCloudUserMgmtPath),
		vi.GetInt(constants.AnalysisCloudUserMgmtPort),
		vi.GetString(constants.AnalysisCloudUserMgmtUser),
		vi.GetString(constants.AnalysisCloudUserMgmtPassword),
	)

	err = auth.Login()
	if err != nil {
		klog.Errorf("cannot login to the system: %v", err)
		os.Exit(1)
	}

	var persist connection.Persist
	for i := 0; i < 10; i++ {
		persist, err = connection.NewPersistPostgreSQL(
			vi.GetString(constants.EdgeDatabaseURL),
			vi.GetString(constants.EdgeDatabaseUser),
			vi.GetString(constants.EdgeDatabasePassword),
			vi.GetString(constants.EdgeDatabaseDatabase),
			vi.GetInt(constants.EdgeDatabasePort),
		)
		if err != nil {
			klog.Infof("DB connection retry: %d/10\n", i+1)
			time.Sleep(15 * time.Second)
		} else {
			klog.Info("Connected to database!")
			i = 10
		}
	}
	if err != nil {
		klog.Errorf("cannot create persist tooling: %s", err)
		time.Sleep(15 * time.Second)
		os.Exit(1)
	}

	endpoint := connection.NewConnection(fmt.Sprintf("%s:%d", vi.GetString(constants.AnalysisCloudConnectorURL), vi.GetInt(constants.AnalysisCloudConnectorPort)), tokenChan, persist)

	buf := buffer.NewLocalBuffer()

	//var uploaderSens uploader.UploaderSensor
	//uploaderSens := uploader.InitUploaderSensor(buf, endpoint)
	uploaderSens := uploader.Sensor{}
	uploaderSensor := &uploaderSens
	uploaderSensor.Init(buf, endpoint)

	conStr := fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=disable dbname=%s",
		vi.GetString(constants.EdgeDatabaseURL),
		vi.GetString(constants.EdgeDatabaseUser),
		vi.GetString(constants.EdgeDatabasePassword),
		vi.GetInt(constants.EdgeDatabasePort),
		vi.GetString(constants.EdgeDatabaseDatabase),
	)
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		klog.Errorf("cannot connect to database: %s\n", err)
		os.Exit(1)
	}

	_ = mapper.NewContractMapper(mqttClient, endpoint, version, db, uploaderSensor)

	sensors, err := database.HandleSensors(db, version)
	if err != nil {
		klog.Errorf("cannot receive sensors which should be handled: %s", err)
	}

	for _, v := range sensors {
		klog.Infof("start handle machine %s sensor %s and duration %s", v.Machine, v.Sensor, v.Duration)
		mapper := mapper.SensorData{}
		if err := mapper.Init(mqttClient, buf, v.Machine, v.Sensor); err != nil {
			klog.Errorf("cannot create mapper on machine %s sensor %s and duration %v", v.Machine, v.Sensor, v.Duration)
			os.Exit(1)
		}

		duration, err := time.ParseDuration(v.Duration)
		if err != nil {
			klog.Errorf("cannot parse v.Duration %s err: %s", v.Duration, err)
			os.Exit(1)
		}

		uploaderSensor.StartHandler(v.Machine, v.Sensor, duration)
	}

	http.Handle("/metrics", promhttp.Handler())
	klog.Fatal(http.ListenAndServe(cli.Monitoring, nil))
}
