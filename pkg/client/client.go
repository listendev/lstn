package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	defaultBaseURL = "https://core.listen.dev/api/v1"
)

type APIClient struct {
	BaseURL    string
	jwtToken   string
	userAgent  string
	httpClient *http.Client
}

func NewAPIClient(jwtToken string) *APIClient {
	return &APIClient{
		BaseURL:   defaultBaseURL,
		jwtToken:  jwtToken,
		userAgent: "listen-go-sdk",
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
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) WithBaseURL(baseURL string) *APIClient {
	c.BaseURL = baseURL
	return c
}

func (c *APIClient) WithUserAgent(userAgent string) *APIClient {
	c.userAgent = userAgent
	return c
}

func (c *APIClient) WithHTTPClient(httpClient *http.Client) *APIClient {
	c.httpClient = httpClient
	return c
}

func (c *APIClient) postRequest(ctx context.Context, endpoint string, data interface{}) (*http.Response, error) {
	url, err := c.buildRequestURL(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error building request URL: %v", err)
	}
	body := &bytes.Buffer{}
	marshaler := json.NewEncoder(body)
	if err := marshaler.Encode(data); err != nil {
		return nil, fmt.Errorf("error marshaling data: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.jwtToken))
	req.Header.Add("User-Agent", c.userAgent)

	return c.httpClient.Do(req)
}

func (c *APIClient) PostEvent(ctx context.Context, event Event) error {

	res, err := c.postRequest(ctx, "events", event)
	if err != nil {
		return fmt.Errorf("error creating event: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("server error creating event: %v", res.Status)
	}
	return nil
}

func (c *APIClient) buildRequestURL(endpoint string) (string, error) {
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	return baseURL.ResolveReference(&url.URL{Path: path.Join(baseURL.Path, "events")}).String(), nil
}
