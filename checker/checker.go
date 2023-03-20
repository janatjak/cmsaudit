package checker

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Checker struct {
	client *http.Client
}

type Audit struct {
	Php      string     `json:"php"`
	Packages []Packages `json:"packages"`
}

type Packages struct {
	Versions map[string]Package `json:"versions"`
}

type Package struct {
	Version string `json:"version"`
}

func (c *Checker) Check(url string) (*Audit, error) {
	response, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("client error")
	}

	defer response.Body.Close()

	var result Audit
	json.NewDecoder(response.Body).Decode(&result)
	return &result, nil
}

func New(timeout time.Duration) *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}
