package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamssix/cf/pkg/util/cmdutil/identify"
)

func init() {
	RootCmd.AddCommand(identifyCmd)
}

var identifyCmd = &cobra.Command{
	Use:   "identify",
	Short: "判断 AccessKey 属于哪个云厂商 (Determine which cloud provider the AccessKey belongs to)",
	Long:  "判断 AccessKey 属于哪个云厂商 (Determine which cloud provider the AccessKey belongs to)",
	Run: func(cmd *cobra.Command, args []string) {
		identify.IdentifyAccessKey()
	},
}
