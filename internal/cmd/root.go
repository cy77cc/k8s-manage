package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
			logger.Init(logger.MustNewZapLogger())
			ctx := cmd.Context()
			return server.Start(ctx)
		},
	}
)

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTRAP)
	defer stop()
	var cfgFile string
	rootCMD.PersistentFlags().StringVar(&cfgFile, "config", "configs/config.yaml", "config file path")
	config.SetConfigFile(cfgFile)
	rootCMD.Flags().BoolVar(&config.Debug, "debug", false, "enable debug mode")
	if err := rootCMD.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
