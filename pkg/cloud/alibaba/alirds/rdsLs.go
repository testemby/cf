package alirds

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/gookit/color"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	"github.com/teamssix/cf/pkg/util"
	"github.com/teamssix/cf/pkg/util/cmdutil"
	"github.com/teamssix/cf/pkg/util/errutil"
	"sort"
	"strconv"
	"strings"
)

var (
	DescribeDBInstancesOut []DBInstances
	TraversedRegions       []string
	TimestampType          = util.ReturnTimestampType("alibaba", "rds")
	header                 = []string{"序号 (SN)", "数据库 ID (DB ID)", "数据库类型 (DB Engine)", "数据库版本 (DB Engine Version)", "数据库状态 (DB Staus)", "区域 ID (Region ID)"}
)

type DBInstances struct {
	DBInstanceId     string
	Engine           string
	EngineVersion    string
	DBInstanceStatus string
	RegionId         string
}

func DescribeDBInstances(region string, running bool, specifiedDBInstanceId string, engine string, NextToken string, output bool) ([]DBInstances, error) {
	if !in(region, TraversedRegions) {
		TraversedRegions = append(TraversedRegions, region)
	}
	request := rds.CreateDescribeDBInstancesRequest()
	request.PageSize = requests.NewInteger(100)
	request.Scheme = "https"
	if NextToken != "" {
		request.NextToken = NextToken
	}
	if running == true {
		request.DBInstanceStatus = "Running"
	}
	if specifiedDBInstanceId != "all" {
		request.DBInstanceId = specifiedDBInstanceId
	}
	if engine != "all" {
		request.Engine = engine
	}
	response, err := RDSClient(region).DescribeDBInstances(request)
	errutil.HandleErrNoExit(err)
	DBInstancesList := response.Items.DBInstance
	if output {
		log.Infof("正在 %s 区域中查找数据库实例 (Looking for DBInstances in the %s region)", region, region)
	}
	if len(DBInstancesList) != 0 {
		if output {
			log.Warnf("在 %s 区域下找到 %d 个数据库实例 (Found %d DBInstances in %s region)", region, len(DBInstancesList), len(DBInstancesList), region)
		}
		for _, i := range DBInstancesList {
			obj := DBInstances{
				DBInstanceId:     i.DBInstanceId,
				Engine:           i.Engine,
				EngineVersion:    i.EngineVersion,
				RegionId:         i.RegionId,
				DBInstanceStatus: i.DBInstanceStatus,
			}
			DescribeDBInstancesOut = append(DescribeDBInstancesOut, obj)
		}
	}
	NextToken = response.NextToken
	if NextToken != "" && !in(region, TraversedRegions) {
		log.Tracef("Next Token: %s", NextToken)
		_, _ = DescribeDBInstances(region, running, specifiedDBInstanceId, engine, NextToken, true)
	}
	return DescribeDBInstancesOut, err
}

func ReturnDBInstancesList(region string, running bool, specifiedDBInstanceId string, engine string) []DBInstances {
	var DBInstancesList []DBInstances
	var DBInstance []DBInstances
	if region == "all" {
		var RegionsList []string
		for _, i := range GetRDSRegions() {
			RegionsList = append(RegionsList, i.RegionId)
		}
		RegionsList = RemoveRepeatedElement(RegionsList)
		for _, j := range RegionsList {
			DBInstance, _ = DescribeDBInstances(j, running, specifiedDBInstanceId, engine, "", true)
			DescribeDBInstancesOut = nil
			for _, i := range DBInstance {
				DBInstancesList = append(DBInstancesList, i)
			}
		}
	} else {
		DBInstancesList, _ = DescribeDBInstances(region, running, specifiedDBInstanceId, engine, "", true)
	}
	return DBInstancesList
}

func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return newArr
}

func PrintDBInstancesListRealTime(region string, running bool, specifiedDBInstanceId string, engine string) {
	DBInstancesList := ReturnDBInstancesList(region, running, specifiedDBInstanceId, engine)
	var data = make([][]string, len(DBInstancesList))
	for i, o := range DBInstancesList {
		SN := strconv.Itoa(i + 1)
		data[i] = []string{SN, o.DBInstanceId, o.Engine, o.EngineVersion, o.DBInstanceStatus, o.RegionId}
	}
	var td = cloud.TableData{Header: header, Body: data}
	if len(data) == 0 {
		log.Info("未发现 RDS (No RDS found)")
	} else {
		Caption := "RDS 资源 (RDS resources)"
		cloud.PrintTable(td, Caption)
		util.WriteTimestamp(TimestampType)
	}
	cmdutil.WriteCacheFile(td, "alibaba", "rds", region, specifiedDBInstanceId)
}

