package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Requester interface {
	DoRequest(
		ctx context.Context,
		method, url string,
		body io.Reader,
		headers map[string]string,
	) ([]byte, error)
	BaseURL() string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type BaseClient struct {
	client  HTTPClient
	baseURL string
}

var _ Requester = (*BaseClient)(nil)

func NewBaseClient(baseURL string) *BaseClient {
	return &BaseClient{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: baseURL,
	}
}

func (c *BaseClient) DoRequest(
	ctx context.Context,
	method, url string,
	body io.Reader,
	headers map[string]string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to form request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return responseBody, nil
}

func (c *BaseClient) BaseURL() string {
	return c.baseURL
}
