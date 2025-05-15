package clients

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type BaseClient struct {
	Client  HTTPClient
	BaseURL string
}

func NewBaseClient(baseURL string) *BaseClient {
	return &BaseClient{
		Client:  &http.Client{Timeout: 5 * time.Second},
		BaseURL: baseURL,
	}
}

func (c *BaseClient) DoRequest(
	method, url string,
	body io.Reader,
	headers map[string]string,
) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to form request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)
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
