package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://core.listen.dev/api/v1"
)

type APIClient struct {
	BaseURL    string
	jwtToken   string
	httpClient *http.Client
}

func NewAPIClient(jwtToken string) *APIClient {
	return &APIClient{
		BaseURL:  defaultBaseURL,
		jwtToken: jwtToken,
		httpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				MaxConnsPerHost:     100,
				MaxIdleConnsPerHost: 100,
			},
			Timeout: 1 * time.Second,
		},
	}
}

func (c *APIClient) WithBaseURL(baseURL string) *APIClient {
	c.BaseURL = baseURL
	return c
}

func (c *APIClient) WithHTTPClient(httpClient *http.Client) *APIClient {
	c.httpClient = httpClient
	return c
}

func (c *APIClient) postRequest(endpoint string, data interface{}) (*http.Response, error) {
	url := c.BaseURL + endpoint
	body := &bytes.Buffer{}
	marshaler := json.NewEncoder(body)
	if err := marshaler.Encode(data); err != nil {
		return nil, fmt.Errorf("error marshaling data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.jwtToken))

	return c.httpClient.Do(req)
}

func (c *APIClient) PostEvent(event Event) error {
	url := c.BaseURL + "/events"

	res, err := c.postRequest(url, event)
	if err != nil {
		return fmt.Errorf("error creating event: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("server error creating event: %v", res.Status)
	}
	return nil
}
