package apichecker

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
	Server   string
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
		return &Audit{
			Server: response.Header.Get("server"),
			Packages: []Packages{
				{Versions: map[string]Package{}},
			},
		}, errors.New("client error")
	}

	defer response.Body.Close()

	var result Audit
	json.NewDecoder(response.Body).Decode(&result)

	result.Server = response.Header.Get("server")

	return &result, nil
}

func New(timeout time.Duration) *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}
