package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string `mapstructure:"app_name" json:"app_name"`
	Env     string `mapstructure:"env" json:"env"` // development | production

	Server ServerConfig `mapstructure:"server" json:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql" json:"mysql"`

	Redis struct {
		Session RedisConfig `mapstructure:"session"`
		Message RedisConfig `mapstructure:"message"`
	} `mapstructure:"redis"`

	Kafka struct {
		Brokers       []string `mapstructure:"brokers" json:"brokers"`
		Topics        []string `mapstructure:"topics" json:"topics"`
		ConsumerGroup string   `mapstructure:"consumer_group" json:"consumer_group"`
	} `mapstructure:"kafka"`

	JWT JWTConfig `mapstructure:"jwt" json:"jwt"`
}

var globalConfig *Config

func LoadConfig(path string) *Config {
	v := viper.New()

	v.SetConfigFile(path) // 指定配置文件路径
	v.SetConfigType("toml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("failed to read config file: %v", err))
	}

	config := &Config{}

	// 解析配置
	if err := v.Unmarshal(config); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}

	globalConfig = config
	return config
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	if globalConfig == nil {
		panic("config not loaded")
	}
	return globalConfig
}

// GetMySQLConfig 获取MySQL配置
func GetMySQLConfig() MySQLConfig {
	return GetConfig().MySQL
}

// GetServerConfig 获取服务器配置
func GetServerConfig() ServerConfig {
	return GetConfig().Server
}

// GetJWTConfig 获取JWT配置
func GetJWTConfig() JWTConfig {
	return GetConfig().JWT
}