func PrintDBInstancesListHistory(region string, running bool, specifiedDBInstanceId string, engine string) {
	cmdutil.PrintRDSCacheFile(header, region, specifiedDBInstanceId, engine, "alibaba", "RDS")
}

func PrintDBInstancesList(region string, running bool, specifiedDBInstanceId string, engine string, lsFlushCache bool, all bool) {
	if all {
		DBInstancesList := ReturnDBInstancesList("all", false, "all", "all")
		if len(DBInstancesList) == 0 {
			log.Info("未发现 RDS 资源 (No RDS resources found)")
		}
		for k, v := range DBInstancesList {
			if len(DBInstancesList) > 1 {
				color.Tag("danger").Println(fmt.Sprintf("\n%d %s 实例信息 (%s Instance information)", k+1, v.DBInstanceId, v.DBInstanceId))
			}

			color.Tag("warn").Println("\n基础信息 (Basic information)")
			color.Tag("info").Print("ID: ")
			fmt.Println(v.DBInstanceId)
			color.Tag("info").Print("类型 (Type): ")
			fmt.Println(v.Engine)
			color.Tag("info").Print("版本 (Version): ")
			fmt.Println(v.EngineVersion)
			color.Tag("info").Print("状态 (Status): ")
			fmt.Println(v.DBInstanceStatus)
			color.Tag("info").Print("区域 (Region): ")
			fmt.Println(v.RegionId)

			color.Tag("warn").Println("\n网络信息 (Network information)")
			for _, v := range GetNetInfo(v.RegionId, v.DBInstanceId) {
				color.Tag("info").Print("端口 (Port): ")
				fmt.Println(v.Port)
				color.Tag("info").Print("类型 (Type): ")
				fmt.Println(v.IPType)
				color.Tag("info").Print("IP 地址 (IP address): ")
				fmt.Println(v.IPAddress)
				color.Tag("info").Print("连接地址 (Connection address): ")
				fmt.Println(v.ConnectionString)
				color.Tag("info").Print("VPC ID: ")
				if v.VPCId == "" {
					fmt.Println("N/A")
				} else {
					fmt.Println(v.VPCId)
				}
				color.Tag("info").Print("交换机 ID (Switch ID): ")
				if v.VPCId == "" {
					fmt.Println("N/A\n")
				} else {
					fmt.Println(v.VSwitchId + "\n")
				}
			}

			color.Tag("warn").Println("安全组信息 (Security groups information)")
			for _, v := range GetWhiteListInfo(v.RegionId, v.DBInstanceId) {
				color.Tag("info").Print("类型 (Type): ")
				fmt.Println(v.SecurityIPType)
				color.Tag("info").Print("名称 (Name): ")
				fmt.Println(v.DBInstanceIPArrayName)
				color.Tag("info").Print("属性 (Attribute): ")
				if v.DBInstanceIPArrayAttribute == "" {
					fmt.Println("N/A")
				} else {
					fmt.Println(v.DBInstanceIPArrayAttribute)
				}
				color.Tag("info").Print("IP 列表 (IP List): ")
				fmt.Println(v.SecurityIPList + "\n")
			}

			color.Tag("warn").Println("帐号信息 (Accounts information)")
			RDSAccountInfo := GetAccountInfo(v.RegionId, v.DBInstanceId)
			if len(RDSAccountInfo) > 0 {
				for _, v := range RDSAccountInfo {
					privileges := make([]string, len(v.DatabasePrivileges.DatabasePrivilege))
					for i, privilege := range v.DatabasePrivileges.DatabasePrivilege {
						privileges[i] = privilege.DBName + "(" + privilege.AccountPrivilege + ", " + privilege.AccountPrivilegeDetail + ")"
					}
					color.Tag("info").Print("名称 (Name): ")
					fmt.Println(v.AccountName)
					color.Tag("info").Print("类型 (Type): ")
					fmt.Println(v.AccountType)
					color.Tag("info").Print("状态 (Status): ")
					fmt.Println(v.AccountStatus)
					color.Tag("info").Print("描述 (Description): ")
					fmt.Println(v.AccountDescription)
					color.Tag("info").Print("权限 (Privilege): ")
					privilege := strings.Join(privileges, ", ")
					if v.AccountType == "Super" {
						fmt.Println("Super \n")
					} else if privilege == "" {
						fmt.Println("N/A \n")
					} else {
						fmt.Println(privilege + "\n")
					}
				}
			} else {
				fmt.Println("没有找到任何信息 (No information found)\n")
			}

			color.Tag("warn").Println("数据库信息 (Databases information)")
			DataBases := GetDataBases(v.RegionId, v.DBInstanceId)
			if len(DataBases) > 0 {
				for _, v := range DataBases {
					color.Tag("info").Print("名称 (Name): ")
					fmt.Println(v.DBName)
					color.Tag("info").Print("状态 (Status): ")
					fmt.Println(v.DBStatus)
					color.Tag("info").Print("描述 (Description): ")
					fmt.Println(v.DBDescription)
					color.Tag("info").Println("帐号 (Account): ")
					for _, v2 := range v.Accounts.AccountPrivilegeInfo {
						color.Tag("info").Print("    帐号名 (Account Name): ")
						fmt.Println(v2.Account)
						color.Tag("info").Print("    帐号权限 (Account Privilege): ")
						fmt.Println(v2.AccountPrivilege)
						color.Tag("info").Print("    帐号权限细节 (Account Privilege Detail): ")
						if v2.AccountPrivilegeDetail == "" {
							fmt.Println("N/A")
						} else {
							fmt.Println(v2.AccountPrivilegeDetail)
						}
					}
				}
			} else {
				fmt.Println("没有找到任何信息 (No information found)\n")
			}
		}
	} else {
		if lsFlushCache {
			PrintDBInstancesListRealTime(region, running, specifiedDBInstanceId, engine)
		} else {
			oldTimestamp := util.ReadTimestamp(TimestampType)
			if oldTimestamp == 0 {
				PrintDBInstancesListRealTime(region, running, specifiedDBInstanceId, engine)
			} else if util.IsFlushCache(oldTimestamp) {
				PrintDBInstancesListRealTime(region, running, specifiedDBInstanceId, engine)
			} else {
				util.TimeDifference(oldTimestamp)
				PrintDBInstancesListHistory(region, running, specifiedDBInstanceId, engine)
			}
		}
	}
}

