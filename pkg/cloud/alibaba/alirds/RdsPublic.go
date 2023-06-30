package alirds

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	"github.com/teamssix/cf/pkg/util/database"
	"github.com/teamssix/cf/pkg/util/errutil"
	"github.com/teamssix/cf/pkg/util/pubutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	PublicIPAddress         string
	PublicConnectionAddress string
	PublicPort              string
)

func RdsPublic(DBInstanceId string) {
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
	log.Infof("正在将 %s RDS 实例设置为公开访问 (Setting the %s RDS instance to public access.)", specifiedDBInstanceId, specifiedDBInstanceId)
	request := rds.CreateAllocateInstancePublicConnectionRequest()
	request.Scheme = "https"
	request.DBInstanceId = specifiedDBInstanceId
	request.ConnectionStringPrefix = specifiedDBInstanceId + "to"
	switch engine {
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
		errutil.HandleErr(err)
		return
	} else {
		for {
			if PublicConnectionAddress != "" {
				break
			}
			log.Debugln("未找到连接地址，延时 3 秒后继续获取 (No connection address found. Retrying to fetch after a 3-second delay.)")
			time.Sleep(time.Duration(3) * time.Second)
			ReturnGetNetInfo(region, specifiedDBInstanceId)
		}
		database.InsertRDSPublicCache("alibaba", specifiedDBInstanceId, engine, PublicIPAddress, PublicConnectionAddress, PublicPort, region)
		log.Infoln("配置公开访问成功 (Public access configuration successful.)")
		var (
			data           [][]string
			RDSPublicCache []pubutil.RDSPublicCache
			SN             int
		)
		RDSPublicCache = database.SelectRDSPublicCache("alibaba")
		for _, v := range RDSPublicCache {
			SN = SN + 1
			dataSingle := []string{strconv.Itoa(SN), v.DBInstanceId, v.Engine, v.IPAddress, v.ConnectionAddress, v.Port, v.Region, v.CreateTime}
			if v.DBInstanceId == specifiedDBInstanceId {
				data = append(data, dataSingle)
			}
		}
		header = []string{"序号 (SN)", "实例 ID (Instance ID)", "数据库类型 (Type)", "IP 地址 (IP Address)", "连接地址 (Connection Address)", "端口 (Port)", "区域 (Region)", "创建时间 (Create Time)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")
	}
}

func RdsPublicLs() {
	var (
		data           [][]string
		RDSPublicCache []pubutil.RDSPublicCache
		SN             int
	)
	RDSPublicCache = database.SelectRDSPublicCache("alibaba")
	for _, v := range RDSPublicCache {
		SN = SN + 1
		dataSingle := []string{strconv.Itoa(SN), v.DBInstanceId, v.Engine, v.IPAddress, v.ConnectionAddress, v.Port, v.Region, v.CreateTime}
		data = append(data, dataSingle)
	}
	if len(data) == 0 {
		log.Infoln("未找到任何信息 (No information found.)")
	} else {
		header = []string{"序号 (SN)", "实例 ID (Instance ID)", "数据库类型 (Type)", "IP 地址 (IP Address)", "连接地址 (Connection Address)", "端口 (Port)", "区域 (Region)", "创建时间 (Create Time)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")
	}
}

func RdsPublicCancel() {
	var (
		RDSPublicCache    []pubutil.RDSPublicCache
		DBInstanceId      string
		ConnectionAddress string
		Region            string
	)
	RDSPublicCache = database.SelectRDSPublicCache("alibaba")

	if len(RDSPublicCache) == 0 {
		log.Infoln("未配置过公开访问，无需取消 (No create of the account, no need to delete)")
		return
	} else if len(RDSPublicCache) == 1 {
		DBInstanceId = RDSPublicCache[0].DBInstanceId
		ConnectionAddress = RDSPublicCache[0].ConnectionAddress
		Region = RDSPublicCache[0].Region
	} else {
		var (
			selectRdsIDList []string
			selectRdsID     string
			SN              int
		)
		for _, i := range RDSPublicCache {
			SN = SN + 1
			selectRdsIDList = append(selectRdsIDList, fmt.Sprintf("%s-%s", strconv.Itoa(SN), i.ConnectionAddress))
		}
		sort.Strings(selectRdsIDList)
		prompt := &survey.Select{
			Message: "选择一个账号 (Choose a RDS instance): ",
			Options: selectRdsIDList,
		}
		survey.AskOne(prompt, &selectRdsID)
		for _, v := range RDSPublicCache {
			if strings.Contains(selectRdsID, v.DBInstanceId) {
				DBInstanceId = v.DBInstanceId
				ConnectionAddress = v.ConnectionAddress
				Region = v.Region
			}
		}
	}
	request := rds.CreateReleaseInstancePublicConnectionRequest()
	request.Scheme = "https"
	request.DBInstanceId = DBInstanceId
	request.CurrentConnectionString = ConnectionAddress
	_, err := RDSClient(Region).ReleaseInstancePublicConnection(request)
	if err != nil {
		errutil.HandleErr(err)
		return
	} else {
		log.Infof("已取消 %s RDS 实例的公开访问 (Public access for the RDS instance %s has been disabled.)", DBInstanceId, DBInstanceId)
		database.DeleteRDSPublicCache("alibaba", DBInstanceId)
	}
}

func ReturnGetNetInfo(region string, specifiedDBInstanceId string) {
	var DBInstanceNetInfo []rds.DBInstanceNetInfo
	DBInstanceNetInfo = GetNetInfo(region, specifiedDBInstanceId)
	for _, v := range DBInstanceNetInfo {
		if v.IPType == "Public" {
			PublicIPAddress = v.IPAddress
			PublicConnectionAddress = v.ConnectionString
			PublicPort = v.Port
		}
	}
	log.Debugf("IPAddress: %s, ConnectionAddress: %s, Port: %s", PublicIPAddress, PublicConnectionAddress, PublicPort)
}
