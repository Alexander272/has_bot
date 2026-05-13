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
		API           APIConfig           `yaml:"api"`
		HomeAssistant HomeAssistantConfig `yaml:"homeassistant"`
		Bot           BotConfig           `yaml:"bot"`
	}

	APIConfig struct {
		Port string `yaml:"port" env:"API_PORT" env-default:"8080"`
	}

	HomeAssistantConfig struct {
		Url   string `yaml:"url" env:"HA_URL"`
		Token string `yaml:"token" env:"HA_TOKEN"`
	}

	BotConfig struct {
		Server string `yaml:"server" env:"MOST_SERVER"`
		Token  string `yaml:"token" env:"MOST_TOKEN"`
	}

)

func Init(path string) (*Config, error) {
	var conf Config

	if err := cleanenv.ReadConfig(path, &conf); err != nil {
		return nil, fmt.Errorf("failed to read config file. error: %w", err)
	}

	return &conf, nil
}


