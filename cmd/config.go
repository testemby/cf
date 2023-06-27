package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/teamssix/cf/pkg/util/cmdutil"
)

var selectAll bool
var queryAccessKeyId string
var querySecretAccessKey string
var querySessionToken string

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(ConfigDel)
	configCmd.AddCommand(ConfigLs)
	configCmd.AddCommand(ConfigMf)
	configCmd.AddCommand(ConfigSw)
	configCmd.AddCommand(ConfigScan)
	configCmd.AddCommand(ConfigQuery)

	ConfigLs.PersistentFlags().BoolVarP(&selectAll, "all", "a", false, "查询全部数据 (Search all data)")
	ConfigScan.PersistentFlags().BoolVarP(&selectAll, "all", "a", false, "查询全部数据 (Search all data)")
	ConfigQuery.Flags().StringVarP(&queryAccessKeyId, "AccessKeyId", "a", "", "输入要查询的访问凭证 ID (Enter the Access Key ID you want to query)")
	ConfigQuery.Flags().StringVarP(&querySecretAccessKey, "SecretAccessKey", "s", "", "输入要查询的访问凭证密钥 (Enter the Secret Access Key you want to query)")
	ConfigQuery.Flags().StringVarP(&querySessionToken, "SessionToken", "t", "", "输入要查询的访问凭证会话令牌 (Enter the Session Token you want to query)")
	ConfigQuery.MarkFlagRequired("AccessKeyId")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置云服务商的访问密钥 (Configure cloud provider access key)",
	Long:  `配置云服务商的访问密钥 (Configure cloud provider access key)`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutil.ConfigureAccessKey()
	},
}

var ConfigDel = &cobra.Command{
	Use:   "del",
	Short: "删除访问密钥 (Delete access key)",
	Long:  `删除访问密钥 (Delete access key)`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutil.ConfigDel()
	},
}

var ConfigLs = &cobra.Command{
	Use:   "ls",
	Short: "列出已配置过的访问密钥 (List configured access key)",
	Long:  `列出已配置过的访问密钥 (List configured access key)`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutil.ConfigLs(selectAll)
	},
}

var ConfigMf = &cobra.Command{
	Use:   "mf",
	Short: "修改已配置过的访问密钥 (Modify configured access key)",
	Long:  `修改已配置过的访问密钥 (Modify configured access key)`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutil.ConfigMf()
	},
}

var ConfigSw = &cobra.Command{
	Use:   "sw",
	Short: "切换访问密钥 (Switch access key)",
	Long:  `切换访问密钥 (Switch access key)`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutil.ConfigSw()
	},
}

var ConfigScan = &cobra.Command{
	Use:   "scan",
	Short: "扫描本地访问密钥 (Scan for local access keys)",
	Long:  `扫描本地访问密钥 (Scan for local access keys)`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutil.ScanAccessKey(selectAll)
	},
}

var ConfigQuery = &cobra.Command{
	Use:   "query",
	Short: "查询访问凭证所属厂商 (Querying the provider to which the Access Key belongs)",
	Long:  `查询访问凭证所属厂商 (Querying the provider to which the Access Key belongs)`,
	Run: func(cmd *cobra.Command, args []string) {
		provider := cmdutil.IdentifyProvider(queryAccessKeyId, querySecretAccessKey, querySessionToken)
		fmt.Println()
		if provider.EN == "" {
			log.Infoln("暂无法判断该访问凭证所归属的云厂商 (Unable to determine the cloud provider associated with the given access key.)")
		} else {
			log.Infof("当前访问凭证可能属于%s (The current access key may belong to %s.)", provider.CN, provider.EN)
		}
	},
}
