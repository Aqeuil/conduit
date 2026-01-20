package config

import (
	"flag"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	Config        string = "conf"
	ServerPort    string = "server.port"
	ServerNetwork string = "server.network"
)

var (
	configPath = flag.String(Config, "./configs/local.yml", "config path")
	_          = flag.String(ServerPort, "8080", "listen port")
	_          = flag.String(ServerNetwork, "tcp", "listen network")
)

func Parse() error {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}

	viper.SetConfigFile(*configPath)
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}
