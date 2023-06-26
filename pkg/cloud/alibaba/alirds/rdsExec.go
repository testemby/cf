package alirds

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	"github.com/teamssix/cf/pkg/util/database"
	"github.com/teamssix/cf/pkg/util/errutil"
	"strings"
	"time"
)

func AddWhiteList(region string, specifiedDBInstanceId string, rdsWhiteList string) {
	request := rds.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.ModifyMode = "Append"
	request.DBInstanceId = specifiedDBInstanceId
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
			database.InsertTakeoverConsoleCache("alibabaRdsWhiteList", specifiedDBInstanceId, "", "", rdsWhiteList, "N/A", "N/A")
			fmt.Println("追加成功，正在查询当前白名单 (Appended successfully and the current whitelist is being queried)")
			time.Sleep(100 * time.Millisecond)
			var data [][]string
			for _, v := range GetWhiteListInfo(region, specifiedDBInstanceId) {
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
		request.DBInstanceId = TakeoverConsoleCache[0].PrimaryAccountId
		request.SecurityIps = TakeoverConsoleCache[0].LoginUrl

		_, err := RDSClient(region).ModifySecurityIps(request)
		errutil.HandleErrNoExit(err)
		fmt.Println("清理完成 (Clean up completed)")
	}
}
