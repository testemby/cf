package cmdutil

import (
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"fmt"
	"strings"
	"github.com/teamssix/cf/pkg/cloud"
	"regexp"
)


func QueryAccessKey() {
	log.Infoln("请输入Accesskey进行查询，输入exit将退出此功能")
	var accesskey string
	var result string
	for {
	    fmt.Println()
	    prompt := &survey.Input{
	        Message: "AccessKey:",
	    }
	    survey.AskOne(prompt, &accesskey)
	    fmt.Println()
	    if accesskey == "exit" {
	        break
	    }

	    switch {
		    case (strings.HasPrefix(accesskey, "AKIA") || strings.HasPrefix(accesskey, "ASIA")): //已验证完全正确
		        result = "AWS"
		    case (strings.HasPrefix(accesskey, "GOOG") || strings.HasPrefix(accesskey, "AIza")):
		        result = "谷歌云 (Google Cloud)"
		    case (strings.HasPrefix(accesskey, "LTAI") || strings.HasPrefix(accesskey, "STS")): //已验证完全正确
		        result = "阿里云 (Aliyun)"
		    case strings.HasPrefix(accesskey, "AKID"): //已验证完全正确
		        result = "腾讯云 (Tencent Cloud)"
		    case strings.HasPrefix(accesskey, "ALTA"): //已验证完全正确
		        result = "百度云 (Baidu Cloud)"
		    case strings.HasPrefix(accesskey, "UCLOUD"):
		        result = "UCloud"
			case regexp.MustCompile("^AKLT[\\w-]{20}$").MatchString(accesskey): //已验证完成正确
			    result = "金山云 (Kingsoft Cloud)"
			case strings.HasPrefix(accesskey, "JDC_"): //已验证完全正确
			    result = "京东云 (Jingdong Cloud)"
			case (strings.HasPrefix(accesskey, "AKL") || strings.HasPrefix(accesskey, "AKTP")):
			    result = "火山引擎 (Volcano Engine)"
			case (regexp.MustCompile("^[A-Z0-9]*$").MatchString(accesskey) && (len(accesskey) == 20 || len(accesskey) == 40)): //已验证完全正确
        		result = "华为云 (Huawei Cloud)"
		    default:
		        result = "未知AK (Unkown)"
	    }
	    data := [][]string{
	        {accesskey, result},
	    }
	    var header = []string{"AK (AccessKey)", "云厂商 (CloudProvider)"}
	    var td = cloud.TableData{Header: header, Body: data}
	    cloud.PrintTable(td, "")
	}
}