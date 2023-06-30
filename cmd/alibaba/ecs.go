package alibaba

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	aliecs2 "github.com/teamssix/cf/pkg/cloud/alibaba/aliecs"
)

var (

	// ecs cmd
	ecsFlushCache bool

	// ecs ls
	running                  bool
	ecsLsAllRegions          bool
	ecsLsRegion              string
	ecsLsSpecifiedInstanceId string

	// ecs exec
	timeOut                    int
	userData                   bool
	batchCommand               bool
	ecsExecAllRegions          bool
	metaDataSTSToken           bool
	lhost                      string
	lport                      string
	command                    string
	scriptType                 string
	commandFile                string
	ecsExecRegion              string
	ecsExecSpecifiedInstanceId string
	userDataBackdoor           string

	// ecs imageShare
	accountId                        string
	ecsImageShareRegion              string
	ecsImageShareSpecifiedInstanceId string
)

func init() {
	alibabaCmd.AddCommand(ecsCmd)
	ecsCmd.AddCommand(ecsLsCmd)
	ecsCmd.AddCommand(ecsExecCmd)
	ecsCmd.AddCommand(ecsImageShareCmd)
	ecsImageShareCmd.AddCommand(ecsImageShareCmdLs)
	ecsImageShareCmd.AddCommand(ecsImageShareCmdCancel)

	// ecs cmd
	ecsCmd.PersistentFlags().BoolVar(&ecsFlushCache, "flushCache", false, "刷新缓存，不使用缓存数据 (Refresh the cache without using cached data)")

	// ecs ls
	ecsLsCmd.Flags().StringVarP(&ecsLsRegion, "region", "r", "all", "指定区域 ID (Specify region ID)")
	ecsLsCmd.Flags().StringVarP(&ecsLsSpecifiedInstanceId, "InstanceId", "i", "all", "指定实例 ID (Specify instance ID)")
	ecsLsCmd.Flags().BoolVar(&running, "running", false, "只显示正在运行的实例 (Show only running instances)")
	ecsLsCmd.Flags().BoolVarP(&ecsLsAllRegions, "allRegions", "a", false, "使用所有区域，包括私有区域 (Use all regions, including private regions)")

	// ecs exec
	ecsExecCmd.Flags().StringVarP(&ecsExecSpecifiedInstanceId, "InstanceId", "i", "all", "指定实例 ID (Specify Instance ID)")
	ecsExecCmd.Flags().StringVarP(&command, "command", "c", "", "设置待执行的命令 (Set the command you want to execute)")
	ecsExecCmd.Flags().StringVarP(&commandFile, "file", "f", "", "设置待执行的命令文件 (Set the command file you want to execute)")
	ecsExecCmd.Flags().StringVarP(&scriptType, "scriptType", "s", "auto", "设置执行脚本的类型 (Specify the type of script to execute) [sh|bat|ps]")
	ecsExecCmd.Flags().StringVar(&lhost, "lhost", "", "设置反弹 shell 的主机 IP (Set the ip of the listening host)")
	ecsExecCmd.Flags().StringVar(&lport, "lport", "", "设置反弹 shell 的主机端口 (Set the port of the listening host)")
	ecsExecCmd.Flags().BoolVarP(&batchCommand, "batchCommand", "b", false, "一键执行三要素，方便 HW (Batch execution of multiple commands used to prove permission acquisition)")
	ecsExecCmd.Flags().BoolVarP(&userData, "userData", "u", false, "一键获取实例中的用户数据 (Get the user data on the instance)")
	ecsExecCmd.Flags().StringVar(&userDataBackdoor, "userDataBackdoor", "", "用户数据后门，需要输入后门命令 (User data backdoor requires inputting the backdoor command)")
	ecsExecCmd.Flags().BoolVarP(&metaDataSTSToken, "metaDataSTSToken", "m", false, "一键获取实例元数据中的临时访问密钥 (Get the STS Token in the instance metadata)")
	ecsExecCmd.Flags().IntVarP(&timeOut, "timeOut", "t", 60, "设置命令执行结果的等待时间 (Set the command execution result waiting time)")
	ecsExecCmd.Flags().BoolVarP(&ecsExecAllRegions, "allRegions", "a", false, "使用所有区域，包括私有区域 (Use all regions, including private regions)")
	ecsExecCmd.Flags().StringVarP(&ecsExecRegion, "region", "r", "all", "指定区域 ID (Specify region ID)")

	// ecs imageShare
	ecsImageShareCmd.Flags().StringVarP(&accountId, "accountId", "a", "", "指定镜像共享的阿里云账号 ID (Specify the Alibaba Cloud account ID to share the image with)")
	ecsImageShareCmd.Flags().StringVarP(&ecsImageShareRegion, "region", "r", "all", "指定区域 ID (Specify region ID)")
	ecsImageShareCmd.Flags().StringVarP(&ecsImageShareSpecifiedInstanceId, "InstanceId", "i", "all", "指定实例 ID (Specify Instance ID)")
	ecsImageShareCmd.MarkFlagRequired("accountId")
}

