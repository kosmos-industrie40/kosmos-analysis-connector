// Package connection this package contains the logic to communication with the http endpoint
package connection

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"k8s.io/klog"

	"github.com/kosmos-industrie40/kosmos-analyse-connector/src/auth"
)

// Connection is the connection to the analyse cloud
type Connection struct {
	baseURL   string
	tokenChan <-chan auth.Token
	token     string
	ready     bool
	persist   Persist
	lock      sync.Mutex
}

func (c *Connection) renewToken() {
	for {
		tok := <-c.tokenChan
		c.token = tok.AuthToken()
		c.ready = true
	}
}

// SendMissingData sends data, which are buffered and could not be uploaded previously
func (c *Connection) SendMissingData() {
	for {
		c.lock.Lock()
		msg := c.persist.Query()

		for _, ms := range msg {
			req, err := http.NewRequest(ms.Method, ms.Address, strings.NewReader(string(ms.Message)))
			if err != nil {
				klog.Errorf("cannot create new http request")
				continue
			}

			client := http.Client{}

			res, err := client.Do(req)
			if err != nil {
				klog.Error(err)
			}

			if res.StatusCode >= 200 && res.StatusCode < 300 {
				c.persist.Remove(msg)
			}
		}

		c.lock.Unlock()
		// exectue only every 10 minutes
		time.Sleep(10 * time.Minute)
	}
}

// NewConnection create a new connection
func NewConnection(baseURL string, token <-chan auth.Token, persist Persist) *Connection {
	u := Connection{baseURL: baseURL, tokenChan: token, persist: persist}
	go u.renewToken()
	return &u
}

// Request makes a request aggainst to the analyse cloud connection
func (c *Connection) Request(method, path string, queryArgs map[string]string, data io.Reader) (*http.Response, error) {
	klog.Infof("making http request against url %s with method %s", c.baseURL+"/"+path, method)
	c.lock.Lock()

	dataArray, err := ioutil.ReadAll(data)
	if err != nil {
		klog.Errorf("cannot read all data from io.Reader: %s", err)
		return nil, err
	}
	msg := []Message{{Address: c.baseURL + "/" + path, Message: dataArray, Method: method}}
	c.persist.Insert(msg)
	req, err := http.NewRequest(method, c.baseURL+"/"+path, strings.NewReader(string(dataArray)))
	if err != nil {
		return nil, err
	}

	for i, v := range queryArgs {
		req.URL.Query().Add(i, v)
	}

	req.Header.Add("token", c.token)
	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return res, err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		c.persist.Remove(msg)
	}
	c.lock.Unlock()

	return res, err
}
