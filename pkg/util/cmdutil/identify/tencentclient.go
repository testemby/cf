package identify

import (
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/util/errutil"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
	"os"
)

func TencentSTSClient(accessKey, secretKey, stsToken string) *sts.Client {
	if accessKey == "" {
		log.Warnln("需要先配置访问密钥 (Access Key need to be configured first)")
		os.Exit(0)
		return nil
	} else {
		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = "sts.tencentcloudapi.com"
		if stsToken == "" {
			credential := common.NewCredential(accessKey, secretKey)
			client, err := sts.NewClient(credential, "ap-beijing", cpf)
			errutil.HandleErr(err)
			if err == nil {
				log.Traceln("STS Client 连接成功 (STS Client connection successful)")
			} else {
				log.Fatal(err)
			}
			return client
		} else {
			credential := common.NewTokenCredential(accessKey, secretKey, stsToken)
			client, err := sts.NewClient(credential, "ap-beijing", cpf)
			errutil.HandleErr(err)
			if err == nil {
				log.Infoln("STS Client 连接成功 (STS Client connection successful)")
			} else {
				log.Fatal(err)
			}

			return client
		}
	}
}
