package aliecs

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/cloud"
	"github.com/teamssix/cf/pkg/util"
	"github.com/teamssix/cf/pkg/util/cmdutil"
	"github.com/teamssix/cf/pkg/util/errutil"
)

type Instances struct {
	InstanceId           string
	InstanceName         string
	OSName               string
	OSType               string
	Status               string
	PrivateIpAddress     string
	PublicIpAddress      string
	RegionId             string
	CloudAssistantStatus string
}

var (
	DescribeInstancesOut []Instances
	TimestampType        = util.ReturnTimestampType("alibaba", "ecs")
	header               = []string{"序号 (SN)", "实例 ID (Instance ID)", "实例名称 (Instance Name)", "系统名称 (OS Name)", "系统类型 (OS Type)", "状态 (Status)", "公网 IP (Public IP)", "区域 ID (Region ID)"}
)

func CloudAssistantStatus(region, SpecifiedInstanceId, OSType string) string {
	request := ecs.CreateDescribeCloudAssistantStatusRequest()
	request.RegionId = region
	request.OSType = OSType
	request.InstanceId = &[]string{SpecifiedInstanceId}
	response, err := ECSClient(region).DescribeCloudAssistantStatus(request)
	errutil.HandleErr(err)
	if len(response.InstanceCloudAssistantStatusSet.InstanceCloudAssistantStatus) > 0 {
		cloudAssistantStatus := response.InstanceCloudAssistantStatusSet.InstanceCloudAssistantStatus[0]
		log.Debugf("Instance Cloud Assistant Status: %s", cloudAssistantStatus.CloudAssistantStatus)
		if cloudAssistantStatus.CloudAssistantStatus == "true" {
			return "True"
		} else {
			return "False"
		}
	} else {
		return "False"
	}
}

func DescribeInstances(region string, running bool, SpecifiedInstanceId string, NextToken string) []Instances {
	var response *ecs.DescribeInstancesResponse
	request := ecs.CreateDescribeInstancesRequest()
	request.PageSize = requests.NewInteger(100)
	if NextToken != "" {
		request.NextToken = NextToken
	}
	request.Scheme = "https"
	if running == true {
		request.Status = "Running"
	}
	if SpecifiedInstanceId != "all" {
		request.InstanceIds = fmt.Sprintf("[\"%s\"]", SpecifiedInstanceId)
	}
	log.Infof("正在 %s 区域中查找实例 (Looking for instances in the %s region)", region, region)
	response, err := ECSClient(region).DescribeInstances(request)
	errutil.HandleErr(err)
	InstancesList := response.Instances.Instance
	if len(InstancesList) != 0 {
		log.Warnf("在 %s 区域下找到 %d 个实例 (Found %d instances in %s region)", region, len(InstancesList), len(InstancesList), region)
		for _, i := range InstancesList {
			// When the instance has multiple IPs, it is presented in a different format.
			var PrivateIpAddressList []string
			var PublicIpAddressList []string
			var PrivateIpAddress string
			var PublicIpAddress string
			for _, m := range i.NetworkInterfaces.NetworkInterface {
				for _, n := range m.PrivateIpSets.PrivateIpSet {
					PrivateIpAddressList = append(PrivateIpAddressList, n.PrivateIpAddress)
				}
			}
			a, _ := json.Marshal(PrivateIpAddressList)

			if len(PrivateIpAddressList) == 1 {
				PrivateIpAddress = PrivateIpAddressList[0]
			} else {
				PrivateIpAddress = string(a)
			}

			PublicIpAddressList = i.PublicIpAddress.IpAddress
			b, _ := json.Marshal(PublicIpAddressList)
			if len(PublicIpAddressList) == 1 {
				PublicIpAddress = i.PublicIpAddress.IpAddress[0]
			} else if len(PublicIpAddressList) == 0 {
				PublicIpAddress = ""
			} else {
				PublicIpAddress = string(b)
			}
			InstanceCloudAssistantStatus := CloudAssistantStatus(i.RegionId, i.InstanceId, i.OSType)
			obj := Instances{
				InstanceId:           i.InstanceId,
				InstanceName:         i.InstanceName,
				OSName:               i.OSName,
				OSType:               i.OSType,
				Status:               i.Status,
				PrivateIpAddress:     PrivateIpAddress,
				PublicIpAddress:      PublicIpAddress,
				RegionId:             i.RegionId,
				CloudAssistantStatus: InstanceCloudAssistantStatus,
			}
			DescribeInstancesOut = append(DescribeInstancesOut, obj)
		}
	}
	NextToken = response.NextToken
	if NextToken != "" {
		log.Tracef("Next Token: %s", NextToken)
		_ = DescribeInstances(region, running, SpecifiedInstanceId, NextToken)
	}
	return DescribeInstancesOut
}

