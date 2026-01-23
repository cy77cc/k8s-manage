package cmd

import (
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
			config.Init()
			logger.Init(logger.NewZapLogger(config.CFG.Log.Level))
			return server.Start()
		},
	}
)

func init() {
	rootCMD.PersistentFlags().StringVar(&config.CfgFile, "config", "configs/config.yaml", "config file path")
	startCMD.Flags().BoolVar(&config.Debug, "debug", false, "enable debug mode")
	rootCMD.AddCommand(startCMD)
}
