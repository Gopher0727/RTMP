package config

// JWTConfig 鉴权配置
type JWTConfig struct {
	Secret           string `mapstructure:"secret" json:"secret"`
	Issuer           string `mapstructure:"issuer" json:"issuer"`
	TokenPrefix      string `mapstructure:"token_prefix" json:"token_prefix"`
	AccessExpMinutes int    `mapstructure:"access_exp_minutes" json:"access_exp_minutes"`
	RefreshExpHours  int    `mapstructure:"refresh_exp_hours" json:"refresh_exp_hours"`
}
