package alibaba

import (
	"github.com/spf13/cobra"
	"github.com/teamssix/cf/pkg/cloud/alibaba/alirds"
)

var (
	rdslsFlushCache            bool
	rdsInfo					   bool
	rdslsRegion                string
	rdslsType                  string
	rdslsSpecifiedDBInstanceID string
	rdsConnect				   string
	rdsConnectCancel		   string
	rdsWhiteList               string
	rdsAccount				   string
)

func init() {
	alibabaCmd.AddCommand(rdsCmd)
	rdsCmd.AddCommand(rdslsCmd)
	rdsCmd.AddCommand(rdsExecCmd)

	rdsCmd.PersistentFlags().BoolVar(&rdslsFlushCache, "flushCache", false, "刷新缓存，不使用缓存数据 (Refresh the cache without using cached data)")
	rdslsCmd.Flags().StringVarP(&rdslsRegion, "region", "r", "all", "指定区域 ID (Specify Region ID)")
	rdslsCmd.Flags().StringVarP(&rdslsSpecifiedDBInstanceID, "DBInstanceID", "i", "all", "指定数据库实例 ID (Specify DBInstance ID)")
	rdslsCmd.Flags().StringVarP(&rdslsType, "type", "t", "all", "指定数据库类型 (Specify DBInstance Type)")
	rdsExecCmd.Flags().BoolVar(&rdsInfo, "ls", false, "列出数据库实例信息 (Lists database instance information)")
	rdsExecCmd.Flags().StringVarP(&rdsConnect, "conn", "c", "", "创建公网连接地址，参数中输入地址前缀，例如crossfire (Create a public network connection address and enter an address prefix in the parameter, such as crossfire)")
	rdsExecCmd.Flags().StringVarP(&rdsConnectCancel, "connCancel", "","", "关闭公网连接地址, 参数中输入地址前缀，例如crossfire (Disable the public IP address and  enter an address prefix in the parameter, such as crossfire)")
	rdsExecCmd.Flags().StringVarP(&rdsWhiteList, "white", "w", "", "追加数据库白名单地址，参数中输入白名单地址，例如127.0.0.1 (Append the whitelist address of the database such as 127.0.0.1")
	rdsExecCmd.Flags().StringVarP(&rdsAccount, "account", "a", "", "为实例添加帐号，参数中输入帐号名称，例如crossfire (To add an account for the instance, enter an account name in the parameter, such as crossfire)")
}

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "执行与云数据库相关的操作 (Perform rds-related operations)",
	Long:  "执行与云数据库相关的操作 (Perform rds-related operations)",
}

var rdslsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出所有的云数据库 (List all DBInstances)",
	Long:  "列出所有的云数据库 (List all DBInstances)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.PrintDBInstancesList(rdslsRegion, running, rdslsSpecifiedDBInstanceID, rdslsType, rdslsFlushCache)
	},
}

var rdsExecCmd = &cobra.Command{
	Use:	"exec",
	Short:  "执行与rds相关的操作 (Perform rds-related operations)",
	Long:	"执行与rds相关的操作 (Perform rds-related operations)",
	Run:	func(cmd *cobra.Command, args []string) {
		alirds.DBInstancesExec(rdslsRegion, running, rdslsSpecifiedDBInstanceID, rdslsType, rdslsFlushCache, rdsInfo, rdsConnect, rdsConnectCancel, rdsWhiteList, rdsAccount)
	},
}
