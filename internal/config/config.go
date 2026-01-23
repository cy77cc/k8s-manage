package config

import (
	"github.com/spf13/viper"
)

var CfgFile string

var CFG Config

var Debug bool

func Init() {
	// load config file
	viper.SetConfigFile(CfgFile)
	viper.ReadInConfig()

	// unmarshal config
	viper.Unmarshal(&CFG)
}
