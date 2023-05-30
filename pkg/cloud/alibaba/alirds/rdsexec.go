package alirds

import (
	"github.com/teamssix/cf/pkg/util/errutil"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	"github.com/teamssix/cf/pkg/util/cmdutil"
	"github.com/AlecAivazis/survey/v2"
	"fmt"
	"sort"
	"strings"
)

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


func PrintDBInstancesInfo(region string, running bool, specifiedDBInstanceID string, engine string, lsFlushCache bool) {
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

		for _, i := range InstancesList {
			specifiedDBInstanceID := i.DBInstanceId
			region := i.RegionId
			PrintNetInfo(region, specifiedDBInstanceID)
			PrintWhiteListInfo(region, specifiedDBInstanceID)
			PrintAccountInfo(region, specifiedDBInstanceID)
		}
	}
}
