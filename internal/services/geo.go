package services

import (
	"context"

	"github.com/Petryanin/love-bot/internal/clients"
)

type GeoService struct {
	geo clients.GeoNamesSearcher
}

func NewGeoService(geo clients.GeoNamesSearcher) *GeoService {
	return &GeoService{geo: geo}
}

func (s *GeoService) ResolveByName(
	ctx context.Context,
	name string,
) (city, tz string, err error) {
	cityInfo, err := s.geo.SearchCity(ctx, name)
	if err != nil {
		return "", "", err
	}

	city = cityInfo.Name
	tz, err = s.geo.Timezone(ctx, cityInfo.Latitude, cityInfo.Longitude)
	if err != nil {
		return "", "", err
	}
	return city, tz, nil
}

func (s *GeoService) ResolveByCoords(
	ctx context.Context,
	lat, lng float64,
) (city, tz string, err error) {
	tz, err = s.geo.Timezone(ctx, lat, lng)
	if err != nil {
		return "", "", err
	}
	city, err = s.geo.ReverseGeocode(ctx, lat, lng)
	if err != nil {
		return "", "", err
	}
	return city, tz, nil
}
