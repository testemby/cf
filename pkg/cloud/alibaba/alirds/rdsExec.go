package alirds

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	"github.com/teamssix/cf/pkg/util"
	"github.com/teamssix/cf/pkg/util/cmdutil"
	"github.com/teamssix/cf/pkg/util/database"
	"github.com/teamssix/cf/pkg/util/errutil"
	"sort"
	"strings"
	"time"
)

func AddAccount(region string, specifiedDBInstanceID string, rdsAccount string, Engine string) {
	password := util.GenerateRandomPasswords()
	request := rds.CreateCreateAccountRequest()
	request.DBInstanceId = specifiedDBInstanceID
	request.AccountName = rdsAccount
	request.AccountPassword = password
	request.AccountType = "Super"
	request.Scheme = "https"

	_, err := RDSClient(region).CreateAccount(request)
	if err != nil {
		fmt.Println("创建失败，请检查是否具备 CreateAccount 权限或已存在同名用户 (Create failed, please check whether have CreateAccount permissions or existing communications address)")
		return
	} else {
		database.InsertTakeoverConsoleCache("alibabaRds", specifiedDBInstanceID, rdsAccount, password, "N/A", "N/A", "N/A")
		fmt.Println("创建成功，当前用户信息： (Creating an external address succeeded. Querying the current connection address)")
		data := [][]string{
			{rdsAccount, password},
		}
		var header = []string{"用户名 (User Name)", "密码 (Password)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")

		request := rds.CreateDescribeDatabasesRequest() //查询实例下数据库名称
		request.Scheme = "https"
		request.DBInstanceId = specifiedDBInstanceID
		request.QueryParams["output"] = "cols=DBName,DBStatus,Engine rows=Databases.Database[]"
		response, err := RDSClient(region).DescribeDatabases(request)
		errutil.HandleErrNoExit(err)

		if len(response.Databases.Database) > 0 {
			var DBNames []string

			for _, v := range response.Databases.Database {
				DBNames = append(DBNames, v.DBName)
			}

			request := rds.CreateGrantAccountPrivilegeRequest()
			request.Scheme = "https"
			request.DBInstanceId = specifiedDBInstanceID
			request.AccountName = rdsAccount

			switch Engine {
			case "MySQL":
				fmt.Println("MySQL数据库无需对数据库具体授权，默认最高权限")
				return
			case "SQLServer":
				request.AccountPrivilege = "DBOwner"
			case "PostgreSQL":
				request.AccountPrivilege = "DBOwner"
			case "MariaDB":
				request.AccountPrivilege = "ReadWrite"
			default:
				request.AccountPrivilege = "ReadWrite"
			}

			for _, v := range DBNames { //为此帐号对实例下所有数据库授予最高权限
				request.DBName = v
				RDSClient(region).GrantAccountPrivilege(request)
			}
			time.Sleep(100 * time.Millisecond)
			DataBases := GetDataBases(region, specifiedDBInstanceID)
			if len(DataBases) > 0 {
				var data [][]string
				for _, v := range DataBases {
					privileges := make([]string, len(v.Accounts.AccountPrivilegeInfo))
					for i, privilege := range v.Accounts.AccountPrivilegeInfo {
						if privilege.AccountPrivilegeDetail == "" {
							privileges[i] = privilege.Account + " (" + privilege.AccountPrivilege + ")"
						} else {
							privileges[i] = privilege.Account + " (" + privilege.AccountPrivilege + ", " + privilege.AccountPrivilegeDetail + ")"
						}
					}
					row := []string{v.DBName, strings.Join(privileges, ", ")}
					data = append(data, row)
				}
				var header = []string{"数据库名 (DBName)", "授权情况 (privileges)"}
				var td = cloud.TableData{Header: header, Body: data}
				cloud.PrintTable(td, "")
			}
		}
	}
}

