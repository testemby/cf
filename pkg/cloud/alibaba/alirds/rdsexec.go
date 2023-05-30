package alirds

import (
	"github.com/teamssix/cf/pkg/util/errutil"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	// "github.com/teamssix/cf/pkg/util"
	"github.com/teamssix/cf/pkg/util/cmdutil"
	"github.com/AlecAivazis/survey/v2"
	"fmt"
	"sort"
	"strings"
)

// func AddAccount(region string, specifiedDBInstanceID string, rdsAccount string) {
// 	password := util.GenerateRandomPasswords()
// 	request := rds.CreateCreateAccountRequest()
// 	request.DBInstanceId = specifiedDBInstanceID
// 	request.AccountName = rdsAccount
// 	request.password = AccountPassword
// 	request.AccountType = "Super"
// 	request.Scheme = "https"

// 	_, err := RDSClient(region).CreateAccount(request)
// 	if err != nil {
// 		fmt.Println("创建失败，请检查是否具备 CreateAccount  权限或已存在同名用户 (Create failed, please check whether have AllocateInstancePublicConnection permissions or existing communications address)")
// 		return
// 	} else {
// 		fmt.Println("创建成功，当前用户信息： (Creating an external address succeeded. Querying the current connection address)")
// 		data := [][]string{
// 			{rdsAccount, password},
// 		}
// 		var header = []string{"用户名 (User Name)", "密码 (Password)"}
// 		var td = cloud.TableData{Header: header, Body: data}
// 		cloud.PrintTable(td, "")
// 	}
// }


func AddWhiteList(region string, specifiedDBInstanceID string, rdsWhiteList string) {
	request := rds.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.ModifyMode = "Append"
	request.DBInstanceId = specifiedDBInstanceID
	request.SecurityIps = rdsWhiteList

	if strings.Contains(rdsWhiteList, "0.0.0.0") {
		fmt.Println("由于该地址会触发安全告警，禁止使用！")
		return
	} else {
		_, err := RDSClient(region).ModifySecurityIps(request)
		if err != nil {
		fmt.Println("追加失败，请检查是否具备 ModifySecurityIps 权限和参数是否符合ip地址格式 (Failed to add the ip address. Check whether you have the ModifySecurityIps permission or whether the parameter matches the IP address format)")
			return
		} else {
			fmt.Println("追加成功，正在查询当前白名单 (Appended successfully and the current whitelist is being queried)")
			PrintWhiteListInfo(region, specifiedDBInstanceID)

			cleanWhiteList := []string{"Yes"}
			var cleanWhiteListChoose string
		    prompt := &survey.Select{
		        Message: "删除添加的白名单地址？按下回车后，将会删除刚才已添加的白名单地址，请确保攻击已结束 (Delete the added whitelist address? Press Enter to delete the whitelist address you just added. Please ensure that the attack is over)",
		        Options: cleanWhiteList,
		    }
		    survey.AskOne(prompt, &cleanWhiteListChoose)
		    if cleanWhiteListChoose == "Yes" {
		    	request.ModifyMode = "Delete"
		    	_, err = RDSClient(region).ModifySecurityIps(request)
		    	errutil.HandleErr(err)
		    	fmt.Println("清理完成 (Clean up completed)")
		    }
		}
	}
}

func CancelConnection(region string, specifiedDBInstanceID string, rdsConnectCancel string) {
	request := rds.CreateReleaseInstancePublicConnectionRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceID
	request.CurrentConnectionString = rdsConnectCancel
	_, err := RDSClient(region).ReleaseInstancePublicConnection(request)
	if err != nil {
		fmt.Println("关闭失败，请确认是否具备 ReleaseInstancePublicConnection 权限和地址是否正确 (Failed to shut down, please confirm whether have ReleaseInstancePublicConnection permissions and address is correct)")
		return
	} else {
		fmt.Println("清理完成 (Clean up completed)")
		PrintNetInfo(region, specifiedDBInstanceID)
	}
}

