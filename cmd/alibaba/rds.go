package alibaba

import (
	"github.com/spf13/cobra"
	"github.com/teamssix/cf/pkg/cloud/alibaba/alirds"
)

var (
	// rdsLs
	rdsLsFlushCache            bool
	rdsLsAllInfo               bool
	rdsLsRegion                string
	rdsLsType                  string
	rdsLsSpecifiedDBInstanceId string

	// rdsAccount
	rdsAccountSpecifiedDBInstanceId string
	rdsAccountUserName              string

	// rdsPublic
	rdsPublicSpecifiedDBInstanceId string

	// rdsWhiteList
	rdsWhiteListSpecifiedDBInstanceId string
	rdsWhiteList                      string
)

func init() {
	alibabaCmd.AddCommand(rdsCmd)
	rdsCmd.AddCommand(rdsLsCmd)
	rdsCmd.AddCommand(rdsAccountCmd)
	rdsCmd.AddCommand(rdsPublicCmd)
	rdsCmd.AddCommand(rdsWhiteListCmd)
	rdsAccountCmd.AddCommand(rdsAccountDelCmd)
	rdsAccountCmd.AddCommand(rdsAccountLsCmd)
	rdsPublicCmd.AddCommand(rdsPublicLsCmd)
	rdsPublicCmd.AddCommand(rdsPublicCancelCmd)
	rdsWhiteListCmd.AddCommand(rdsWhiteListLsCmd)
	rdsWhiteListCmd.AddCommand(rdsWhiteListDelCmd)

	// rdsCmd
	rdsCmd.PersistentFlags().BoolVar(&rdsLsFlushCache, "flushCache", false, "刷新缓存，不使用缓存数据 (Refresh the cache without using cached data)")

	// rdsLsCmd
	rdsLsCmd.Flags().StringVarP(&rdsLsRegion, "region", "r", "all", "指定区域 ID (Specify Region ID)")
	rdsLsCmd.Flags().StringVarP(&rdsLsSpecifiedDBInstanceId, "DBInstanceId", "i", "all", "指定数据库实例 ID (Specify DBInstance ID)")
	rdsLsCmd.Flags().StringVarP(&rdsLsType, "type", "t", "all", "指定数据库类型 (Specify DBInstance Type)")
	rdsLsCmd.Flags().BoolVarP(&rdsLsAllInfo, "all", "a", false, "列出更多数据库相关的信息 (List more information related to the database)")

	// rdsAccount
	rdsAccountCmd.Flags().StringVarP(&rdsAccountSpecifiedDBInstanceId, "DBInstanceId", "i", "all", "指定 RDS 实例 ID (Specify the RDS instance ID)")
	rdsAccountCmd.Flags().StringVarP(&rdsAccountUserName, "userName", "u", "crossfire", "指定用户名 (Specify user name)")

	// rdsPublic
	rdsPublicCmd.Flags().StringVarP(&rdsPublicSpecifiedDBInstanceId, "DBInstanceId", "i", "all", "指定 RDS 实例 ID (Specify the RDS instance ID)")

	// rdsWhiteListCmd
	rdsWhiteListCmd.Flags().StringVarP(&rdsWhiteList, "WhiteList", "w", "", "指定要添加的白名单地址 (Specify the whitelist address to be added)")
	rdsWhiteListCmd.Flags().StringVarP(&rdsWhiteListSpecifiedDBInstanceId, "DBInstanceId", "i", "all", "指定数据库实例 ID (Specify DBInstance ID)")
	rdsWhiteListCmd.MarkFlagRequired("WhiteList")

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
		alirds.PrintDBInstancesList(rdsLsRegion, running, rdsLsSpecifiedDBInstanceId, rdsLsType, rdsLsFlushCache, rdsLsAllInfo)
	},
}

// RDS Account 相关操作
var rdsAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "添加云数据库帐号 (Add RDS account)",
	Long:  "添加云数据库帐号 (Add RDS account)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.AddRdsAccount(rdsAccountSpecifiedDBInstanceId, rdsAccountUserName)
	},
}

var rdsAccountLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出添加过的云数据库帐号 (Listing the added RDS accounts)",
	Long:  "列出添加过的云数据库帐号 (Listing the added RDS accounts)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.LsRdsAccount()
	},
}

var rdsAccountDelCmd = &cobra.Command{
	Use:   "del",
	Short: "删除所添加的云数据库帐号 (Deleting the added RDS account)",
	Long:  "删除所添加的云数据库帐号 (Deleting the added RDS account)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.DelRdsAccount()
	},
}

// RDS Public Access 相关操作
var rdsPublicCmd = &cobra.Command{
	Use:   "public",
	Short: "将云数据库设置为公开访问 (Set RDS to be publicly accessible)",
	Long:  "将云数据库设置为公开访问 (Set RDS to be publicly accessible)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.RdsPublic(rdsPublicSpecifiedDBInstanceId)
	},
}

var rdsPublicLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出已经设置过的公开访问地址 (List the public access addresses that have been configured)",
	Long:  "列出已经设置过的公开访问地址 (List the public access addresses that have been configured)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.RdsPublicLs()
	},
}

var rdsPublicCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "取消云数据库的公开访问 (Disable public access for the RDS instance)",
	Long:  "取消云数据库的公开访问 (Disable public access for the RDS instance)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.RdsPublicCancel()
	},
}

// RDS White List 相关操作
var rdsWhiteListCmd = &cobra.Command{
	Use:   "whiteList",
	Short: "为 RDS 添加白名单 (Add whitelist for RDS)",
	Long:  "为 RDS 添加白名单 (Add whitelist for RDS)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.RdsWhiteList(rdsWhiteListSpecifiedDBInstanceId, rdsWhiteList)
	},
}

var rdsWhiteListLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出已添加的白名单地址 (List the whitelisted addresses that have been added)",
	Long:  "列出已添加的白名单地址 (List the whitelisted addresses that have been added)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.RdsWhiteListLs()
	},
}

var rdsWhiteListDelCmd = &cobra.Command{
	Use:   "del",
	Short: "删除已添加的白名单 (Remove the whitelisted addresses that have been added)",
	Long:  "删除已添加的白名单 (Remove the whitelisted addresses that have been added)",
	Run: func(cmd *cobra.Command, args []string) {
		alirds.RdsWhiteListDel()
	},
}
