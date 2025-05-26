package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Petryanin/love-bot/internal/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTodaySummary(t *testing.T) {
	city := "TestCity"

	t.Run("success", func(t *testing.T) {
		clientMock := clients.NewMockWeatherFetcher(t)
		clientMock.EXPECT().Fetch(mock.Anything, city).Return(
			clients.WeatherInfo{
				City:        "Тестоград",
				Description: "ясно",
				Temp:        20,
				FeelsLike:   19,
				Humidity:    40,
				WindSpeed:   3,
			}, nil,
		)
		svc := NewWeatherService(clientMock, city)

		summary, err := svc.TodaySummary(context.Background(), city)

		assert.NoError(t, err)
		assert.Contains(t, summary, "Тестоград")
		clientMock.AssertExpectations(t)
	})
	t.Run("error", func(t *testing.T) {
		clientMock := clients.NewMockWeatherFetcher(t)
		clientMock.EXPECT().Fetch(mock.Anything, city).Return(
			clients.WeatherInfo{}, errors.New("error"),
		)
		svc := NewWeatherService(clientMock, city)

		_, err := svc.TodaySummary(context.Background(), city)

		assert.Error(t, err)
		clientMock.AssertExpectations(t)
	})
}