func DeleteAccount(region string) {
	TakeoverConsoleCache := database.SelectTakeoverConsoleCache("alibabaRds")
	if len(TakeoverConsoleCache) == 0 {
		log.Infoln("未创建过帐号，无需取消 (No create of the account, no need to cancel)")
	} else {
		request := rds.CreateDeleteAccountRequest()
		request.DBInstanceId = TakeoverConsoleCache[0].PrimaryAccountID
		request.AccountName = TakeoverConsoleCache[0].UserName
		request.Scheme = "https"
		_, err := RDSClient(region).DeleteAccount(request)
		if err != nil {
			fmt.Println("删除失败，请检查是否具备 DeleteAccount 权限 (Delete failed, please check whether have DeleteAccount permissions)")
			return
		} else {
			database.DeleteTakeoverConsoleCache("alibabaRds")
			fmt.Println("清理完成 (Clean up completed)")
		}
	}
}

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
			database.InsertTakeoverConsoleCache("alibabaRdsWhiteList", specifiedDBInstanceID, "", "", rdsWhiteList, "N/A", "N/A")
			fmt.Println("追加成功，正在查询当前白名单 (Appended successfully and the current whitelist is being queried)")
			time.Sleep(100 * time.Millisecond)
			var data [][]string
			for _, v := range GetWhiteListInfo(region, specifiedDBInstanceID) {
				row := []string{v.DBInstanceIPArrayName, v.SecurityIPType, v.SecurityIPList}
				data = append(data, row)
			}
			var header = []string{"白名单名称 (IPArrayName)", "IP类型 (SecurityIPType)", "IP列表 (SecurityIPList)"}
			var td = cloud.TableData{Header: header, Body: data}
			cloud.PrintTable(td, "")

		}
	}
}

func DeleteWhiteList(region string) {
	TakeoverConsoleCache := database.SelectTakeoverConsoleCache("alibabaRdsWhiteList")
	if len(TakeoverConsoleCache) == 0 {
		log.Infoln("未追加过白名单，无需取消 (No append of the whitelist, no need to cancel)")
	} else {
		request := rds.CreateModifySecurityIpsRequest()
		request.Scheme = "https"
		request.ModifyMode = "Delete"
		request.DBInstanceId = TakeoverConsoleCache[0].PrimaryAccountID
		request.SecurityIps = TakeoverConsoleCache[0].LoginUrl

		_, err := RDSClient(region).ModifySecurityIps(request)
		errutil.HandleErrNoExit(err)
		fmt.Println("清理完成 (Clean up completed)")
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
		database.InsertTakeoverConsoleCache("alibabaRdsConnect", specifiedDBInstanceID, "", "", rdsConnect, "N/A", "N/A")
		fmt.Println("创建外联地址成功，正在查询当前连接地址 (Creating an external address succeeded. Querying the current connection address)")
		time.Sleep(100 * time.Millisecond)
		var data [][]string
		for _, v := range GetNetInfo(region, specifiedDBInstanceID) {
			row := []string{v.IPAddress, v.ConnectionString, v.IPType, v.Port}
			data = append(data, row)
		}
		var header = []string{"IP地址 (IPAddress)", "连接地址 (ConnectionString)", "连接类型 (IPType)", "端口 (Port)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")
	}
}

func CancelConnection(region string) {
	TakeoverConsoleCache := database.SelectTakeoverConsoleCache("alibabaRdsConnect")
	if len(TakeoverConsoleCache) == 0 {
		log.Infoln("未创建过外联地址，无需取消 (No create of the connection, no need to cancel)")
	} else {
		request := rds.CreateReleaseInstancePublicConnectionRequest()
		request.Scheme = "https"
		request.DBInstanceId = TakeoverConsoleCache[0].PrimaryAccountID
		request.CurrentConnectionString = TakeoverConsoleCache[0].LoginUrl
		_, err := RDSClient(region).ReleaseInstancePublicConnection(request)
		if err != nil {
			fmt.Println("关闭失败，请确认是否具备 ReleaseInstancePublicConnection 权限和地址是否正确 (Failed to shut down, please confirm whether have ReleaseInstancePublicConnection permissions and address is correct)")
			return
		} else {
			fmt.Println("清理完成 (Clean up completed)")
		}
	}
}

func DBInstancesExec(region string, running bool, specifiedDBInstanceID string, engine string, lsFlushCache bool, rdsConnect string, rdsConnectCancel bool, rdsWhiteList string, rdsWhiteListCancel bool, rdsAccount string, rdsAccountCancel bool) {
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

			if i.DBInstanceStatus == "Running" {
				num += 1
				if rdsConnect != "" {
					CreateConnection(region, specifiedDBInstanceID, rdsConnect, i.Engine)
				} else if rdsConnectCancel {
					CancelConnection(region)
				} else if rdsWhiteList != "" {
					AddWhiteList(region, specifiedDBInstanceID, rdsWhiteList)
				} else if rdsWhiteListCancel {
					DeleteWhiteList(region)
				} else if rdsAccount != "" {
					AddAccount(region, specifiedDBInstanceID, rdsAccount, i.Engine)
				} else if rdsAccountCancel {
					DeleteAccount(region)
				}
			}
		}
	}
}
