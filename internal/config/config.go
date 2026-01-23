package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

var cfgFile string

var CFG Config

var Debug bool

func SetConfigFile(path string) {
	cfgFile = path
}

func MustNewConfig() {
	// load config file
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	replace := strings.NewReplacer(".", "_") // 替换点为下划线
	viper.SetEnvKeyReplacer(replace)         // 设置环境变量的替换器
	viper.AutomaticEnv()

	// unmarshal config
	if err := viper.Unmarshal(&CFG); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}