func CreateConnection(region string, specifiedDBInstanceID string, rdsConnect string, Engine string) {
	request := rds.CreateAllocateInstancePublicConnectionRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceID
	request.ConnectionStringPrefix = rdsConnect
	switch Engine {
		case "MySQL":
			request.Port = "3306"
		case "SQLServer":
			request.Port = "1433"
		case "PostgreSQL":
			request.Port = "5432"
		case "MariaDB":
			request.Port = "3306"
		default:
			request.Port = "3306"
	}

	_, err := RDSClient(region).AllocateInstancePublicConnection(request)
	if err != nil {
		fmt.Println("创建失败，请检查是否具备 AllocateInstancePublicConnection 权限或已存在外联地址 (Create failed, please check whether have AllocateInstancePublicConnection permissions or existing communications address)")
		return
	} else {
		fmt.Println("创建外联地址成功，正在查询当前连接地址 (Creating an external address succeeded. Querying the current connection address)")
		PrintNetInfo(region, specifiedDBInstanceID)
	}
}

func PrintNetInfo(region string, specifiedDBInstanceID string) {
	request := rds.CreateDescribeDBInstanceNetInfoRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceID
	request.QueryParams["output"] = "cols=IPAddress,ConnectionString,IPType,Port rows=DBInstanceNetInfos.DBInstanceNetInfo[]"
	response, err := RDSClient(region).DescribeDBInstanceNetInfo(request)
	errutil.HandleErrNoExit(err)

	var data [][]string
	for _, v := range response.DBInstanceNetInfos.DBInstanceNetInfo {
		row := []string{v.IPAddress, v.ConnectionString, v.IPType, v.Port}
		data = append(data, row)
	}
	var header = []string{"IP地址 (IPAddress)", "连接地址 (ConnectionString)", "连接类型 (IPType)", "端口 (Port)"}
	var td = cloud.TableData{Header: header, Body: data}
	cloud.PrintTable(td, "")
}

func PrintWhiteListInfo(region string, specifiedDBInstanceID string) {
	request := rds.CreateDescribeDBInstanceIPArrayListRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceID
	response, err := RDSClient(region).DescribeDBInstanceIPArrayList(request)
	errutil.HandleErrNoExit(err)

	var data [][]string
	for _, v := range response.Items.DBInstanceIPArray {
	    row := []string{v.DBInstanceIPArrayName, v.SecurityIPType, v.SecurityIPList}
	    data = append(data, row)
	}
	var header = []string{"白名单名称 (IPArrayName)", "IP类型 (SecurityIPType)", "IP列表 (SecurityIPList)"}
	var td = cloud.TableData{Header: header, Body: data}
	cloud.PrintTable(td, "")
}

func PrintAccountInfo(region string, specifiedDBInstanceID string) {
	request := rds.CreateDescribeAccountsRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceID
	response, err := RDSClient(region).DescribeAccounts(request)
	errutil.HandleErrNoExit(err)

	if len(response.Accounts.DBInstanceAccount) > 0 {
		var data [][]string
		for _, v := range response.Accounts.DBInstanceAccount {
			privileges := make([]string, len(v.DatabasePrivileges.DatabasePrivilege))
			for i, privilege := range v.DatabasePrivileges.DatabasePrivilege {
				privileges[i] = privilege.DBName + "(" + privilege.AccountPrivilege + ", " + privilege.AccountPrivilegeDetail + ")"
			}
			row := []string{v.AccountStatus, v.AccountName, v.AccountType, strings.Join(privileges, ", ")}
			data = append(data, row)
		}
		var header = []string{"账号状态 (AccountStatus)", "账号 (AccountName)", "账号类型 (AccountType)", "数据库权限 (DatabasePrivilege)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")
	}

}


