package connection

import (
	"net/http"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func initDb() (Persist, *gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		return persist{}, nil, err
	}

	db.AutoMigrate(&message{})
	return persist{db: db}, db, nil
}

func cleanUp(db *gorm.DB) {
	db.DropTable(&message{})
}

func TestInsert(t *testing.T) {
	testTable := []struct {
		msg         []Message
		description string
	}{
		{
			msg: []Message{
				{
					Address: "localhost",
					Message: []byte("hallo welt"),
					Method:  http.MethodGet,
				},
			},
			description: "insert one element",
		},
		{
			msg: []Message{
				{
					Address: "localhost",
					Message: []byte("hallo welt"),
					Method:  http.MethodGet,
				},
				{
					Address: "local",
					Message: []byte("hallo welt"),
					Method:  http.MethodGet,
				},
			},
			description: "insert two elements",
		},
	}

	for _, test := range testTable {
		t.Run(test.description, func(t *testing.T) {
			pers, db, err := initDb()
			if err != nil {
				t.Fatal(err)
			}

			pers.Insert(test.msg)

			var ms []message
			db.Find(&ms)

			if len(ms) != len(test.msg) {
				t.Errorf("length of the inserted elements != expected length")
			}

			for i := 0; i > len(ms); i++ {
				found := false
				for _, m := range test.msg {
					if m.Address == ms[i].Address {
						found = true
					} else {
						continue
					}

					break
				}

				if !found {
					t.Errorf("address, msg could not found")
				}

			}

			cleanUp(db)
			if err := pers.Close(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	testTable := []struct {
		description string
		msg         []Message
	}{
		{
			description: "remove one",
			msg: []Message{
				{
					Address: "foo",
					Message: []byte("bar"),
				},
			},
		},
		{
			description: "remove two",
			msg: []Message{
				{
					Address: "foo",
					Message: []byte("bar"),
					Method:  http.MethodGet,
				},
				{
					Address: "fool",
					Message: []byte("barl"),
					Method:  http.MethodGet,
				},
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.description, func(t *testing.T) {
			pers, db, err := initDb()
			if err != nil {
				t.Fatal(err)
			}

			for _, m := range test.msg {
				ms := message{Address: m.Address, Message: m.Message}
				db.Create(&ms)
			}

			var ms []message
			db.Find(&ms)
			if len(ms) != len(test.msg) {
				t.Error("expected length is not equal to length in db")
			}

			pers.Remove(test.msg)

			db.Find(&ms)
			if len(ms) != 0 {
				t.Errorf("table is not empty")
			}

			if err := pers.Close(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	testTable := []struct {
		description string
		msg         []Message
	}{
		{
			description: "on element in the table",
			msg: []Message{
				{
					Address: "local",
					Message: []byte("foo"),
					Method:  http.MethodGet,
				},
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.description, func(t *testing.T) {
			pers, db, err := initDb()
			if err != nil {
				t.Fatal(err)
			}

			for _, m := range test.msg {
				ms := message{Message: m.Message, Address: m.Address}
				db.Create(ms)
			}

			msg := pers.Query()

			if len(msg) != len(test.msg) {
				t.Errorf("unexpected return length")
			}

			cleanUp(db)

			if err := pers.Close(); err != nil {
				t.Error(err)
			}
		})
	}
}
