package cmd

import (
	"os"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/server"
	"github.com/cy77cc/k8s-manage/version"
	"github.com/spf13/cobra"
)

var (
	startCMD = &cobra.Command{
		Use:     "start",
		Short:   "start k8s-manage server",
		Version: version.VERSION,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.MustNewConfig()
			os.Mkdir("log", 0644)
			logger.Init(logger.MustNewZapLogger())
			return server.Start()
		},
	}
)

func init() {
	var cfgFile string
	rootCMD.PersistentFlags().StringVar(&cfgFile, "config", "configs/config.yaml", "config file path")
	config.SetConfigFile(cfgFile)
	startCMD.Flags().BoolVar(&config.Debug, "debug", false, "enable debug mode")
	rootCMD.AddCommand(startCMD)
}
