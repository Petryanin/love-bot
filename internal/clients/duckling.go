package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ParseResponse struct {
	Body  string `json:"body"`
	Dim   string `json:"dim"`
	Value struct {
		Value any `json:"value"` // обычно строка ISO
	} `json:"value"`
}

type parseRequest struct {
	Text    string `json:"text"`
	Locale  string `json:"locale"`
	TZ      string `json:"tz"`
	Reftime int64  `json:"reftime"`
}

type DucklingParser interface {
	Parse(ctx context.Context, text string, ref time.Time, tz string) ([]ParseResponse, error)
}

type DucklingClient struct {
	api     Requester
	locale  string
	headers map[string]string
}

var _ DucklingParser = (*DucklingClient)(nil)

func NewDucklingClient(baseURL, locale string) *DucklingClient {
	return &DucklingClient{
		api:     NewBaseClient(baseURL),
		locale:  locale,
		headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	}
}

func (c *DucklingClient) Parse(ctx context.Context, text string, ref time.Time, tz string) ([]ParseResponse, error) {
	data := url.Values{}
	data.Set("text", text)
	data.Set("locale", c.locale)
	data.Set("tz", tz)
	data.Set("reftime", strconv.FormatInt(ref.UnixMilli(), 10))

	endpoint := fmt.Sprintf("%s/parse", c.api.BaseURL())

	log.Printf("requesting %q", endpoint)
	responseBody, err := c.api.DoRequest(
		ctx,
		"POST",
		endpoint,
		strings.NewReader(data.Encode()),
		c.headers,
	)
	if err != nil {
		log.Print("duckling: failed to get response: %w", err)
		return nil, err
	}

	var result []ParseResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Print("duckling: failed to parse json response: %w", err)
		return nil, err
	}

	return result, nil
}
