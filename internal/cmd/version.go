package cmd

import (
	"fmt"

	"github.com/cy77cc/k8s-manage/version"
	"github.com/spf13/cobra"
)

var (
	versionCMD = &cobra.Command{
		Use:   "version",
		Short: "print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.VERSION)
		},
	}
)

func init() {
	rootCMD.AddCommand(versionCMD)
}
