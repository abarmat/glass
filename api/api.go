package api

import (
	"encoding/json"
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

type UrlQueryParams interface {
	Map() map[string]string
}

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

// Client is a client for the Catalyst content API
type Client struct {
	endpoint string
}

// NewClient instantiates a new Client
func NewClient(endpoint string) *Client {
	return &Client{endpoint}
}

func (client *Client) buildURL(path string, query map[string]string) (*url.URL, error) {
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

func (client *Client) getJSON(path string, query map[string]string, data interface{}) error {
	body, err := client.get(path, query)
	if err == nil {
		json.Unmarshal(body, data)
	}
	return err
}

func (client *Client) get(path string, query map[string]string) ([]byte, error) {
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
func (client *Client) GetStatus() (*ServerStatus, error) {
	var data ServerStatus
	return &data, client.getJSON(getStatusURL, nil, &data)
}

// GetHistory gets the full node history
func (client *Client) GetHistory() (*HistoryResult, error) {
	return client.GetHistoryWithOpts(GetHistoryParams{})
}

// GetHistoryWithOpts gets the node history
func (client *Client) GetHistoryWithOpts(query UrlQueryParams) (*HistoryResult, error) {
	var data HistoryResult
	params := query.Map()
	return &data, client.getJSON(getHistoryURL, params, &data)
}

// GetSceneEntityByID gets a scene entity by ID
func (client *Client) GetSceneEntityByID(entityID string) (*SceneEntity, error) {
	var data []SceneEntity
	params := map[string]string{"id": entityID}
	return &data[0], client.getJSON(getSceneEntityURL, params, &data)
}

// GetSceneEntityByPointer gets a scene entity by pointer
func (client *Client) GetSceneEntityByPointer(pointer string) (*SceneEntity, error) {
	var data []SceneEntity
	params := map[string]string{"pointer": pointer}
	return &data[0], client.getJSON(getSceneEntityURL, params, &data)
}

// GetPointers gets a list of string pointers for an EntityType
func (client *Client) GetPointers(entityType string) (data []string, err error) {
	return data, client.getJSON(getPointersURL+"/"+entityType, nil, &data)
}

// GetContent gets the raw content stored in the server
func (client *Client) GetContent(hashID string) (data []byte, err error) {
	data, err = client.get(getContentURL+"/"+hashID, nil)
	return
}

// GetAudit gets the audit information for a particular entity
func (client *Client) GetAudit(entityType string, entityID string) (*AuditInfo, error) {
	var data AuditInfo
	return &data, client.getJSON(getAuditURL+"/"+entityType+"/"+entityID, nil, &data)
}
