package config

type ServerConfig struct {
	Address string `mapstructure:"address" json:"address"`
	Port    int    `mapstructure:"port" json:"port"`
}
