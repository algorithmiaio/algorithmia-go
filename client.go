package algorithmia

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Client struct {
	ApiKey     string
	ApiAddress string
}

func NewClient(apiKey, apiAddress string) *Client {
	c := &Client{
		ApiKey:     apiKey,
		ApiAddress: apiAddress,
	}
	if apiAddress == "" {
		c.ApiAddress = "https://api.algorithmia.com"
	}
	return c
}

func (c *Client) Algo(ref string) (*Algorithm, error) {
	return NewAlgorithm(c, ref)
}

func (c *Client) File(dataUrl string) *DataFile {
	return NewDataFile(c, dataUrl)
}

func (c *Client) postJsonHelper(url string, input interface{}, params url.Values) (*http.Response, error) {
	headers := http.Header{}
	if c.ApiKey != "" {
		headers.Add("Authorization", c.ApiKey)
	}

	var (
		inputJson []byte
		err       error
	)
	if input == nil {
		headers.Add("Content-Type", "application/json")
		inputJson, err = json.Marshal(input)
		if err != nil {
			return nil, err
		}
	}

	switch inp := input.(type) {
	case string:
		headers.Add("Content-Type", "text/plain")
		inputJson = []byte(inp)
	case []byte:
		headers.Add("Content-Type", "application/octet-stream")
		inputJson = inp
	default:
		headers.Add("Content-Type", "application/json")
		inputJson, err = json.Marshal(input)
		if err != nil {
			return nil, err
		}
	}

	return Request{Url: c.ApiAddress + url, Data: inputJson, Headers: headers, Params: params}.Post()
}

func (c *Client) getHelper(url string, params url.Values) (*http.Response, error) {
	headers := http.Header{}
	if c.ApiKey != "" {
		headers.Add("Authorization", c.ApiKey)
	}

	return Request{Url: c.ApiAddress + url, Headers: headers, Params: params}.Get()
}

func (c *Client) headHelper(url string) (*http.Response, error) {
	headers := http.Header{}
	if c.ApiKey != "" {
		headers.Add("Authorization", c.ApiKey)
	}

	return Request{Url: c.ApiAddress + url, Headers: headers}.Head()
}

func (c *Client) putHelper(url string, data []byte) (*http.Response, error) {
	headers := http.Header{}
	if c.ApiKey != "" {
		headers.Add("Authorization", c.ApiKey)
	}

	return Request{Url: c.ApiAddress + url, Headers: headers, Data: data}.Put()
}

func (c *Client) deleteHelper(url string) (*http.Response, error) {
	headers := http.Header{}
	if c.ApiKey != "" {
		headers.Add("Authorization", c.ApiKey)
	}

	return Request{Url: c.ApiAddress + url, Headers: headers}.Delete()
}
