package cmd

import (
	"strings"
	"time"

	"github.com/rendau/dop/dopTools"
	"github.com/spf13/viper"
)

var conf = struct {
	Debug            bool    `mapstructure:"DEBUG"`
	LogLevel         string  `mapstructure:"LOG_LEVEL"`
	HttpListen       string  `mapstructure:"HTTP_LISTEN"`
	HttpCors         bool    `mapstructure:"HTTP_CORS"`
	SwagHost         string  `mapstructure:"SWAG_HOST"`
	SwagBasePath     string  `mapstructure:"SWAG_BASE_PATH"`
	SwagSchema       string  `mapstructure:"SWAG_SCHEMA"`
	DirPath          string  `mapstructure:"DIR_PATH"`
	CleanApiUrl      string  `mapstructure:"CLEAN_API_URL"`
	ImgMaxWidth      int     `mapstructure:"IMG_MAX_WIDTH"`
	ImgMaxHeight     int     `mapstructure:"IMG_MAX_HEIGHT"`
	WmPath           string  `mapstructure:"WM_PATH"`
	WmOpacity        float64 `mapstructure:"WM_OPACITY"`
	WmDirPaths       string  `mapstructure:"WM_DIR_PATHS"`
	WmDirPathsParsed []string
	CacheCount       int           `mapstructure:"CACHE_COUNT"`
	CacheDuration    time.Duration `mapstructure:"CACHE_DURATION"`
}{}

func confLoad() {
	dopTools.SetViperDefaultsFromObj(conf)

	viper.SetDefault("DEBUG", "false")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("HTTP_LISTEN", ":80")
	viper.SetDefault("SWAG_HOST", "example.com")
	viper.SetDefault("SWAG_BASE_PATH", "/")
	viper.SetDefault("SWAG_SCHEMA", "https")

	viper.SetConfigFile("conf.yml")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	_ = viper.Unmarshal(&conf)
}

func confParse() {
	conf.WmDirPathsParsed = confParseWMarkDirPaths(conf.WmDirPaths)
}

func confParseWMarkDirPaths(src string) []string {
	result := make([]string, 0)

	for _, p := range strings.Split(src, ";") {
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}
