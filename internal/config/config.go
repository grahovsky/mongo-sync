package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type DBConn struct {
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
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`
	DBSource      DBConn `mapstructure:"dbSource"`
	DBDestination DBConn `mapstructure:"dbDestination"`
	Common        struct {
		SyncDateField string        `mapstructure:"syncDateField"`
		NumWorkers    int           `mapstructure:"numWorkers"`
		Timeout       time.Duration `mapstructure:"timeout"`
		TimeDelta     time.Duration `mapstructure:"timeDelta"`
		From          string        `mapstructure:"from"`
	} `mapstructure:"common"`
}

type Args struct {
	ConfigPath string
	LogLevel   string
}

var Settings *Config

func Init() {
	parseArgs()

	viper.SetDefault("log.level", "ERROR")
	viper.SetDefault("config", "./configs/config.yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(viper.GetString("config"))

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("Failed to read config file", "error", err)
	}

	if err := viper.Unmarshal(&Settings); err != nil {
		slog.Error("Failed to unmarshal configuration", "error", err)
		os.Exit(1)
	}
}

func parseArgs() {
	pflag.String("config", "./configs/config.yaml", "Path to configuration file")
	pflag.String("log.level", "", "log level app")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}
