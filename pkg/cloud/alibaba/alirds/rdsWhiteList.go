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
	WhiteListIPArrayName string
	WhiteListIPType      string
	WhiteListIPList      string
)

func RdsWhiteList(DBInstanceId string, WhiteList string) {
	var (
		specifiedDBInstanceId string
		region                string
		engine                string
		isSure                bool
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

	if WhiteList == "0.0.0.0/0" {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("添加 0.0.0.0/0 白名单可能会引发安全告警，确定添加吗？(Adding the 0.0.0.0/0 whitelist may trigger security alerts. Are you sure you want to proceed with the addition?)"),
			Default: false,
		}
		err := survey.AskOne(prompt, &isSure)
		errutil.HandleErr(err)
		if !isSure {
			log.Infoln("已中止操作 (The operation has been aborted.)")
			return
		}
	}

	log.Infof("正在为 %s RDS 实例添加 %s 白名单 (Adding the %s whitelist for the %s RDS instance.)", specifiedDBInstanceId, WhiteList, WhiteList, specifiedDBInstanceId)

	request := rds.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.ModifyMode = "Append"
	request.DBInstanceId = specifiedDBInstanceId
	request.SecurityIps = WhiteList

	_, err := RDSClient(region).ModifySecurityIps(request)
	errutil.HandleErr(err)
	for {
		if strings.Contains(WhiteListIPList, WhiteList) {
			break
		}
		log.Debugln("未发现添加的白名单地址，延时 1 秒后继续获取 (No added whitelist address found. Please wait for 3 seconds and try again to retrieve it.)")
		time.Sleep(time.Duration(1) * time.Second)
		ReturnGetWhiteListInfo(region, specifiedDBInstanceId, WhiteList)
	}
	database.InsertRDSWhiteListCache("alibaba", specifiedDBInstanceId, engine, WhiteListIPArrayName, WhiteListIPType, WhiteListIPList, WhiteList, region)
	log.Infof("%s 白名单添加成功 (Whitelist for %s has been successfully added.)", WhiteList, WhiteList)
	var (
		data              [][]string
		RDSWhiteListCache []pubutil.RDSWhiteListCache
		SN                int
	)
	RDSWhiteListCache = database.SelectRDSWhiteListCache("alibaba")
	for _, v := range RDSWhiteListCache {
		SN = SN + 1
		dataSingle := []string{strconv.Itoa(SN), v.DBInstanceId, v.Engine, v.IPArrayName, v.IPType, v.IPList, v.Region, v.CreateTime}
		if v.DBInstanceId == specifiedDBInstanceId && v.WhiteList == WhiteList {
			data = append(data, dataSingle)
		}
	}
	header = []string{"序号 (SN)", "实例 ID (Instance ID)", "数据库类型 (Type)", "IP 组名 (IP Array Name)", "IP 类型 (IP Type)", "白名单列表 (White List)", "区域 (Region)", "创建时间 (Create Time)"}
	var td = cloud.TableData{Header: header, Body: data}
	cloud.PrintTable(td, "")
}

func RdsWhiteListLs() {
	var (
		data              [][]string
		RDSWhiteListCache []pubutil.RDSWhiteListCache
		SN                int
	)
	RDSWhiteListCache = database.SelectRDSWhiteListCache("alibaba")
	for _, v := range RDSWhiteListCache {
		SN = SN + 1
		dataSingle := []string{strconv.Itoa(SN), v.DBInstanceId, v.Engine, v.IPArrayName, v.IPType, v.IPList, v.Region, v.CreateTime}
		data = append(data, dataSingle)
	}
	if len(data) == 0 {
		log.Infoln("未找到任何信息 (No information found.)")
	} else {
		header = []string{"序号 (SN)", "实例 ID (Instance ID)", "数据库类型 (Type)", "IP 组名 (IP Array Name)", "IP 类型 (IP Type)", "白名单列表 (White List)", "区域 (Region)", "创建时间 (Create Time)"}
		var td = cloud.TableData{Header: header, Body: data}
		cloud.PrintTable(td, "")
	}
}

func RdsWhiteListDel() {
	var (
		RDSWhiteListCache []pubutil.RDSWhiteListCache
		DBInstanceId      string
		WhiteList         string
		Region            string
	)
	RDSWhiteListCache = database.SelectRDSWhiteListCache("alibaba")

	if len(RDSWhiteListCache) == 0 {
		log.Infoln("未配置过白名单，无需删除 (No whitelist configured, no need to delete.)")
		return
	} else if len(RDSWhiteListCache) == 1 {
		DBInstanceId = RDSWhiteListCache[0].DBInstanceId
		WhiteList = RDSWhiteListCache[0].WhiteList
		Region = RDSWhiteListCache[0].Region
	} else {
		var (
			selectRdsIDList []string
			selectRdsID     string
			SN              int
		)
		for _, i := range RDSWhiteListCache {
			SN = SN + 1
			selectRdsIDList = append(selectRdsIDList, fmt.Sprintf("%s-%s-%s", strconv.Itoa(SN), i.DBInstanceId, i.WhiteList))
		}
		sort.Strings(selectRdsIDList)
		prompt := &survey.Select{
			Message: "选择一个帐号 (Choose a RDS instance): ",
			Options: selectRdsIDList,
		}
		survey.AskOne(prompt, &selectRdsID)
		for _, v := range RDSWhiteListCache {
			if strings.Contains(selectRdsID, v.DBInstanceId) {
				DBInstanceId = v.DBInstanceId
				WhiteList = v.WhiteList
				Region = v.Region
			}
		}
	}
	log.Infof("正在删除 %s RDS 实例中的 %s 白名单 (Deleting %s whitelist from %s RDS instance.)", DBInstanceId, WhiteList, WhiteList, DBInstanceId)
	request := rds.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.ModifyMode = "Delete"
	request.DBInstanceId = DBInstanceId
	request.SecurityIps = WhiteList

	_, err := RDSClient(Region).ModifySecurityIps(request)
	errutil.HandleErr(err)
	log.Infof("%s 白名单删除成功 (Whitelist for %s has been successfully deleted.)", WhiteList, WhiteList)
	database.DeleteRDSWhiteListCache("alibaba", DBInstanceId, WhiteList)
}

func ReturnGetWhiteListInfo(region string, specifiedDBInstanceId string, WhiteList string) {
	var DBInstanceIPArray []rds.DBInstanceIPArray
	DBInstanceIPArray = GetWhiteListInfo(region, specifiedDBInstanceId)
	for _, v := range DBInstanceIPArray {
		if strings.Contains(v.SecurityIPList, WhiteList) {
			WhiteListIPArrayName = v.DBInstanceIPArrayName
			WhiteListIPType = v.SecurityIPType
			WhiteListIPList = v.SecurityIPList
		}
	}
	log.Debugf("IPArrayName: %s, IPType: %s, IPList: %s", WhiteListIPArrayName, WhiteListIPType, WhiteListIPList)
}
