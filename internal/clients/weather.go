package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

type WeatherInfo struct {
	City        string
	Description string
	Icon        string
	Temp        float64
	FeelsLike   float64
	Humidity    int
	WindSpeed   float64
}

type WeatherResponse struct {
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

type WeatherFetcher interface {
	Fetch(ctx context.Context, city string) (WeatherInfo, error)
}

type OpenWeatherMapClient struct {
	api    Requester
	apiKey string
}

var _ WeatherFetcher = (*OpenWeatherMapClient)(nil)

func NewOpenWeatherMapClient(baseURL, apiKey string) *OpenWeatherMapClient {
	return &OpenWeatherMapClient{
		api:    NewBaseClient(baseURL),
		apiKey: apiKey,
	}
}

func (c *OpenWeatherMapClient) Fetch(ctx context.Context, city string) (WeatherInfo, error) {
	endpoint := fmt.Sprintf("%s/weather", c.api.BaseURL())
	u, err := url.Parse(endpoint)
	if err != nil {
		return WeatherInfo{}, fmt.Errorf("failed to parse url %q: %w", endpoint, err)
	}

	q := u.Query()
	q.Set("q", city)
	q.Set("appid", c.apiKey)
	q.Set("units", "metric")
	q.Set("lang", "ru")
	u.RawQuery = q.Encode()

	log.Printf("requesting %q", u.String())
	responseBody, err := c.api.DoRequest(ctx, "GET", u.String(), nil, nil)
	if err != nil {
		return WeatherInfo{}, fmt.Errorf("failed to request %q: %w", u.String(), err)
	}

	var raw WeatherResponse
	if err := json.Unmarshal(responseBody, &raw); err != nil {
		return WeatherInfo{}, fmt.Errorf("failed to decode json %+v: %w", responseBody, err)
	}

	return WeatherInfo{
		City:        raw.Name,
		Description: raw.Weather[0].Description,
		Icon:        raw.Weather[0].Icon,
		Temp:        raw.Main.Temp,
		FeelsLike:   raw.Main.FeelsLike,
		Humidity:    raw.Main.Humidity,
		WindSpeed:   raw.Wind.Speed,
	}, nil
}
