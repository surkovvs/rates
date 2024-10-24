package config

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type (
	AppCfg struct {
		DB            *DB
		GRPC          *GRPC
		HTTP          *HTTP
		LogLvl        int8   `mapstructure:"log_lvl"`
		Market        string `mapstructure:"market"`
		GatherMetrics bool   `mapstructure:"metrics"`
	}
	DB struct {
		Name          string `mapstructure:"db_name"`
		User          string `mapstructure:"db_user"`
		Password      string `mapstructure:"db_password"`
		Host          string `mapstructure:"db_host"`
		Port          string `mapstructure:"db_port"`
		MigrationPath string `mapstructure:"db_migr_path"`
	}
	GRPC struct {
		Host string `mapstructure:"grpc_host"`
		Port string `mapstructure:"grpc_port"`
	}
	HTTP struct {
		Host string `mapstructure:"http_host"`
		Port string `mapstructure:"http_port"`
	}
)

func pflagAndViperStringReg(vi *viper.Viper, fs *pflag.FlagSet, envName, defValue string) {
	fs.String(envName, "", "")
	vi.SetDefault(envName, defValue)
}

func NewAppConfig() (AppCfg, error) {
	cfg := AppCfg{}
	vi := viper.New()
	fs := pflag.NewFlagSet("custom", pflag.ContinueOnError)
	// Регистрация переменных
	vi.SetDefault("log_lvl", 2)
	fs.Int8("log_lvl", 0, "")
	vi.SetDefault("metrics", false)
	fs.Bool("metrics", false, "")
	pflagAndViperStringReg(vi, fs, "market", "usdtusd")
	pflagAndViperStringReg(vi, fs, "db_name", "usdt")
	pflagAndViperStringReg(vi, fs, "db_user", "postgres")
	pflagAndViperStringReg(vi, fs, "db_password", "postgres")
	pflagAndViperStringReg(vi, fs, "db_host", "localhost")
	pflagAndViperStringReg(vi, fs, "db_port", "5432")
	pflagAndViperStringReg(vi, fs, "db_migr_path", "defaultname")
	pflagAndViperStringReg(vi, fs, "grpc_host", "localhost")
	pflagAndViperStringReg(vi, fs, "grpc_port", "9090")
	pflagAndViperStringReg(vi, fs, "http_host", "localhost")
	pflagAndViperStringReg(vi, fs, "http_port", "8080")
	// Мапинг переменных окружения
	vi.AutomaticEnv()
	// Мапинг переменных из файла (если путь задан флагом)
	fs.StringP("dotenvpath", "c", "", "Path to dotenv file if exists")
	if err := fs.Parse(os.Args); err != nil {
		return cfg, fmt.Errorf("flag set parse failed: %w", err)
	}
	dotenvFlag := fs.Lookup("dotenvpath")
	if dotenvFlag.Changed {
		wd, err := os.Getwd()                           // delete
		fmt.Println(wd, err, dotenvFlag.Value.String()) // delete
		vi.SetConfigFile(dotenvFlag.Value.String())
		if err := vi.ReadInConfig(); err != nil {
			return cfg, fmt.Errorf("read dotenv file failed: %w", err)
		}
	}
	// Мапинг флагов
	if err := vi.BindPFlags(fs); err != nil {
		return cfg, fmt.Errorf("binding flags in viper failed: %w", err)
	}
	// Запись в структуру
	cfg.DB = &DB{}
	cfg.GRPC = &GRPC{}
	cfg.HTTP = &HTTP{}
	if err := vi.Unmarshal(cfg.DB); err != nil {
		return cfg, fmt.Errorf("failed env map decoding: %w", err)
	}
	if err := vi.Unmarshal(cfg.GRPC); err != nil {
		return cfg, fmt.Errorf("failed env map decoding: %w", err)
	}
	if err := vi.Unmarshal(cfg.HTTP); err != nil {
		return cfg, fmt.Errorf("failed env map decoding: %w", err)
	}
	if err := vi.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed env map decoding: %w", err)
	}
	return cfg, nil
}
