package clients

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ParseResponse struct {
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

type DucklingClient struct {
	*BaseClient
	Locale  string
	TZ      string
	Headers map[string]string
}

func NewDucklingClient(baseURL, locale, tz string) *DucklingClient {
	return &DucklingClient{
		BaseClient: NewBaseClient(baseURL),
		Locale:     locale,
		TZ:         tz,
		Headers:    map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	}
}

func (c *DucklingClient) Parse(text string, ref time.Time) ([]ParseResponse, error) {
	data := url.Values{}
	data.Set("text", text)
	data.Set("locale", c.Locale)
	data.Set("tz", c.TZ)
	data.Set("reftime", strconv.FormatInt(ref.UnixMilli(), 10))

	endpoint := fmt.Sprintf("%s/parse", c.BaseURL)

	log.Printf("requesting %q", endpoint)
	responseBody, err := c.DoRequest(
		"POST",
		endpoint,
		strings.NewReader(data.Encode()),
		c.Headers,
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
