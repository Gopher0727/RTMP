package config

// ServerConfig 服务配置
type ServerConfig struct {
	Address string `mapstructure:"address" json:"address"`
	Port    int    `mapstructure:"port" json:"port"`
}
