package connection

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Message defines the message, which should be uploaded to the analysis cloud
type message struct {
	//gorm.Model
	Method  string
	Address string
	Message []byte
}

// Persist is the interface, to store enable a persistent in this tool
type Persist interface {
	Close() error
	Insert([]Message)
	Remove([]Message)
	Query() []Message
}

// NewPersistPostgreSQL create a new Persist Tool and using PostgreSQL in the background
func NewPersistPostgreSQL(host, user, password, database string, port int) (Persist, error) {
	var conStr string
	if password == "" {
		conStr = fmt.Sprintf("host=%s user=%s port=%d sslmode=disable dbname=%s",
			host,
			user,
			port,
			database,
		)
	} else {
		conStr = fmt.Sprintf("host=%s user=%s password=%s port=%d sslmode=disable dbname=%s",
			host,
			user,
			password,
			port,
			database,
		)
	}
	
	db, err := gorm.Open("postgres", conStr)
	if err != nil {
		return persist{}, err
	}

	db.AutoMigrate(&message{})
	return persist{db: db}, nil
}

type persist struct {
	db *gorm.DB
}

// Close the database connection
func (p persist) Close() error {
	return p.db.Close()
}

// Message represent a message which should be send to ther analysis platform
type Message struct {
	// Method describe the used REST method
	Method string
	// Address is the current address of the analysis platform
	Address string
	// Message contains the complete message of the analysis platform
	Message []byte
}

// Insert insert a message into the database
func (p persist) Insert(msg []Message) {
	for _, v := range msg {
		var ms message
		ms.Address = v.Address
		ms.Message = v.Message
		ms.Method = v.Method

		p.db.Create(&ms)
	}
}

func (p persist) Remove(msg []Message) {
	for _, v := range msg {
		p.db.Where("address = ? AND message = ?", v.Address, v.Message).Delete(message{})
	}
}

func (p persist) Query() []Message {
	var ms []message
	p.db.Find(&ms)

	var msg []Message
	for _, v := range ms {
		msg = append(msg, Message(v))
	}

	return msg
}
