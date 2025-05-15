package clients

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

type CurrentWeatherResponse struct {
	Name    string `json:"name"`
	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

type OpenWeatherMapClient struct {
	*BaseClient
	APIKey string
}

func NewOpenWeatherMapClient(baseURL, apiKey string) *OpenWeatherMapClient {
	return &OpenWeatherMapClient{
		BaseClient: NewBaseClient(baseURL),
		APIKey:     apiKey,
	}
}

func (c *OpenWeatherMapClient) CurrentWeather(city string) (*CurrentWeatherResponse, error) {
	endpoint := fmt.Sprintf("%s/weather", c.BaseURL)
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %q: %w", endpoint, err)
	}

	q := u.Query()
	q.Set("q", city)
	q.Set("appid", c.APIKey)
	q.Set("units", "metric")
	q.Set("lang", "ru")
	u.RawQuery = q.Encode()

	log.Printf("requesting %q", u.String())
	responseBody, err := c.DoRequest("GET", u.String(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request %q: %w", u.String(), err)
	}

	var data CurrentWeatherResponse
	if err := json.Unmarshal(responseBody, &data); err != nil {
		return nil, fmt.Errorf("failed to decode json %+v: %w", responseBody, err)
	}

	return &data, nil
}
