package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pborman/uuid"
)

const (
	StorageServicePath = "read/org.apache.cassandra.db:type=StorageService"
)

type StorageService struct {
	HostIdMap        map[string]uuid.UUID
	LiveNodes        []string
	UnreachableNodes []string
	LeavingNodes     []string
	JoiningNodes     []string
	MovingNodes      []string
	LocalHostId      uuid.UUID
}

type Interface interface {
	StorageService() (*StorageService, error)
}

type client struct {
	StorageServiceURL *url.URL
	client            *http.Client
}

var _ Interface = &client{}

func New(baseURL *url.URL, c *http.Client) *client {
	storageServiceURL, err := url.Parse(StorageServicePath)
	if err != nil {
		panic(err)
	}
	storageServiceURL = baseURL.ResolveReference(storageServiceURL)
	return &client{
		StorageServiceURL: storageServiceURL,
		client:            c,
	}
}

type JolokiaResponse struct {
	Value *StorageService `json:"value"`
}

func (c *client) StorageService() (*StorageService, error) {
	req, err := http.NewRequest(http.MethodGet, c.StorageServiceURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to instantiate HTTP request. '%s'", err)
	}
	req.Header.Set("User-Agent", "navigator-cassandra-nodetool-client")

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send HTTP request. '%s'", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Unexpected server response code. Expected %v. Got %v.",
			http.StatusOK,
			response.StatusCode,
		)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body. '%s'", err)
	}

	out := &JolokiaResponse{}

	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal JSON response. '%s'", err)
	}
	if out.Value == nil {
		return nil, fmt.Errorf("The response had an empty Jolokia value. '%#v'", out)
	}
	return out.Value, nil
}
