package alirds

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	"github.com/teamssix/cf/pkg/util"
	"github.com/teamssix/cf/pkg/util/database"
	"github.com/teamssix/cf/pkg/util/errutil"
	"github.com/teamssix/cf/pkg/util/pubutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

func AddRdsAccount(DBInstanceId string, userName string) {
	var (
		specifiedDBInstanceId string
		region                string
		engine                string
	)
	RDSCache := database.SelectRDSCacheFilter("alibaba", "all", "all", "all")
	if len(RDSCache) == 0 {
		log.Warnln("未找到 RDS 实例资源，请先使用 \"cf alibaba rds ls --flushCache\" 命令列出 RDS 实例 (RDS instance resource not found. Please use the \"cf alibaba rds ls --flushCache\" command to list RDS instances first.)")
		return
	}
	if DBInstanceId == "all" {
		if len(RDSCache) == 1 {
			specifiedDBInstanceId = RDSCache[0].DBInstanceId
			region = RDSCache[0].RegionId
			engine = RDSCache[0].Engine
		} else {
			var (
				selectRdsIDList []string
				selectRdsID     string
				SN              int
			)
			for _, i := range RDSCache {
				SN = SN + 1
				selectRdsIDList = append(selectRdsIDList, fmt.Sprintf("%s-%s-%s-%s)", strconv.Itoa(SN), i.DBInstanceId, i.Engine, i.RegionId))
			}
			sort.Strings(selectRdsIDList)
			prompt := &survey.Select{
				Message: "选择一个 RDS 实例 (Choose a RDS instance): ",
				Options: selectRdsIDList,
			}
			survey.AskOne(prompt, &selectRdsID)
			for _, v := range RDSCache {
				if strings.Contains(selectRdsID, v.DBInstanceId) {
					specifiedDBInstanceId = v.DBInstanceId
					region = v.RegionId
					engine = v.Engine
				}
			}
		}
	} else {
		for _, v := range RDSCache {
			if v.DBInstanceId == DBInstanceId {
				specifiedDBInstanceId = v.DBInstanceId
				region = v.RegionId
				engine = v.Engine
			}
		}
		if specifiedDBInstanceId == "" {
			log.Warnln("未找到 RDS 实例资源，请先使用 \"cf alibaba rds ls --flushCache\" 命令列出 RDS 实例 (RDS instance resource not found. Please use the \"cf alibaba rds ls --flushCache\" command to list RDS instances first.)")
			return
		}
	}

	password := util.GenerateRandomPasswords()
	request := rds.CreateCreateAccountRequest()
	request.DBInstanceId = specifiedDBInstanceId
	request.AccountName = userName
	request.AccountPassword = password
	request.AccountType = "Super"
	request.Scheme = "https"

	log.Infof("正在创建 %s 用户 (Creating user %s.)", userName, userName)

	_, err := RDSClient(region).CreateAccount(request)
	if err != nil {
		errutil.HandleErr(err)
	} else {
		database.InsertRDSAccountsCache("alibaba", specifiedDBInstanceId, engine, userName, password, region)
		log.Infof("%s 用户创建成功 (User %s created successfully.)", userName, userName)
		log.Debugln("正在授予权限 (Granting permissions.)")

		request := rds.CreateDescribeDatabasesRequest() // 查询实例下数据库名称
		request.Scheme = "https"
		request.DBInstanceId = specifiedDBInstanceId
		request.QueryParams["output"] = "cols=DBName,DBStatus,Engine rows=Databases.Database[]"
		response, err := RDSClient(region).DescribeDatabases(request)
		errutil.HandleErrNoExit(err)

		var (
			data             [][]string
			RDSAccountsCache []pubutil.RDSAccountsCache
			SN               int
		)
		RDSAccountsCache = database.SelectRDSAccountCache("alibaba")
		for _, v := range RDSAccountsCache {
			SN = SN + 1
			dataSingle := []string{strconv.Itoa(SN), v.DBInstanceId, v.Engine, v.UserName, v.Password, v.Region, v.CreateTime}
			if v.DBInstanceId == specifiedDBInstanceId && v.UserName == userName {
				data = append(data, dataSingle)
			}
		}
		header = []string{"序号 (SN)", "实例 ID (Instance ID)", "数据库类型 (Type)", "用户名 (User Name)", "密码 (Password)", "区域 (Region)", "创建时间 (Create Time)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")

		if len(response.Databases.Database) > 0 {
			var DBNames []string
			for _, v := range response.Databases.Database {
				DBNames = append(DBNames, v.DBName)
			}

			request := rds.CreateGrantAccountPrivilegeRequest()
			request.Scheme = "https"
			request.DBInstanceId = specifiedDBInstanceId
			request.AccountName = userName

			switch engine {
			case "MySQL":
				log.Debugln("MySQL 数据库无需对数据库具体授权，默认最高权限 (MySQL databases do not require specific database-level permissions as they have inherent high-level privileges by default.)")
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

			for _, v := range DBNames { // 为此帐号对实例下所有数据库授予最高权限
				request.DBName = v
				RDSClient(region).GrantAccountPrivilege(request)
			}
			time.Sleep(100 * time.Millisecond)
			DataBases := GetDataBases(region, specifiedDBInstanceId)
			if len(DataBases) > 0 {
				for _, v := range DataBases {
					privileges := make([]string, len(v.Accounts.AccountPrivilegeInfo))
					for i, privilege := range v.Accounts.AccountPrivilegeInfo {
						if privilege.AccountPrivilegeDetail == "" {
							privileges[i] = privilege.Account + " (" + privilege.AccountPrivilege + ")"
						} else {
							privileges[i] = privilege.Account + " (" + privilege.AccountPrivilege + ", " + privilege.AccountPrivilegeDetail + ")"
						}
					}
				}
			}
			log.Debugln("已为所有库授予权限 (Permissions have been granted to all databases.)")
		}
	}
}

func LsRdsAccount() {
	var (
		data             [][]string
		RDSAccountsCache []pubutil.RDSAccountsCache
		SN               int
	)
	RDSAccountsCache = database.SelectRDSAccountCache("alibaba")
	for _, v := range RDSAccountsCache {
		SN = SN + 1
		dataSingle := []string{strconv.Itoa(SN), v.DBInstanceId, v.Engine, v.UserName, v.Password, v.Region, v.CreateTime}
		data = append(data, dataSingle)
	}
	if len(data) == 0 {
		log.Infoln("未找到任何信息 (No information found.)")
	} else {
		header = []string{"序号 (SN)", "实例 ID (Instance ID)", "数据库类型 (Type)", "用户名 (User Name)", "密码 (Password)", "区域 (Region)", "创建时间 (Create Time)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")
	}
}

func DelRdsAccount() {
	var (
		RDSAccountsCache []pubutil.RDSAccountsCache
		DBInstanceId     string
		UserName         string
		Region           string
	)
	RDSAccountsCache = database.SelectRDSAccountCache("alibaba")

	if len(RDSAccountsCache) == 0 {
		log.Infoln("未创建过帐号，无需删除 (No create of the account, no need to delete)")
		return
	} else if len(RDSAccountsCache) == 1 {
		DBInstanceId = RDSAccountsCache[0].DBInstanceId
		UserName = RDSAccountsCache[0].UserName
		Region = RDSAccountsCache[0].Region
	} else {
		var (
			selectRdsIDList []string
			selectRdsID     string
			SN              int
		)
		for _, i := range RDSAccountsCache {
			SN = SN + 1
			selectRdsIDList = append(selectRdsIDList, fmt.Sprintf("%s-%s-%s-%s-%s", strconv.Itoa(SN), i.UserName, i.DBInstanceId, i.Engine, i.Region))
		}
		sort.Strings(selectRdsIDList)
		prompt := &survey.Select{
			Message: "选择一个帐号 (Choose a RDS instance): ",
			Options: selectRdsIDList,
		}
		survey.AskOne(prompt, &selectRdsID)
		for _, v := range RDSAccountsCache {
			if strings.Contains(selectRdsID, v.DBInstanceId) {
				DBInstanceId = v.DBInstanceId
				UserName = v.UserName
				Region = v.Region
			}
		}
	}
	request := rds.CreateDeleteAccountRequest()
	request.DBInstanceId = DBInstanceId
	request.AccountName = UserName
	request.Scheme = "https"
	_, err := RDSClient(Region).DeleteAccount(request)
	errutil.HandleErr(err)
	if err == nil {
		database.DeleteRDSAccountCache("alibaba", DBInstanceId)
		log.Infof("%s 用户删除成功 (%s user delete completed)", UserName, UserName)
	}
}
