package aliecs

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/util/cmdutil"
	"github.com/teamssix/cf/pkg/util/database"
	"github.com/teamssix/cf/pkg/util/errutil"
	"github.com/teamssix/cf/pkg/util/pubutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ECSImageShare(aliyunAccount string, region string, specifiedInstanceId string) {
	if specifiedInstanceId == "all" {
		var (
			selectInstanceIdList []string
			selectInstanceId     string
			InstancesList        []Instances
			SN                   int
		)
		InstancesList = ReturnCacheInstanceList(specifiedInstanceId, region, "alibaba")
		if len(InstancesList) == 0 {
			log.Warnf("未发现实例，可以使用 --flushCache 刷新缓存后再试 (No instances found, You can use the --flushCache command to flush the cache and try again)")
			return
		} else if len(InstancesList) == 1 {
			specifiedInstanceId = InstancesList[0].InstanceId
			region = InstancesList[0].RegionId
		} else {
			for _, i := range InstancesList {
				SN = SN + 1
				selectInstanceIdList = append(selectInstanceIdList, fmt.Sprintf("%s %s (%s)", strconv.Itoa(SN), i.InstanceId, i.OSName))
			}
			sort.Strings(selectInstanceIdList)
			prompt := &survey.Select{
				Message: "选择一个实例 (Choose a instance): ",
				Options: selectInstanceIdList,
			}
			survey.AskOne(prompt, &selectInstanceId)
			for _, j := range InstancesList {
				if selectInstanceId != "all" {
					if selectInstanceId == fmt.Sprintf("%s (%s)", j.InstanceId, j.OSName) {
						InstancesList = nil
						InstancesList = append(InstancesList, j)
					}
				}
			}
			for _, v := range InstancesList {
				if strings.Contains(selectInstanceId, v.InstanceId) {
					specifiedInstanceId = v.InstanceId
					region = v.RegionId
				}
			}
		}
	}
	log.Infoln(fmt.Sprintf("即将为 %s 实例创建镜像 (Preparing to create an image for the instance %s.)", specifiedInstanceId, specifiedInstanceId))
	createImageRequest := ecs.CreateCreateImageRequest()
	createImageRequest.Scheme = "https"
	createImageRequest.InstanceId = specifiedInstanceId
	createImageRequest.QueryParams["Tag.1.value"] = "testMKzrHZyk"
	createImageRequest.QueryParams["Tag.1.Key"] = "testMKzrHZyk"
	createImageResponse, err := ECSClient(region).CreateImage(createImageRequest)
	if err != nil {
		errutil.HandleErr(err)
		return
	}

	ImageId := createImageResponse.ImageId
	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.Scheme = "https"
	describeImagesRequest.ImageOwnerAlias = "self"
	describeImagesRequest.QueryParams["waiter"] = "expr='Images.Image[0].Status' to=Available"
	describeImagesRequest.QueryParams["output"] = "cols=ImageId,Tags.Tag[0].TagValue,Status rows=Images.Image[]"

	log.Infoln("正在创建目标实例镜像，请耐心等待…… (Creating target instance image, please wait patiently...)")

	for {
		describeImagesResponse, _ := ECSClient(region).DescribeImages(describeImagesRequest)
		if len(describeImagesResponse.Images.Image) > 0 && string(describeImagesResponse.Images.Image[0].Tags.Tag[0].TagValue) == "testMKzrHZyk" {
			log.Infof(fmt.Sprintf("创建完成，正在共享此镜像到 %s 阿里云账户中 (Creation completed, currently sharing this image with the %s Alibaba Cloud account.)", aliyunAccount, aliyunAccount))
			break
		}
		time.Sleep(5 * time.Second)
	}

	modifyImageSharePermissionRequest := ecs.CreateModifyImageSharePermissionRequest()
	modifyImageSharePermissionRequest.Scheme = "https"
	modifyImageSharePermissionRequest.ImageId = ImageId
	modifyImageSharePermissionRequest.QueryParams["AddAccount.1"] = aliyunAccount
	_, err = ECSClient(region).ModifyImageSharePermission(modifyImageSharePermissionRequest)
	var status string
	if err != nil {
		aliyunAccount = ""
		status = "共享失败 (Sharing failed)"
		if strings.Contains(err.Error(), "Message: The specified Account does not yourself.") {
			log.Errorln("不能将镜像共享给自己，共享失败 (It is not possible to share an image with oneself. Sharing failed.)")
		} else {
			errutil.HandleErr(err)
		}
	} else {
		status = "共享成功 (Sharing successful)"
		log.Infoln("镜像共享成功，如果想取消共享镜像并删除所创建的镜像与快照，请使用 cf alibaba ecs imageShare cancel 命令 (Image sharing successful. If you want to cancel the shared image and delete the created image and snapshot, please use the \"cf alibaba ecs imageShare cancel\" command.)")
	}
	var ImageShareCache pubutil.ImageShareCache
	ImageShareCache.AccessKeyId = cmdutil.GetConfig("alibaba").AccessKeyId
	ImageShareCache.ImageId = ImageId
	ImageShareCache.InstanceId = specifiedInstanceId
	ImageShareCache.Provider = "alibaba"
	ImageShareCache.Region = region
	ImageShareCache.ShareAccountId = aliyunAccount
	ImageShareCache.Status = status
	ImageShareCache.Time = pubutil.CurrentTime()
	database.InsertImageShareCache(ImageShareCache)
}

