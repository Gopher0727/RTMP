package config

type RedisConfig struct {
	Address  string `mapstructure:"address" json:"address"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
	PoolSize int    `mapstructure:"pool_size" json:"pool_size"`
}
