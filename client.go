/*
Package nfsn contains abstractions for interacting with the NearlyFreeSpeech.NET API.

https://members.nearlyfreespeech.net/wiki/API/Reference
*/
package nfsn

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	user, apiKey string

	rngMu sync.Mutex
	rng   *rand.Rand
}

func NewClient() (*Client, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("finding user's home dir: %v", err)
	}
	configFile := filepath.Join(home, ".nfsn-api")
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %v", configFile, err)
	}
	var cfg struct {
		User string `json:"login"`
		Key  string `json:"api-key"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %v", configFile, err)
	}
	return &Client{
		user:   cfg.User,
		apiKey: cfg.Key,

		rng: rand.New(rand.NewSource(time.Now().Unix())),
	}, nil
}

const saltBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (c *Client) authHeader(reqURI string, body []byte) string {
	// Generate salt.
	// This is the only part of this that needs to be serialised.
	var salt [16]byte
	c.rngMu.Lock()
	for i := range salt {
		salt[i] = saltBytes[c.rng.Intn(len(saltBytes))]
	}
	c.rngMu.Unlock()

	// Header is "login;timestamp;salt;hash".
	// hash is SHA1("login;timestamp;salt;api-key;request-uri;body-hash")
	// and body-hash is SHA1(body).
	bodyHash := sha1.Sum(body)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	hashInput := fmt.Sprintf("%s;%s;%s;%s;%s;%02x", c.user, ts, salt, c.apiKey, reqURI, bodyHash)

	return fmt.Sprintf("%s;%s;%s;%02x", c.user, ts, salt, sha1.Sum([]byte(hashInput)))
}

func (c *Client) http(method, path string, body interface{}) ([]byte, error) {
	hdr := make(http.Header)

	var encBody []byte
	switch x := body.(type) {
	case nil:
	case []byte:
		encBody = x
	case url.Values:
		encBody = []byte(x.Encode())
		hdr.Set("Content-Type", "application/x-www-form-urlencoded")
	default:
		panic(fmt.Sprintf("invalid body type %T", x))
	}

	req := &http.Request{
		Method: method,
		URL: &url.URL{
			Scheme: "https",
			Host:   "api.nearlyfreespeech.net",
			Path:   path,
		},
		Header: hdr,
		Body:   ioutil.NopCloser(bytes.NewReader(encBody)),
	}
	req.Header.Set("X-NFSN-Authentication", c.authHeader(req.URL.Path, encBody))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", resp.Status, respBody)
	}
	return respBody, nil
}

type DNSRR struct {
	Name string // e.g. name of subdomain
	Type string // e.g. "A", "CNAME", "NS"
	Data string // IP for A, FQDN for CNAME, NS
	TTL  int    // seconds

	//Scope string // e.g. "system", "member"
}

type DNS struct {
	c      *Client
	domain string
}

func (c *Client) DNS(domain string) DNS {
	return DNS{c: c, domain: domain}
}

func (d DNS) RRs() ([]DNSRR, error) {
	body, err := d.c.http("POST", "/dns/"+url.PathEscape(d.domain)+"/listRRs", nil)
	if err != nil {
		return nil, err
	}
	var resp []DNSRR
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("bad JSON response: %v", err)
	}
	return resp, nil
}

func (d DNS) AddRR(rr DNSRR) error {
	args := url.Values{
		"name": []string{rr.Name},
		"type": []string{rr.Type},
		"data": []string{rr.Data},
	}
	if rr.TTL > 0 {
		args["ttl"] = []string{strconv.Itoa(rr.TTL)}
	}
	_, err := d.c.http("POST", "/dns/"+url.PathEscape(d.domain)+"/addRR", args)
	return err
}
