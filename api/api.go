package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	getStatusURL      = "/status"
	getHistoryURL     = "/history"
	getSceneEntityURL = "/entities/scene"
	getPointersURL    = "/pointers"
	getContentURL     = "/contents"
	getAuditURL       = "/audit"
)

const (
	EntityTypeScene   = "scene"
	EntityTypeProfile = "profile"
)

// URLQueryParams is an interface for url queries
type URLQueryParams interface {
	Map() map[string]string
}

// GetHistoryParams get history endpoint params
type GetHistoryParams struct {
	From       int64
	To         int64
	ServerName string
	Offset     int
	Limit      int
}

// Map converts parameters into a hashmap representation
func (opts GetHistoryParams) Map() map[string]string {
	var params = make(map[string]string)

	if opts.From > 0 {
		params["from"] = strconv.FormatInt(opts.From, 10)
	}
	if opts.To > 0 {
		params["to"] = strconv.FormatInt(opts.To, 10)
	}
	if opts.ServerName != "" {
		params["serverName"] = opts.ServerName
	}
	if opts.Offset > 0 {
		params["offset"] = strconv.Itoa(opts.Offset)
	}
	if opts.Limit > 0 {
		params["limit"] = strconv.Itoa(opts.Limit)
	}

	return params
}

// CatalystClient is a client for the Catalyst content API
type CatalystClient struct {
	endpoint string
}

// NewClient instantiates a new CatalystClient
func NewClient(endpoint string) *CatalystClient {
	return &CatalystClient{endpoint}
}

func (client *CatalystClient) buildURL(path string, query map[string]string) (*url.URL, error) {
	reqURL, err := url.Parse(client.endpoint + path)
	if err != nil {
		return nil, err
	}

	qry := reqURL.Query()
	for arg, val := range query {
		qry.Set(arg, val)
	}
	reqURL.RawQuery = qry.Encode()

	return reqURL, nil
}

func (client *CatalystClient) getJSON(path string, query map[string]string, data interface{}) error {
	body, err := client.get(path, query)
	if err == nil {
		json.Unmarshal(body, data)
	}
	return err
}

func (client *CatalystClient) get(path string, query map[string]string) ([]byte, error) {
	reqURL, err := client.buildURL(path, query)
	if err != nil {
		return nil, err
	}

	// TODO: set user agent
	res, err := http.Get(reqURL.String())
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetStatus gets the server status
func (client *CatalystClient) GetStatus() (*ServerStatus, error) {
	var data ServerStatus
	return &data, client.getJSON(getStatusURL, nil, &data)
}

// GetHistory gets the full node history
func (client *CatalystClient) GetHistory() (*HistoryResult, error) {
	return client.GetHistoryWithOpts(GetHistoryParams{})
}

// GetHistoryWithOpts gets the node history
func (client *CatalystClient) GetHistoryWithOpts(query URLQueryParams) (*HistoryResult, error) {
	var data HistoryResult
	params := query.Map()
	return &data, client.getJSON(getHistoryURL, params, &data)
}

// GetSceneEntityByID gets a scene entity by ID
func (client *CatalystClient) GetSceneEntityByID(entityID string) (*SceneEntity, error) {
	var data []SceneEntity
	params := map[string]string{"id": entityID}
	err := client.getJSON(getSceneEntityURL, params, &data)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("Not found")
	}
	return &data[0], nil
}

// GetSceneEntityByPointer gets a scene entity by pointer
func (client *CatalystClient) GetSceneEntityByPointer(pointer string) (*SceneEntity, error) {
	var data []SceneEntity
	params := map[string]string{"pointer": pointer}
	err := client.getJSON(getSceneEntityURL, params, &data)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("Not found")
	}
	return &data[0], nil
}

// GetPointers gets a list of string pointers for an EntityType
func (client *CatalystClient) GetPointers(entityType string) (data []string, err error) {
	return data, client.getJSON(getPointersURL+"/"+entityType, nil, &data)
}

// GetContent gets the raw content stored in the server
func (client *CatalystClient) GetContent(hashID string) (data []byte, err error) {
	data, err = client.get(getContentURL+"/"+hashID, nil)
	return
}

// GetAudit gets the audit information for a particular entity
func (client *CatalystClient) GetAudit(entityType string, entityID string) (*AuditInfo, error) {
	var data AuditInfo
	return &data, client.getJSON(getAuditURL+"/"+entityType+"/"+entityID, nil, &data)
}
