package alibaba

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/teamssix/cf/pkg/cloud/alibaba/alirds"
)

var (
	// rdsLs

	rdsLsFlushCache            bool
	rdsLsAllInfo               bool
	rdsLsRegion                string
	rdsLsType                  string
	rdsLsSpecifiedDBInstanceID string

	// rdsExec
	rdsAccountCancel   bool
	rdsConnectCancel   bool
	rdsWhiteListCancel bool
	rdsConnect         string
	rdsWhiteList       string
	rdsAccount         string
)

func init() {
	alibabaCmd.AddCommand(rdsCmd)
	rdsCmd.AddCommand(rdsLsCmd)
	rdsCmd.AddCommand(rdsExecCmd)

	// rdsCmd
	rdsCmd.PersistentFlags().BoolVar(&rdsLsFlushCache, "flushCache", false, "刷新缓存，不使用缓存数据 (Refresh the cache without using cached data)")

	// rdsLsCmd
	rdsLsCmd.Flags().StringVarP(&rdsLsRegion, "region", "r", "all", "指定区域 ID (Specify Region ID)")
	rdsLsCmd.Flags().StringVarP(&rdsLsSpecifiedDBInstanceID, "DBInstanceID", "i", "all", "指定数据库实例 ID (Specify DBInstance ID)")
	rdsLsCmd.Flags().StringVarP(&rdsLsType, "type", "t", "all", "指定数据库类型 (Specify DBInstance Type)")
	rdsLsCmd.Flags().BoolVarP(&rdsLsAllInfo, "all", "a", false, "列出更多数据库相关的信息 (List more information related to the database)")

	// rdsExecCmd

	rdsExecCmd.Flags().StringVarP(&rdsConnect, "conn", "c", "", "创建公网连接地址，参数中输入地址前缀，例如crossfire (Create a public network connection address and enter an address prefix in the parameter, such as crossfire)")
	rdsExecCmd.Flags().BoolVar(&rdsConnectCancel, "connCancel", false, "关闭通过cf创建的公网连接地址 (Disable the public IP address created through the cf)")
	rdsExecCmd.Flags().StringVarP(&rdsWhiteList, "white", "w", "", "追加数据库白名单地址，参数中输入白名单地址，例如127.0.0.1 (Append the whitelist address of the database such as 127.0.0.1")
	rdsExecCmd.Flags().BoolVar(&rdsWhiteListCancel, "whiteCancel", false, "删除通过cf追加的白名单地址 (Example Delete the whitelist addresses added to cf")
	rdsExecCmd.Flags().StringVarP(&rdsAccount, "account", "a", "", "为实例添加超管帐号，参数中输入帐号名称，例如crossfire (To add an administrator account for the instance, enter an account name in the parameter, such as crossfire)")
	rdsExecCmd.Flags().BoolVar(&rdsAccountCancel, "accountCancel", false, "删除通过cf添加的帐号 (Example Delete an account that is added through cf)")
}

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "执行与云数据库相关的操作 (Perform rds-related operations)",
	Long:  "执行与云数据库相关的操作 (Perform rds-related operations)",
}

var rdsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出所有的云数据库 (List all DBInstances)",
	Long:  "列出所有的云数据库 (List all DBInstances)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.PrintDBInstancesList(rdsLsRegion, running, rdsLsSpecifiedDBInstanceID, rdsLsType, rdsLsFlushCache, rdsLsAllInfo)
	},
}

var rdsExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "执行与rds相关的操作 (Perform rds-related operations)",
	Long:  "执行与rds相关的操作 (Perform rds-related operations)",
	Run: func(cmd *cobra.Command, args []string) {
		if rdsConnect == "" && rdsConnectCancel == false && rdsWhiteList == "" && rdsWhiteListCancel == false && rdsAccount == "" && rdsAccountCancel == false {
			log.Warnf("还未指定要执行的命令 (The command to be executed has not been specified yet)\n")
			cmd.Help()
		} else {
			alirds.DBInstancesExec(rdsLsRegion, running, rdsLsSpecifiedDBInstanceID, rdsLsType, rdsLsFlushCache, rdsConnect, rdsConnectCancel, rdsWhiteList, rdsWhiteListCancel, rdsAccount, rdsAccountCancel)
		}
	},
}
