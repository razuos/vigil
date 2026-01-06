package commands

import (
	"log/slog"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("No .env file found, reading from ENV/Flags")
	}
}
