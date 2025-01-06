/*
Copyright 2024 The CloudPilot AI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	Endpoint = "http://100.100.100.200"
	regionID = "region-id"
)

// MetaData wrap http client
type MetaData struct {
	// mock for unit test.
	mock   requestMock
	client *http.Client
}

// NewMetaData returns MetaData
func NewMetaData(client *http.Client) *MetaData {
	if client == nil {
		client = &http.Client{}
	}
	return &MetaData{
		client: client,
	}
}

// New returns MetaDataRequest
func (m *MetaData) New() *MetaDataRequest {
	return &MetaDataRequest{
		client:      m.client,
		sendRequest: m.mock,
	}
}

// Region returns region
func (m *MetaData) Region() (string, error) {
	var region ResultList
	err := m.New().Resource(regionID).Do(&region)
	if err != nil {
		return "", err
	}
	return region.result[0], nil
}

type requestMock func(resource string) (string, error)

// ResultList struct
type ResultList struct {
	result []string
}

// nolint: stylecheck
// RoleAuth struct
type RoleAuth struct {
	AccessKeyId     string
	AccessKeySecret string
	Expiration      time.Time
	SecurityToken   string
	LastUpdated     time.Time
	Code            string
}

// MetaDataRequest struct
type MetaDataRequest struct {
	version      string
	resourceType string
	resource     string
	subResource  string
	client       *http.Client

	sendRequest requestMock
}

// Version sets version
func (r *MetaDataRequest) Version(version string) *MetaDataRequest {
	r.version = version
	return r
}

// ResourceType sets resource type
func (r *MetaDataRequest) ResourceType(rtype string) *MetaDataRequest {
	r.resourceType = rtype
	return r
}

// Resource sets resource
func (r *MetaDataRequest) Resource(resource string) *MetaDataRequest {
	r.resource = resource
	return r
}

// SubResource set sub resource
func (r *MetaDataRequest) SubResource(sub string) *MetaDataRequest {
	r.subResource = sub
	return r
}

// URL returns url
func (r *MetaDataRequest) URL() (string, error) {
	if r.version == "" {
		r.version = "latest"
	}
	if r.resourceType == "" {
		r.resourceType = "meta-data"
	}
	if r.resource == "" {
		return "", errors.New("the resource you want to visit must not be nil")
	}
	endpoint := os.Getenv("METADATA_ENDPOINT")
	if endpoint == "" {
		endpoint = Endpoint
	}
	url := fmt.Sprintf("%s/%s/%s/%s", endpoint, r.version, r.resourceType, r.resource)
	if r.subResource == "" {
		return url, nil
	}
	return fmt.Sprintf("%s/%s", url, r.subResource), nil
}

// Do try to do MetaDataRequest
func (r *MetaDataRequest) Do(api interface{}) (err error) {
	res := ""

	if r.sendRequest != nil {
		res, err = r.sendRequest(r.resource)
	} else {
		res, err = r.send()
	}

	if err != nil {
		return err
	}
	return r.Decode(res, api)
}

// Decode returns decoded content
func (r *MetaDataRequest) Decode(data string, api interface{}) error {
	if data == "" {
		url, _ := r.URL()
		return fmt.Errorf("metadata: alivpc decode data must not be nil. url=[%s]", url)
	}
	switch api := api.(type) {
	case *ResultList:
		api.result = strings.Split(data, "\n")
		return nil
	case *RoleAuth:
		return json.Unmarshal([]byte(data), api)
	default:
		return fmt.Errorf("metadata: unknow type to decode, type=%s", reflect.TypeOf(api))
	}
}

func (r *MetaDataRequest) send() (string, error) {
	url, err := r.URL()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return "", err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("aliyun Metadata API Error: Status Code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
