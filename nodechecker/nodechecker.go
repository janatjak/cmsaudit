package nodechecker

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
	Node     string             `json:"node"`
	Yarn     string             `json:"yarn"`
	Next     string             `json:"next"`
	React    string             `json:"react"`
	ReactDom string             `json:"react-dom"`
	Packages map[string]Package `json:"packages"`
}

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	IsError string `json:"isError"`
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
