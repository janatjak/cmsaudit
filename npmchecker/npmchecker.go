package npmchecker

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Checker struct {
	client *http.Client
}

type Dist struct {
	LatestVersion string `json:"latest"`
}

type Package struct {
	Name string `json:"name"`
	Dist Dist   `json:"dist-tags"`
}

func (c *Checker) Check(name string) (*Package, error) {
	response, err := c.client.Get("https://registry.npmjs.org/" + name)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("client error")
	}

	defer response.Body.Close()

	var result Package
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
