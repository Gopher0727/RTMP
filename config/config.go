package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string       `mapstructure:"app_name" json:"app_name"`
	Env     string       `mapstructure:"env" json:"env"` // development | production
	Server  ServerConfig `mapstructure:"server" json:"server"`
	MySQL   MySQLConfig  `mapstructure:"mysql" json:"mysql"`
	Redis   RedisConfig  `mapstructure:"redis" json:"redis"`
	JWT     JWTConfig    `mapstructure:"jwt" json:"jwt"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Address string `mapstructure:"address" json:"address"`
	Port    int    `mapstructure:"port" json:"port"`
}

// MySQLConfig 数据库配置
type MySQLConfig struct {
	UserName string `mapstructure:"user_name" json:"user_name"`
	Password string `mapstructure:"password" json:"password"`
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
}

// RedisConfig 缓存配置
type RedisConfig struct {
	Address  []string `mapstructure:"address" json:"address"` // 支持单机和集群
	Password string   `mapstructure:"password" json:"password"`
}

// JWTConfig 鉴权配置
type JWTConfig struct {
	Secret           string `mapstructure:"secret" json:"secret"`
	Issuer           string `mapstructure:"issuer" json:"issuer"`
	TokenPrefix      string `mapstructure:"token_prefix" json:"token_prefix"`
	AccessExpMinutes int    `mapstructure:"access_exp_minutes" json:"access_exp_minutes"`
	RefreshExpHours  int    `mapstructure:"refresh_exp_hours" json:"refresh_exp_hours"`
}

func LoadConfig(path string) (config *Config, err error) {
	v := viper.New()

	v.SetConfigFile(path) // 指定配置文件路径
	v.SetConfigType("toml")

	// 读取配置文件
	err = v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	err = v.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return
}
