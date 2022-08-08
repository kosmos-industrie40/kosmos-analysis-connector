package mqtt

// ColumnMeta is the container for column meta messages
type ColumnMeta struct {
	Unit        string `json:"unit"`
	Description string `json:"description"`
}

// Column contains the sensor update column definition
type Column struct {
	Name string     `json:"name"`
	Type string     `json:"type"`
	Meta ColumnMeta `json:"meta"`
}

// Meta contains the sensor update meta data
type Meta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}

// SensorDataBody contains the body of the sensor data
type SensorDataBody struct {
	Timestamp string     `json:"timestamp"`
	Machine   string     `json:"machine"`
	Sensor    string     `json:"sensor"`
	Columns   []Column   `json:"columns"`
	Data      [][]string `json:"data"`
	Meta      []Meta     `json:"meta,omitempty"`
}

// SensorData is the sensor update message
type SensorData struct {
	Body      SensorDataBody `json:"body"`
	Signature string         `json:"signature,omitempty"`
}
