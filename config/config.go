package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string       `mapstructure:"app_name" json:"app_name"`
	Env     string       `mapstructure:"env" json:"env"` // development | production
	Server  ServerConfig `mapstructure:"server" json:"server"`

	Redis struct {
		Session RedisConfig `mapstructure:"session"`
		Message RedisConfig `mapstructure:"message"`
	} `mapstructure:"redis"`

	JWT JWTConfig `mapstructure:"jwt" json:"jwt"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path) // 指定配置文件路径
	v.SetConfigType("toml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}

	// 解析配置
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