func in(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	if index < len(str_array) && str_array[index] == target {
		return true
	}
	return false
}

func GetNetInfo(region string, specifiedDBInstanceId string) []rds.DBInstanceNetInfo {
	request := rds.CreateDescribeDBInstanceNetInfoRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceId
	request.QueryParams["output"] = "cols=IPAddress,ConnectionString,IPType,Port rows=DBInstanceNetInfos.DBInstanceNetInfo[]"
	response, err := RDSClient(region).DescribeDBInstanceNetInfo(request)
	errutil.HandleErrNoExit(err)
	return response.DBInstanceNetInfos.DBInstanceNetInfo
}

func GetWhiteListInfo(region string, specifiedDBInstanceId string) []rds.DBInstanceIPArray {
	request := rds.CreateDescribeDBInstanceIPArrayListRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceId
	response, err := RDSClient(region).DescribeDBInstanceIPArrayList(request)
	errutil.HandleErrNoExit(err)
	return response.Items.DBInstanceIPArray
}

func GetAccountInfo(region string, specifiedDBInstanceId string) []rds.DBInstanceAccount {
	request := rds.CreateDescribeAccountsRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceId
	response, err := RDSClient(region).DescribeAccounts(request)
	errutil.HandleErrNoExit(err)
	return response.Accounts.DBInstanceAccount
}

func GetDataBases(region string, specifiedDBInstanceId string) []rds.Database {
	request := rds.CreateDescribeDatabasesRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceId
	request.QueryParams["output"] = "cols=DBName,DBStatus,Engine rows=Databases.Database[]"
	response, err := RDSClient(region).DescribeDatabases(request)
	errutil.HandleErrNoExit(err)
	return response.Databases.Database
}
