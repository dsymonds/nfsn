/*
Package nfsn contains abstractions for interacting with the NearlyFreeSpeech.NET API.

https://members.nearlyfreespeech.net/wiki/API/Reference
*/
package nfsn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Client struct {
	user, apiKey string
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
	}, nil
}
