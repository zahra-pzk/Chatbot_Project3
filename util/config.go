package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	AIBaseURL           string        `mapstructure:"AIBaseURL"`
	AIAPIKey            string        `mapstructure:"AI_API_KEY"`
	APIURL              string        `mapstructure:"API_URL"`
	WS_URL              string        `mapstructure:"WS_URL"`
	BotUsername         string        `mapstructure:"BOT_USERNAME"`
	BotPass             string        `mapstructure:"BOT_PASSWORD"`
	ChunkSize           int64         `mapstructure:"ChunkSize"`
	ChunkOverlap        int64         `mapstructure:"ChunkOverlap"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
