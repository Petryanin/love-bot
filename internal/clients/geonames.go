// internal/clients/geonames.go
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
)

type CityInfo struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type GeoNamesSearcher interface {
	SearchCity(ctx context.Context, name string) (*CityInfo, error)
	Timezone(ctx context.Context, lat, lng float64) (string, error)
	ReverseGeocode(ctx context.Context, lat, lng float64) (string, error)
}

type GeoNamesClient struct {
	api      Requester
	username string
	lang     string
}

var _ GeoNamesSearcher = (*GeoNamesClient)(nil)

func NewGeoNamesClient(baseURL, username, lang string) *GeoNamesClient {
	return &GeoNamesClient{
		api:      NewBaseClient(baseURL),
		username: username,
		lang:     lang,
	}
}

type geonamesSearchResponse struct {
	Geonames []struct {
		Name        string `json:"name"`
		Lat         string `json:"lat"`
		Lng         string `json:"lng"`
		CountryName string `json:"countryName"`
	} `json:"geonames"`
}

func (c *GeoNamesClient) SearchCity(ctx context.Context, name string) (*CityInfo, error) {
	u, err := url.Parse(c.api.BaseURL() + "/searchJSON")
	if err != nil {
		return nil, fmt.Errorf("geonames: invalid base URL %q: %w", c.api.BaseURL(), err)
	}
	q := u.Query()
	q.Set("q", name)
	q.Set("maxRows", "1")
	q.Set("username", c.username)
	q.Set("lang", c.lang)
	u.RawQuery = q.Encode()
	endpoint := u.String()

	log.Printf("geonames: requesting %q", endpoint)
	body, err := c.api.DoRequest(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("geonames: search request failed: %w", err)
	}

	var resp geonamesSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("geonames: invalid search JSON: %w", err)
	}
	if len(resp.Geonames) == 0 {
		return nil, fmt.Errorf("geonames: city %q not found", name)
	}
	g := resp.Geonames[0]
	lat, err := strconv.ParseFloat(g.Lat, 64)
	if err != nil {
		return nil, fmt.Errorf("geonames: invalid latitude %q: %w", g.Lat, err)
	}
	lng, err := strconv.ParseFloat(g.Lng, 64)
	if err != nil {
		return nil, fmt.Errorf("geonames: invalid longitude %q: %w", g.Lng, err)
	}

	return &CityInfo{
		Name:      g.Name,
		Latitude:  lat,
		Longitude: lng,
	}, nil
}

type geonamesTZResponse struct {
	TimezoneID string  `json:"timezoneId"`
	GMTOffset  float64 `json:"gmtOffset"`
}

func (c *GeoNamesClient) Timezone(ctx context.Context, lat, lng float64) (string, error) {
	u, err := url.Parse(c.api.BaseURL() + "/timezoneJSON")
	if err != nil {
		return "", fmt.Errorf("geonames: invalid base URL %q: %w", c.api.BaseURL(), err)
	}
	q := u.Query()
	q.Set("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	q.Set("lng", strconv.FormatFloat(lng, 'f', 6, 64))
	q.Set("username", c.username)
	q.Set("lang", c.lang)
	u.RawQuery = q.Encode()
	endpoint := u.String()

	log.Printf("geonames: requesting %q", endpoint)
	body, err := c.api.DoRequest(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return "", fmt.Errorf("geonames: geonames timezone request failed: %w", err)
	}

	var resp geonamesTZResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("geonames: invalid geonames timezone JSON: %w", err)
	}
	if resp.TimezoneID == "" {
		return "", fmt.Errorf("geonames: timezone not found for %f,%f", lat, lng)
	}

	return resp.TimezoneID, nil
}

type geonamesReverseResponse struct {
	Geonames []struct {
		Name string `json:"name"`
	} `json:"geonames"`
}

func (c *GeoNamesClient) ReverseGeocode(ctx context.Context, lat, lng float64) (string, error) {
	u, err := url.Parse(c.api.BaseURL() + "/findNearbyPlaceNameJSON")
	if err != nil {
		return "", fmt.Errorf("geonames: invalid base URL %q: %w", c.api.BaseURL(), err)
	}
	q := u.Query()
	q.Set("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	q.Set("lng", strconv.FormatFloat(lng, 'f', 6, 64))
	q.Set("username", c.username)
	q.Set("lang", c.lang)
	u.RawQuery = q.Encode()
	endpoint := u.String()

	log.Printf("geonames: requesting %q", endpoint)
	body, err := c.api.DoRequest(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return "", fmt.Errorf("geonames: reverse geocode request failed: %w", err)
	}

	var resp geonamesReverseResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("geonames: invalid reverse JSON: %w", err)
	}
	if len(resp.Geonames) == 0 {
		return "", fmt.Errorf("geonames: no place found near %f,%f", lat, lng)
	}
	g := resp.Geonames[0]

	return g.Name, nil
}