func GetImageShare() {
	var (
		data            [][]string
		ImageShareCache []pubutil.ImageShareCache
		SN              int
	)
	ImageShareCache = database.SelectImageShareCache("alibaba")
	for _, v := range ImageShareCache {
		SN = SN + 1
		dataSingle := []string{strconv.Itoa(SN), v.InstanceId, v.ImageId, v.ShareAccountId, v.Status, v.Region, v.Time}
		data = append(data, dataSingle)
	}
	header = []string{"序号 (SN)", "实例 ID (Instance ID)", "镜像 ID (Image Name)", "共享帐号 ID (Share Account ID)", "状态 (Status)", "区域 ID (Region ID)", "时间 (Time)"}
	cmdutil.PrintTable(data, header, "Images Sahre")
}

func ImageDelete() {
	var (
		aliyunAccount       string
		ImageId             string
		ImageShareCache     []pubutil.ImageShareCache
		region              string
		specifiedInstanceId string
	)
	ImageShareCache = database.SelectImageShareCache("alibaba")
	if len(ImageShareCache) == 0 {
		log.Warnln("未找到共享镜像信息，无需删除 (Shared image information not found, no need for deletion.)")
		return
	} else if len(ImageShareCache) == 1 {
		aliyunAccount = ImageShareCache[0].ShareAccountId
		ImageId = ImageShareCache[0].ImageId
		region = ImageShareCache[0].Region
		specifiedInstanceId = ImageShareCache[0].InstanceId
	} else {
		var (
			selectImageIdList []string
			selectImageId     string
			SN                int
		)

		for _, i := range ImageShareCache {
			SN = SN + 1
			selectImageIdList = append(selectImageIdList, fmt.Sprintf("%s-%s-%s-%s)", strconv.Itoa(SN), i.InstanceId, i.ImageId, i.Region))
		}
		sort.Strings(selectImageIdList)
		prompt := &survey.Select{
			Message: "选择一个镜像 (Choose a image): ",
			Options: selectImageIdList,
		}
		survey.AskOne(prompt, &selectImageId)
		for _, v := range ImageShareCache {
			if strings.Contains(selectImageId, v.ImageId) {
				aliyunAccount = v.ShareAccountId
				ImageId = v.ImageId
				region = v.Region
				specifiedInstanceId = v.InstanceId
			}
		}
	}

	log.Debugln(fmt.Sprintf("已选择实例 ID 为 %s，镜像 ID 为 %s，共享帐号为 %s，区域为 %s (Instance ID selected: %s, Image ID: %s, Shared account: %s, Region: %s.)", specifiedInstanceId, ImageId, aliyunAccount, region))

	var isSure bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("确定取消共享并删除 %s 实例下的 %s 镜像与快照吗？(Are you sure you want to cancel the %s sharing and delete the image and snapshot under the %s instance?)", specifiedInstanceId, ImageId, ImageId, specifiedInstanceId),
		Default: true,
	}
	err := survey.AskOne(prompt, &isSure)
	errutil.HandleErr(err)
	if !isSure {
		log.Infoln("已中止操作 (The operation has been aborted.)")
		return
	}

	// 删除镜像
	modifyImageSharePermissionRequest := ecs.CreateModifyImageSharePermissionRequest()
	modifyImageSharePermissionRequest.Scheme = "https"
	modifyImageSharePermissionRequest.ImageId = ImageId

	modifyImageSharePermissionRequest.QueryParams["RemoveAccount.1"] = aliyunAccount
	ECSClient(region).ModifyImageSharePermission(modifyImageSharePermissionRequest)

	deleteImageRequest := ecs.CreateDeleteImageRequest()
	deleteImageRequest.ImageId = ImageId
	deleteImageRequest.QueryParams["Force"] = "true"
	ECSClient(region).DeleteImage(deleteImageRequest)

	// 删除快照
	describeSnapshotsRequest := ecs.CreateDescribeSnapshotsRequest()
	describeSnapshotsRequest.InstanceId = specifiedInstanceId
	describeSnapshotsRequest.SnapshotType = "user"
	describeSnapshotsRequest.QueryParams["output"] = "cols=SnapshotName,SnapshotId,LastModifiedTime rows=Snapshots.Snapshot[]"
	describeSnapshotsResponse, err := ECSClient(region).DescribeSnapshots(describeSnapshotsRequest)
	errutil.HandleErr(err)

	var (
		lastModifiedTime time.Time
		snapShotId       string
	)

	for _, snapshot := range describeSnapshotsResponse.Snapshots.Snapshot {
		snapshotTime, _ := time.Parse(time.RFC3339, snapshot.LastModifiedTime)
		if snapshotTime.After(lastModifiedTime) {
			lastModifiedTime = snapshotTime
			snapShotId = snapshot.SnapshotId
		}
	}

	deleteSnapshotRequest := ecs.CreateDeleteSnapshotRequest()
	deleteSnapshotRequest.SnapshotId = snapShotId
	ECSClient(region).DeleteSnapshot(deleteSnapshotRequest)

	// 删除本地缓存信息
	database.DeleteImageShareCache(ImageId)
	log.Infoln("已取消共享并已删除镜像与快照 (Sharing has been canceled, and the image and snapshot have been deleted.)")
}
