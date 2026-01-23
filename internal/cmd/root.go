package cmd

import (
	"os"

	"github.com/cy77cc/k8s-manage/version"
	"github.com/spf13/cobra"
)

var (
	rootCMD = &cobra.Command{
		Use:     "k8s-manage",
		Short:   "k8s-manage is a tool to manage k8s cluster",
		Version: version.VERSION,
	}
)

func Execute() {
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
}
