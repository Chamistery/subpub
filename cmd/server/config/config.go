package config

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

func Load() (*Config, error) {
	viper.SetConfigName("dev")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
