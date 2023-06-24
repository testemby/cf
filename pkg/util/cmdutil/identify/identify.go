package cmdutil

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/util/errutil"
	"strings"
)

func main() {
	request := sts.CreateGetCallerIdentityRequest()
	request.Scheme = "https"
	response, err := STSClient().GetCallerIdentity(request)
	errutil.HandleErr(err)
	accountArn := response.Arn
	var userName string
	if accountArn[len(accountArn)-4:] == "root" {
		userName = "root"
	} else {
		userName = strings.Split(accountArn, "/")[1]
	}
	log.Debugf("获得到当前凭证的用户名为 %s (The user name to get the current credentials is %s)", userName, userName)
	//return userName
}
