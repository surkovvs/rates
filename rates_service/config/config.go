package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type (
	AppCfg struct {
		DB
		GRPC
		Log
	}
	Log struct {
		Lvl int `mapstructure:"LOG_LVL"`
	}
	DB struct {
		Name          string `mapstructure:"DB_NAME"`
		User          string `mapstructure:"DB_USER"`
		Password      string `mapstructure:"DB_PASSWORD"`
		Host          string `mapstructure:"DB_HOST"`
		Port          string `mapstructure:"DB_PORT"`
		MigrationPath string `mapstructure:"DB_MIGR_PATH"`
	}
	GRPC struct {
		Port     string `mapstructure:"GRPC_PORT"`
		UserHost string `mapstructure:"GRPC_USER_HOST"`
		GwPort   string `mapstructure:"GRPC_GW_PORT"`
	}
)

func NewAppConfig(env string) (*AppCfg, error) {
	// Устанавливаем дефолтные значения
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("HOST", "localhost")
	viper.SetDefault("DATABASE", "mydb")

	// Мапим переменные из файла
	viper.SetConfigFile(env)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed read dotenv file: %w", err)
	}

	// Мапим переменные окружения
	viper.AutomaticEnv()

	// Устанавливаем флаги
	pflag.Int("port", 8080, "Port to run the application on")
	pflag.String("host", "localhost", "Host of the application")
	pflag.String("database", "mydb", "Database name")
	pflag.Parse()

	// Мапим флаги
	viper.BindPFlags(pflag.CommandLine)

	// Записываем в структуру
	var config = &AppCfg{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed env map decoding: %w", err)
	}
	return config, nil
}
