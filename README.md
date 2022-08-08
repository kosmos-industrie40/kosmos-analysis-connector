# Analyse Connector (Client)

This tool is an example client of the [analyse connector endpoint](https://github.com/kosmos-industrie40/kosmos-analyses-cloud-connector). To download the full project please use `git clone --recursive`. To update the submodule after cloning you can run `git submodule update --init --recursive`.

## Table of Content
- [Requirements](#requirements)
- [Build](#build)
- [Test](#test)
	- [Unittest](#unittest)
	- [Lint](#lint)
	- [Integration](#integration)
	- [Full](#full)
- [Configuration](#config))

## Requirements
The following programs should be installed in the program if you want to build and test it on the local environment:
- golang (developed with go1.14.7)
- gnu/make (version 4.3)
- golangci-lint (verstion 1.31.0)
- yamllint (1.24.2)
- mosquitto\_sub (version 1.6.12-1)
- postgresql (used version 12.4)
- docker (version 19.03.12)

## Dependencies and Licences
| Dependency | Licence | Run Dependency | URL |
| ---------- | ------- | ---- | --- |
| sqlmock | BSD-3-Clause |  | https://pkg.go.dev/github.com/DATA-DOG/go-sqlmock |
| mqtt | EPL-1.0 | X | https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang |
| yaml | Aache-2.0, MIT | X | https://pkg.go.dev/github.com/go-yaml/yaml |
| pq driver | MIT | X | https://pkg.go.dev/mod/github.com/lib/pq |
| pretty (indirect) | MIT | X | https://pkg.go.dev/mod/github.com/niemeyer/pretty |
| prometheus client | Apache-2.0 | X | https://pkg.go.dev/mod/github.com/prometheus/client\_golang |
| net (indirect) | BSD-3-Clause | X | https://pkg.go.dev/mod/golang.org/x/net@v0.0.0-20200904194848-62affa334b73 |
| check (indirect) | BSD-2-Clause |  | https://pkg.go.dev/mod/gopkg.in/check.v1 |
| klog | Apache-2.0 | X | https://pkg.go.dev/mod/k8s.io/klog |
| gocloak | Apache-2.0 | X | https://pkg.go.dev/mod/github.com/Nerzal/gocloak/v7 |

## Build
The simplest way to build the app on a local system, is the execution of `make`.

To build the docker container local you can execute `docker build -t <your favorite tag> -f Dockerfile .` In the previous command you have to change the string `<your favorite tag>` with the tag you want to use.

To start the application on you local system you can execute `./app` in the root directory of the repository or you can execute it with `docker run <tag>`

## Test
In this chapter the test cases will be described. The CI/CD system, will be executing the lint and the unit tests on every `git push` action.

### Lint

In this tests we are using `golangci-lint` to check the linting of the program code.

### Unittest
There are defined different  unit tests in this program. The can be executed by using `go test src/...`. To find the definition of all unit tests, you can use
the following find command: `find ./ -name '*_test.go'`

### Integration
In the integration tests, you have to set up different different program. The [cloud part of the connector](https://github.com/kosmos-industrie40/kosmos-analyses-cloud-connector)
as described of the project website and a MQTT broker. On the local system,
you have to set up a mqtt broker. To install mosquitto check out [moquitto download webpage](http:s//mosquitto.org/download) or 
a [mosquitto container](https://hub.docker.com/_/eclipse-mosquitto). On the local system, you have to deploy a postgresql as well. This can be make through the steps on [postgresql web page](https://www.postgresql.org/download/) or with a [docker container](https://hub.docker.com/_/postgres). Since the postgres database needs to be preconfigured it can be helpful to use the following Dockerfile
````docker
FROM postgres

COPY createTables.sql /docker-entrypoint-initdb.d/createTables.sql
````
Just make sure to add the following two lines to the beginning of createTables.sql
````sql
CREATE DATABASE kosmos;
\c kosmos;
````

After setting up all the systems and execute them the test is
to publish specific messages on different mqtt topics. This can be made with the `mosquitto_pub` tool or with a container mosquitto-client container. But be careful, it could
be, that you have to change the contract id or other parameters, because they already
exists in the cloud.

The contract messages have to publish to the following topic: `kosmos/contracts/create` the easiest way to do this is the following:
```
mosquitto_pub -t 'kosmos/contracts/create' -f ./kosmos-json-specifications/mqtt_payloads/contract-example.json
```

The sensor upload messages has to be send to one of the following mqtt-topics:
`kosmos/machine-data/84bab968-e6b7-11ea-b10c-54e1ad207114/sensor/temperature/update`
or 
`kosmos/machine-data/84bab968-e6b7-11ea-b10c-54e1ad207114/sensor/alarms/update`
To upload the message you can use the following command:
```
find ./ -name 'data-example*.json' -exec mosquitto_pub -t <topic> -f {} \;
```

The last point is to send a contract deletion message. This can be made through:
```
mosquitto_pub -t kosmos/contracts/delete -f ./kosmos-json-specifications/mqtt_payloads/contractRemove-example.json
```

## Configuration
There are two configuration methods, which both working hand in hand. The first one is the configuration with the cli interface. This method is be used to configure the path of the configuration file and to configure the listen address of the monitoring. The second method is the configuration file. Which will configure the needed connections to the different environments or tools.

### CLI

The cli configuration parameters, the default values and the description is given in the following table:

| parameter | description | default values |
| --------- | ----------- | -------------- |
| config | defines the path, where the configuration file can be found | exampleConfiguration.yaml |
| address | is the listening address of the webserver. The webserver will prove the metrics which can be used with prometheus | :8080 |

### Configuration File
The configuration file is written in yaml. The following table will show the configurations and a description to them.

| parameter | description |
| --------- | ----------- |
| edge | defines the edge components |
| edge.mqtt | defines the mqtt connection on the edge | 
| edge.mqtt.url | defines the url of the mqtt broker |
| edge.mqtt.port | defines the port of the mqtt port |
| edge.mqtt.user | if user password authentication is been used, this will define the used user |
| edge.mqtt.password | if user password authentication is been used, this will define the used password |
| edge.database.url | is the url of the database on the edge |
| edge.database.port | is the port of the database on the edge |
| edge.database.user | is the user of the database on the edge |
| edge.database.password | is the password of the database on the edge |
| database.database | is the name of the database, which stores the used tables |
| analyseCloud | defines the analyse cloud specifics |
| analyseCloud.connector.url | defines the analyse cloud url |
| analyseCloud.connector.port | defines the port where, the analyse cloud endpoint is listening |
| analyseCloud.userMgmt.user | defines the user of the analyse cloud |
| analyseCloud.userMgmt.password | defines the password of the analyse cloud |
| analyseCloud.userMgmt.url | defines the url of the user management of the analyse cloud |
| analyseCloud.userMgmt.port | defines the port, where the user management server listening |
