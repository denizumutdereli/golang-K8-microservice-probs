package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	BaseURL string
}

type HTTPResponse struct {
	Body       []byte
	StatusCode int
}

func NewRestClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) DoRequest(method, path string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	jsonBody, err := json.Marshal(body)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &HTTPResponse{Body: respBody, StatusCode: resp.StatusCode}, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	return &HTTPResponse{Body: respBody, StatusCode: resp.StatusCode}, nil
}

func (c *Client) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
}
