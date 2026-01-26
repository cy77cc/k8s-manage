package cmd

import (
	"fmt"
	"os"

	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/logger"
	"github.com/cy77cc/k8s-manage/internal/server"
	"github.com/cy77cc/k8s-manage/version"
	"github.com/spf13/cobra"
)

var (
	rootCMD = &cobra.Command{
		Use:     "k8s-manage",
		Short:   "k8s-manage is a tool to manage k8s cluster",
		Version: version.VERSION,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.MustNewConfig()
			fmt.Println(config.CFG)
			logger.Init(logger.MustNewZapLogger())
			return server.Start()
		},
	}
)

func Execute() {
	var cfgFile string
	rootCMD.PersistentFlags().StringVar(&cfgFile, "config", "configs/config.yaml", "config file path")
	config.SetConfigFile(cfgFile)
	rootCMD.Flags().BoolVar(&config.Debug, "debug", false, "enable debug mode")
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
}
