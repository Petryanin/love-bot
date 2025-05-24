package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	DefaultTZ *time.Location `mapstructure:"DEFAULT_TZ"`
	FontPath  string         `mapstructure:"FONT_PATH"`

	TgToken           string  `mapstructure:"TG_TOKEN"`
	TgDebug           bool    `mapstructure:"TG_DEBUG"`
	TgPartnersChatIDs []int64 `mapstructure:"TG_PARTNERS_CHAT_IDS"`

	DBPath string `mapstructure:"DB_PATH"`

	WeatherAPIKey  string `mapstructure:"WEATHER_API_KEY"`
	WeatherAPIURL  string `mapstructure:"WEATHER_API_URL"`
	WeatherAPICity string `mapstructure:"WEATHER_API_CITY"`
	WeatherAPILang string `mapstructure:"WEATHER_API_LANG"`

	CatAPIURL string `mapstructure:"CAT_API_URL"`

	DatingStartDate time.Time      `mapstructure:"DATING_START_DATE"`
	DatingStartTZ   *time.Location `mapstructure:"DATING_START_TZ"`

	DucklingAPIURL string `mapstructure:"DUCKLING_API_URL"`
	DucklingTZ     string `mapstructure:"DUCKLING_TZ"`
	DucklingLocale string `mapstructure:"DUCKLING_LOCALE"`

	MagicBallImagesPath string `mapstructure:"MAGIC_BALL_IMAGES_PATH"`

	GeoNamesAPIURL      string `mapstructure:"GEONAMES_API_URL"`
	GeoNamesAPIUsername string `mapstructure:"GEONAMES_API_USERNAME"`
	GeoNamesAPILang     string `mapstructure:"GEONAMES_API_LANG"`

	ComplimentsFilePath string `mapstructure:"COMPLIMENTS_FILE_PATH"`
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
		stringToInt64SliceHook,
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

func stringToInt64SliceHook(f reflect.Type, t reflect.Type, data any) (any, error) {
	if f.Kind() == reflect.String &&
		t.Kind() == reflect.Slice &&
		t.Elem().Kind() == reflect.Int64 {

		str := data.(string)
		if str == "" {
			return []int64{}, nil
		}

		parts := strings.Split(str, ",")
		out := make([]int64, 0, len(parts))

		for _, p := range parts {
			p = strings.TrimSpace(p)
			number, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid int: %q: %w", p, err)
			}

			out = append(out, number)
		}
		return out, nil
	}
	return data, nil
}
