package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

type ServerConfig struct {
	Port    string `mapstructure:"port"`
	Mode    string `mapstructure:"mode"`
	BaseURL string `mapstructure:"base_url"`
}

type DatabaseConfig struct {
	SQLitePath string `mapstructure:"sqlite_path"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置环境变量
	viper.AutomaticEnv()
	viper.SetEnvPrefix("WECHAT_QR")
	viper.BindEnv("server.base_url", "BASE_URL")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