func ReturnInstancesList(region string, running bool, specifiedInstanceId string, ecsLsAllRegions bool) []Instances {
	var InstancesList []Instances
	var Instance []Instances
	if region == "all" {
		for _, j := range GetECSRegions(ecsLsAllRegions) {
			region := j.RegionId
			Instance = DescribeInstances(region, running, specifiedInstanceId, "")
			DescribeInstancesOut = nil
			for _, i := range Instance {
				InstancesList = append(InstancesList, i)
			}
		}
	} else {
		InstancesList = DescribeInstances(region, running, specifiedInstanceId, "")
	}
	return InstancesList
}

func PrintInstancesListRealTime(region string, running bool, specifiedInstanceId string, ecsLsAllRegions bool) {
	var InstanceCloudAssistantStatus string
	InstancesList := ReturnInstancesList(region, running, specifiedInstanceId, ecsLsAllRegions)
	var data1 = make([][]string, len(InstancesList))
	for i, o := range InstancesList {
		SN := strconv.Itoa(i + 1)
		data1[i] = []string{SN, o.InstanceId, o.InstanceName, o.OSName, o.OSType, o.Status, o.PublicIpAddress, o.RegionId}
	}
	var td1 = cloud.TableData{Header: header, Body: data1}
	if len(data1) == 0 {
		log.Info("未发现 ECS 资源，可能是因为当前访问密钥权限不够 (No ECS instances found, Probably because the current Access Key do not have enough permissions)")
	} else {
		Caption := "ECS 资源 (ECS resources)"
		cloud.PrintTable(td1, Caption)
		util.WriteTimestamp(TimestampType)
	}
	var data2 = make([][]string, len(InstancesList))
	for i, o := range InstancesList {
		SN := strconv.Itoa(i + 1)
		if o.CloudAssistantStatus == "True" {
			InstanceCloudAssistantStatus = "True"
		} else {
			InstanceCloudAssistantStatus = "False"
		}
		data2[i] = []string{SN, o.InstanceId, o.InstanceName, o.OSName, o.OSType, o.Status, o.PrivateIpAddress, o.PublicIpAddress, InstanceCloudAssistantStatus, o.RegionId}
	}
	var td2 = cloud.TableData{Header: header, Body: data2}
	cmdutil.WriteCacheFile(td2, "alibaba", "ecs", region, specifiedInstanceId)
}

func PrintInstancesListHistory(region string, running bool, specifiedInstanceId string) {
	cmdutil.PrintECSCacheFile(header, region, specifiedInstanceId, "alibaba", "ECS", running)
}

func PrintInstancesList(region string, running bool, specifiedInstanceId string, ecsFlushCache bool, ecsLsAllRegions bool) {
	if ecsFlushCache {
		PrintInstancesListRealTime(region, running, specifiedInstanceId, ecsLsAllRegions)
	} else {
		oldTimestamp := util.ReadTimestamp(TimestampType)
		if oldTimestamp == 0 {
			PrintInstancesListRealTime(region, running, specifiedInstanceId, ecsLsAllRegions)
		} else if util.IsFlushCache(oldTimestamp) {
			PrintInstancesListRealTime(region, running, specifiedInstanceId, ecsLsAllRegions)
		} else {
			util.TimeDifference(oldTimestamp)
			PrintInstancesListHistory(region, running, specifiedInstanceId)
		}
	}
}