var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "执行与弹性计算服务相关的操作 (Perform ecs-related operations)",
	Long:  "执行与弹性计算服务相关的操作 (Perform ecs-related operations)",
}

var ecsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列出所有的实例 (List all instances)",
	Long:  "列出所有的实例 (List all instances)",
	Run: func(cmd *cobra.Command, args []string) {
		aliecs2.PrintInstancesList(ecsLsRegion, running, ecsLsSpecifiedInstanceId, ecsFlushCache, ecsLsAllRegions)
	},
}

var ecsExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "在实例上执行命令 (Execute the command on the instance)",
	Long:  "在实例上执行命令 (Execute the command on the instance)",
	Run: func(cmd *cobra.Command, args []string) {
		if lhost != "" && lport == "" {
			log.Warnln("未指定反弹 shell 的主机端口 (The port of the listening host is not set)")
			cmd.Help()
		} else if lhost == "" && lport != "" {
			log.Warnln("未指定反弹 shell 的主机 IP (The ip of the listening host is not set)")
			cmd.Help()
		} else if command == "" && batchCommand == false && userData == false && metaDataSTSToken == false && commandFile == "" && lhost == "" && lport == "" && userDataBackdoor == "" {
			log.Warnf("还未指定要执行的命令 (The command to be executed has not been specified yet)\n")
			cmd.Help()
		} else {
			aliecs2.ECSExec(command, commandFile, scriptType, ecsExecSpecifiedInstanceId, ecsExecRegion, batchCommand, userData, metaDataSTSToken, ecsFlushCache, lhost, lport, timeOut, ecsExecAllRegions, userDataBackdoor)
		}
	},
}

var ecsImageShareCmd = &cobra.Command{
	Use:   "imageShare",
	Short: "共享实例镜像 (Share instance image)",
	Long:  "共享实例镜像 (Share instance image)",
	Run: func(cmd *cobra.Command, args []string) {
		if ecsImageShareRegion != "all" && ecsImageShareSpecifiedInstanceId == "all" {
			log.Warnln("未指定实例 ID (Instance ID not specified.)")
		} else if ecsImageShareRegion == "all" && ecsImageShareSpecifiedInstanceId != "all" {
			log.Warnln("未指定区域 (Region not specified.)")
		} else if len(accountId) != 16 {
			log.Warnln("账号 ID 输入有误，请确认后再进行尝试 (Incorrect account ID input. Please verify it and try again.)")
		} else {
			var isSure bool
			fmt.Println()
			prompt := &survey.Confirm{
				Message: "使用该利用手法，会在目标账户上新建一个镜像和快照，这会导致目标阿里云账户产生一定的费用。\n" +
					"  如果想取消镜像共享并删除所创建的镜像和快照，请使用 cf alibaba ecs imageShare cancel 命令，\n" +
					"  请在仔细权衡后再确认是否使用此利用手法。\n" +
					"  By using this exploitation technique, a new image and snapshot will be created \n" +
					"  on the target account, resulting in certain charges for the target Alibaba Cloud \n" +
					"  account. If you want to cancel image sharing and delete the created image and \n" +
					"  snapshot, please use the \"cf alibaba ecs imageShare cancel\" command. Please \n" +
					"  carefully consider before confirming whether to use this exploitation technique.)",
				Default: false,
			}
			survey.AskOne(prompt, &isSure)
			fmt.Println()
			if isSure {
				aliecs2.ECSImageShare(accountId, ecsImageShareRegion, ecsImageShareSpecifiedInstanceId)
			} else {
				log.Infoln("已中止操作 (The operation has been aborted.)")
			}
		}
	},
}

var ecsImageShareCmdLs = &cobra.Command{
	Use:   "ls",
	Short: "列出共享实例镜像信息 (Listing shared instance image information)",
	Long:  "列出共享实例镜像信息 (Listing shared instance image information)",
	Run: func(cmd *cobra.Command, args []string) {
		aliecs2.GetImageShare()
	},
}

var ecsImageShareCmdCancel = &cobra.Command{
	Use:   "cancel",
	Short: "取消共享并删除所创建的镜像与快照 (Canceling the sharing and deleting the created image and snapshot)",
	Long:  "取消共享并删除所创建的镜像与快照 (Canceling the sharing and deleting the created image and snapshot)",
	Run: func(cmd *cobra.Command, args []string) {
		aliecs2.ImageDelete()
	},
}
