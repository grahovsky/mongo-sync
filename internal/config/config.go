package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type DbConn struct {
	URI     string        `mapstructure:"uri"`
	Name    string        `mapstructure:"name"`
	Timeout time.Duration `mapstructure:"timeout"`
	Auth    struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Source   string `mapstructure:"source"`
	} `mapstructure:"auth"`
}

type Config struct {
	Log struct {
		Level string `mapstructure:"level" env:"LOG_LEVEL"`
	} `mapstructure:"log"`
	DBSource      DbConn `mapstructure:"dbSource"`
	DBDestination DbConn `mapstructure:"dbDestination"`
	Common        struct {
		SyncDateField string        `mapstructure:"syncDateField" env:"SYNC_DATE_FIELD"`
		NumWorkers    int           `mapstructure:"numWorkers" env:"SYNC_NUM_WORKERS"`
		Timeout       time.Duration `mapstructure:"timeout" env:"SYNC_TIMEOUT"`
		From          string        `mapstructure:"from" env:"SYNC_FROM"`
	} `mapstructure:"common"`
}

type Args struct {
	ConfigPath string
	LogLevel   string
}

var Settings *Config

func Init() {
	Settings = defaultSettings()

	args := Args{}
	pflag.StringVarP(&args.ConfigPath, "config", "c", "./configs/config.yaml", "Path to configuration file")
	pflag.StringVarP(&args.LogLevel, "loglevel", "l", "INFO", "log level app")
	pflag.Parse()

	viper.AutomaticEnv()
	if _, err := os.Stat(args.ConfigPath); err == nil {
		viper.SetConfigFile(args.ConfigPath)
		if err := viper.ReadInConfig(); err != nil {
			slog.Debug("not found default config", err)
		}
	}

	if err := viper.Unmarshal(&Settings); err != nil {
		slog.Error(err.Error())
	}

	// env priority
	envPriority()

	// arg priority
	if argExist("--loglevel") || argExist("-l") {
		Settings.Log.Level = args.LogLevel
	}
}

func defaultSettings() *Config {
	return &Config{
		Log: struct {
			Level string "mapstructure:\"level\" env:\"LOG_LEVEL\""
		}{Level: "ERROR"},
	}
}

func envPriority() {
	if val := viper.GetString("LOG_LEVEL"); val != "" {
		Settings.Log.Level = val
	}
}

func argExist(arg string) bool {
	for _, s := range os.Args {
		if strings.Contains(s, arg) {
			return true
		}
	}

	return false
}
