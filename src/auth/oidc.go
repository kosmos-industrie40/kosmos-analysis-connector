package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"k8s.io/klog"
)

type oidcAuth struct {
	tokenChan    chan<- Token
	url          string
	path         string
	user         string
	password     string
	schema       string
	port         int
	refreshURL   string
	refreshQuery string
}

type oidcToken struct {
	Token string
	Valid time.Time
}

// AuthToken return the token as string
func (t oidcToken) AuthToken() string {
	return t.Token
}

// NewOidcAuth create an new authentication client with oidc
func NewOidcAuth(tokenChan chan<- Token, schema, baseURL, path string, port int, user, password string) Auth {
	return oidcAuth{
		tokenChan: tokenChan,
		url:       baseURL,
		path:      path,
		schema:    schema,
		user:      user,
		password:  password,
		port:      port,
	}
}

// Login perform the login with oidc
func (o oidcAuth) Login() error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s://%s:%d/%s", o.schema, o.url, o.port, o.path), nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("initial request was not successfull")
	}

	body, err := html.Parse(res.Body)
	if err != nil {
		return err
	}

	o.refreshQuery = res.Request.URL.RawQuery
	o.refreshURL = fmt.Sprintf("%s://%s%s", res.Request.URL.Scheme, res.Request.URL.Host, res.Request.URL.Path)

	targetAddress := o.findTargetInForm(body)
	if targetAddress == "" {
		return fmt.Errorf("no address found -> cannot login")
	}

	go o.renew(targetAddress, res.Cookies())

	return err
}

func (o oidcAuth) renew(address string, cookies []*http.Cookie) {
	token, err := o.getInitialToken(address, cookies)
	if err != nil {
		klog.Errorf("cannot receive token: %s", err)
	}
	klog.V(2).Infof("token: %v", token)

	for {

		o.tokenChan <- token
		now := time.Now()
		dura := token.Valid.Sub(now)
		time.Sleep(dura)

		klog.Infof("refreshURL: %s", o.refreshURL)
		token, err = o.getFollowUpToken(o.refreshURL, cookies)
		if err != nil {
			klog.Errorf("cannot receive follow up token: %s", err)
		}

		klog.V(2).Infof("token: %v", token)
	}
}

func (o oidcAuth) getFollowUpToken(targetAddress string, cookies []*http.Cookie) (oidcToken, error) {
	klog.Infof("target address: %s\n", targetAddress)
	req, err := http.NewRequest(http.MethodGet, targetAddress, nil)
	if err != nil {
		return oidcToken{}, err
	}

	klog.V(2).Infof("raw query: %s", o.refreshQuery)
	req.URL.RawQuery = o.refreshQuery

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	//req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	//req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:84.0) Gecko/20100101 Firefox/84.0")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return oidcToken{}, err
	}

	//bod, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	return oidcToken{}, err
	//}

	body, err := html.Parse(res.Body)
	if err != nil {
		return oidcToken{}, err
	}
	addre := o.findTargetInForm(body)
	return o.getInitialToken(addre, cookies)

}

func (o oidcAuth) getInitialToken(targetAddress string, cookies []*http.Cookie) (oidcToken, error) {
	data := url.Values{
		"username": []string{o.user},
		"password": []string{o.password},
	}

	req, err := http.NewRequest(http.MethodPost, targetAddress, strings.NewReader(data.Encode()))
	if err != nil {
		return oidcToken{}, err
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return oidcToken{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return oidcToken{}, err
	}

	bod, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return oidcToken{}, err
	}

	klog.Infof("StatusCode in get initial token is: %d", res.StatusCode)
	klog.V(2).Infoln(string(bod))

	var respJSON struct {
		Token string `json:"token"`
		Valid string `json:"valid"`
	}

	if err := json.Unmarshal(bod, &respJSON); err != nil {
		return oidcToken{}, err
	}

	date, err := time.Parse(time.RFC3339, respJSON.Valid)
	if err != nil {
		return oidcToken{}, err
	}

	tok := oidcToken{
		Token: respJSON.Token,
		Valid: date,
	}

	return tok, nil
}

// Logout perform the logout
func (o oidcAuth) Logout() error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s://%s/%s", o.schema, o.url, o.path), nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("deletion was not successfull")
}

func (o oidcAuth) findTargetInForm(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "form" {
		for _, attr := range node.Attr {
			if attr.Key == "action" {
				return attr.Val
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if addr := o.findTargetInForm(c); addr != "" {
			return addr
		}
	}
	return ""
}