func DBInstancesExec(region string, running bool, specifiedDBInstanceID string, engine string, lsFlushCache bool, rdsInfo bool, rdsConnect string, rdsConnectCancel string, rdsWhiteList string , rdsAccount string) {
	var InstancesList []DBInstances
	if lsFlushCache == false {
		data := cmdutil.ReadRDSCache("alibaba")
		for _, v := range data {
			switch {
			case specifiedDBInstanceID != "all" && region != "all":
				if specifiedDBInstanceID == v.DBInstanceId && region == v.RegionId {
					obj := DBInstances{
						DBInstanceId:     v.DBInstanceId,
						Engine:           v.Engine,
						EngineVersion:    v.EngineVersion,
						DBInstanceStatus: v.DBInstanceStatus,
						RegionId:         v.RegionId,
					}
					InstancesList = append(InstancesList, obj)
				}
			case specifiedDBInstanceID != "all" && region == "all":
				if specifiedDBInstanceID == v.DBInstanceId {
					obj := DBInstances{
						DBInstanceId:     v.DBInstanceId,
						Engine:           v.Engine,
						EngineVersion:    v.EngineVersion,
						DBInstanceStatus: v.DBInstanceStatus,
						RegionId:         v.RegionId,
					}
					InstancesList = append(InstancesList, obj)
				}
			case specifiedDBInstanceID == "all" && region != "all":
				if region == v.RegionId {
					obj := DBInstances{
					DBInstanceId:     v.DBInstanceId,
					Engine:           v.Engine,
					EngineVersion:    v.EngineVersion,
					DBInstanceStatus: v.DBInstanceStatus,
					RegionId:         v.RegionId,
					}
					InstancesList = append(InstancesList, obj)
				}
			case specifiedDBInstanceID == "all" && region == "all":
				obj := DBInstances{
					DBInstanceId:     v.DBInstanceId,
					Engine:           v.Engine,
					EngineVersion:    v.EngineVersion,
					DBInstanceStatus: v.DBInstanceStatus,
					RegionId:         v.RegionId,
				}
				InstancesList = append(InstancesList, obj)
			}
		}
	} else {
		InstancesList = ReturnDBInstancesList(region, running, specifiedDBInstanceID, engine)
	}

	if len(InstancesList) == 0 {
		if specifiedDBInstanceID == "all" {
			log.Warnf("未发现实例，可以使用 --flushCache 刷新缓存后再试 (No instances found, You can use the --flushCache command to flush the cache and try again)")
		} else {
			log.Warnf("未找到 %s 实例的相关信息 (No information found about the %s instance)", specifiedDBInstanceID, specifiedDBInstanceID)
		}
	} else {
		if specifiedDBInstanceID == "all" {
			var (
				selectInstanceIDList []string
				selectInstanceID     string
			)
			selectInstanceIDList = append(selectInstanceIDList, "全部实例 (all instances)")
			for _, i := range InstancesList {
				selectInstanceIDList = append(selectInstanceIDList, fmt.Sprintf("%s (%s)", i.DBInstanceId, i.Engine))
			}
			sort.Strings(selectInstanceIDList)
			prompt := &survey.Select{
				Message: "选择一个实例 (Choose a instance): ",
				Options: selectInstanceIDList,
			}
			survey.AskOne(prompt, &selectInstanceID)
			for _, j := range InstancesList {
				if selectInstanceID != "all" {
					if selectInstanceID == fmt.Sprintf("%s (%s)", j.DBInstanceId, j.Engine) {
						InstancesList = nil
						InstancesList = append(InstancesList, j)
					}
				}
			}
		}

		var num = 0
		for _, i := range InstancesList {
			specifiedDBInstanceID := i.DBInstanceId
			region := i.RegionId

			if rdsInfo {
				PrintNetInfo(region, specifiedDBInstanceID)
				PrintWhiteListInfo(region, specifiedDBInstanceID)
				PrintAccountInfo(region, specifiedDBInstanceID)			
			}

			if i.DBInstanceStatus == "Running"{
				num += 1
				if rdsConnect != "" {
					CreateConnection(region, specifiedDBInstanceID, rdsConnect, i.Engine)
				} else if rdsConnectCancel != "" {
					var isSure string
					prompt := &survey.Input{
				        Message: "警告：该方法应当用于取消攻击者创建的实例外联地址，请勿用于删除原有的实例外联地址，否则将造成不可逆的后果。如果已明确并确认要使用此方法，请键入Yes (Warning: This method should be used to cancel the external address of the instance created by the attacker. Do not delete the original external address of the instance. Otherwise, irreversible consequences will be caused. If you are explicit and sure to use this method, type Yes)",
				    }
				    survey.AskOne(prompt, &isSure)
				    if isSure == "Yes" {
				    	CancelConnection(region, specifiedDBInstanceID, rdsConnectCancel)
				    }
				} else if rdsWhiteList != "" {
					AddWhiteList(region, specifiedDBInstanceID, rdsWhiteList)
				} else if rdsAccount != "" {
					fmt.Println("功能暂不可用")
				}
			}
		}
	}
}
