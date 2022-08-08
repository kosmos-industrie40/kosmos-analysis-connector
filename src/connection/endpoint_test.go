package connection

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestRequest(t *testing.T) {
	testTable := []struct {
		method      string
		err         error
		description string
		data        string
		token       string
		path        string
	}{
		{
			"GET",
			nil,
			"get request successful",
			"data",
			"token",
			"path",
		},
		{
			"POST",
			nil,
			"post request successful",
			"data",
			"token",
			"path",
		},
		{
			"PUT",
			nil,
			"put request successful",
			"data",
			"token",
			"path",
		},
		{
			"HEAD",
			nil,
			"head request successful",
			"data",
			"token",
			"path",
		},
	}

	for _, v := range testTable {
		t.Run(v.description, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != v.method {
					t.Errorf("unexpected method %s expect method %s", r.Method, v.method)
				}

				data, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.Errorf("cannot read transmitted data %s", err)
				}

				if string(data) != v.data {
					t.Errorf("transmitted data != expected data; %s != %s", string(data), v.data)
				}

				if r.Header.Get("token") != v.token {
					t.Errorf("use wrong token: %s != %s", v.token, r.Header.Get("token"))
				}

				if r.URL.Path != "/"+v.path {
					t.Errorf("unexpected path %s != %s", v.path, r.URL.Path)
				}
				w.WriteHeader(200)
			}))

			defer ts.Close()

			db, err := gorm.Open("sqlite3", "test_endpoint.db")
			db.AutoMigrate(&message{})
			if err != nil {
				t.Fatal(err)
			}

			c := Connection{baseURL: ts.URL, token: v.token, persist: persist{db: db}}
			reader := strings.NewReader(v.data)
			_, err = c.Request(v.method, v.path, nil, reader)
			if err != nil {
				t.Errorf("unexpected error %s\n", err)
			}

		})
	}
}
