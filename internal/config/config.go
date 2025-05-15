package config

import (
	"fmt"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	DefaultTZ *time.Location `mapstructure:"DEFAULT_TZ"`
	FontPath  string         `mapstructure:"FONT_PATH"`

	TgToken         string `mapstructure:"TG_TOKEN"`
	TgDebug         bool   `mapstructure:"TG_DEBUG"`
	TgPartnerCharID int64  `mapstructure:"TG_PARTNER_CHAT_ID"`

	DBPath string `mapstructure:"DB_PATH"`

	WeatherAPIKey  string `mapstructure:"WEATHER_API_KEY"`
	WeatherAPIURL  string `mapstructure:"WEATHER_API_URL"`
	WeatherAPICity string `mapstructure:"WEATHER_API_CITY"`

	CatAPIURL string `mapstructure:"CAT_API_URL"`

	DatingStartDate time.Time      `mapstructure:"DATING_START_DATE"`
	DatingStartTZ   *time.Location `mapstructure:"DATING_START_TZ"`

	DucklingAPIURL string `mapstructure:"DUCKLING_API_URL"`
	DucklingTZ     string `mapstructure:"DUCKLING_TZ"`
	DucklingLocale string `mapstructure:"DUCKLING_LOCALE"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	decodeHook := mapstructure.ComposeDecodeHookFunc(
		stringToTimeHook,
		stringToLocationHook,
	)

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(decodeHook)); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &cfg, nil
}

func stringToTimeHook(f reflect.Type, t reflect.Type, data any) (any, error) {
	if f.Kind() == reflect.String && t == reflect.TypeOf(time.Time{}) {
		s := data.(string)
		if s == "" {
			return time.Time{}, nil
		}
		return time.Parse("2006-01-02", s)
	}
	return data, nil
}

func stringToLocationHook(f reflect.Type, t reflect.Type, data any) (any, error) {
	if f.Kind() == reflect.String && t == reflect.TypeOf(&time.Location{}) {
		name := data.(string)
		if name == "" {
			return time.Local, nil
		}

		loc, err := time.LoadLocation(name)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", name, err)
		}
		return loc, nil
	}
	return data, nil
}
