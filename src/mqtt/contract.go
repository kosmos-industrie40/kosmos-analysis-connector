package mqtt

// Contract is the contract message definition
type Contract struct {
	Body      ContractBody
	Signature Signature
}

// ContractBody contains the contract body
type ContractBody struct {
	Contract                    ContractInfos    `json:"contract"`
	Machine                     string           `json:"machine"`
	RequiredTechnicalContainers []TechContainer  `json:"requiredTechnicalContainers"`
	KosmosLocalSystems          []string         `json:"kosmosLocalSystems"`
	Sensors                     []ContractSensor `json:"sensors"`
	CheckSignature              bool             `json:"checkSignatures"`
	Analysis                    ContractAnalysis `json:"analysis"`
	Blockchain                  interface{}      `json:"blockchain"`
	MachineConnection           interface{}      `json:"machineConnection"`
	Metadata                    interface{}      `json:"metadata"`
}

// TechContainer defines the required technical containers
type TechContainer struct {
	System     string      `json:"system"`
	Containers []Container `json:"containers"`
}

// ContractInfos contains contract info specifics
type ContractInfos struct {
	Valid          ContractInfosValid `json:"valid"`
	ParentContract string             `json:"parentContract"`
	CreationTime   string             `json:"creationTime"`
	Partners       []string           `json:"partners"`
	Permissions    Permissions        `json:"permissions"`
	ID             string             `json:"id"`
	Version        string             `json:"version"`
}

// Permissions defines the contract permissions
type Permissions struct {
	Read  []string `json:"read"`
	Write []string `json:"write"`
}

// ContractInfosValid contains the information about the validation duration
// of the contract
type ContractInfosValid struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ContractSensorDuration is the storage duration fo the sensor
type ContractSensorDuration struct {
	SystemName string      `json:"systemName"`
	Duration   string      `json:"duration"`
	Meta       interface{} `json:"meta"`
}

// ContractSensor contains the contract sensor information
type ContractSensor struct {
	Name            string `json:"name"`
	StorageDuration []ContractSensorDuration
	Meta            interface{} `json:"meta"`
}

// ContractAnalysis defines the analysis in the contract
type ContractAnalysis struct {
	Enable  bool                     `json:"enable"`
	Systems []ContractAnalysisSystem `json:"systems"`
}

// ContractAnalysisSystem contains the definition of each system
type ContractAnalysisSystem struct {
	System     string             `json:"system"`
	Enable     bool               `json:"enable"`
	Pipelines  []Pipelines        `json:"pipelines"`
	Connection AnalysisConnection `json:"connection"`
}

// Pipelines is the analysis pipeline
type Pipelines struct {
	MlTrigger PipelinesMlTrigger  `json:"ml-trigger"`
	Pipeline  []PipelinesPipeline `json:"pipeline"`
	Sensor    []string            `json:"sensors"`
}

// PipelinesPipeline defines each pipeline
type PipelinesPipeline struct {
	Container     Container `json:"container"`
	PersistOutput bool      `json:"persistOutput"`
	From          *Model    `json:"from"`
	To            *Model    `json:"to"`
}

// Container is the container in the contract message
type Container struct {
	URL         string   `json:"url"`
	Tag         string   `json:"tag"`
	Arguments   []string `json:"arguments"`
	Environment []string `json:"environment"`
}

// Model is the model in the contract message
type Model struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

// PipelinesMlTriggerDefinition contains the definition of the ml trigger
type PipelinesMlTriggerDefinition struct {
	After string `json:"after"`
}

// PipelinesMlTrigger contains the datatype of the ml trigger
type PipelinesMlTrigger struct {
	Type       string                       `json:"type"`
	Definition PipelinesMlTriggerDefinition `json:"definition"`
}

// AnalysisConnection defines the connection to the analysis platform
type AnalysisConnection struct {
	Container Container `json:"container"`
	Interval  string    `json:"interval"`
	UserMgmt  string    `json:"user-mgmt"`
	URL       string    `json:"url"`
}
