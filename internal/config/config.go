package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Environment   string              `yaml:"environment" env:"APP_ENV" env-default:"dev"`
		LogLevel      string              `yaml:"log_level" env-default:"info"`
		LogSource     bool                `yaml:"log_source" env-default:"false"`
		HomeAssistant HomeAssistantConfig `yaml:"homeassistant"`
		Bot           BotConfig           `yaml:"bot"`
	}

	HomeAssistantConfig struct {
		Url   string `yaml:"url" env:"HA_URL"`
		Token string `yaml:"token" env:"HA_TOKEN"`
	}

	BotConfig struct {
		Server  string          `yaml:"server" env:"MOST_SERVER"`
		Token   string          `yaml:"token" env:"MOST_TOKEN"`
		Channels []ChannelConfig `yaml:"channels"`
	}

	ChannelConfig struct {
		ChannelId string         `yaml:"channel_id"`
		Sensors   []SensorConfig `yaml:"sensors"`
	}

	SensorConfig struct {
		Name     string `yaml:"name"`
		EntityID string `yaml:"entity_id"`
		Room     string `yaml:"room"`
	}
)

func Init(path string) (*Config, error) {
	var conf Config

	if err := cleanenv.ReadConfig(path, &conf); err != nil {
		return nil, fmt.Errorf("failed to read config file. error: %w", err)
	}

	return &conf, nil
}
