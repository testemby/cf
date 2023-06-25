package identify

import (
	"os"

	"github.com/teamssix/cf/pkg/util/errutil"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	log "github.com/sirupsen/logrus"
)

func AlibabaSTSClient(accessKey, secretKey, stsToken string) *sts.Client {
	// 判断是否已经配置了访问密钥
	if accessKey == "" {
		log.Warnln("需要先配置访问密钥 (Access Key need to be configured first)")
		os.Exit(0)
		return nil
	} else {
		// 判断是否已经配置了 STS Token
		config := sdk.NewConfig()
		if stsToken == "" {
			credential := credentials.NewAccessKeyCredential(accessKey, secretKey)
			client, err := sts.NewClientWithOptions("cn-beijing", config, credential)
			errutil.HandleErr(err)
			if err == nil {
				log.Traceln("RAM Client 连接成功 (RAM Client connection successful)")
			} else {
				log.Fatal(err)
			}
			return client
		} else {
			// 使用 STS Token 连接
			credential := credentials.NewStsTokenCredential(accessKey, secretKey, stsToken)
			client, err := sts.NewClientWithOptions("cn-beijing", config, credential)
			errutil.HandleErr(err)
			if err == nil {
				log.Traceln("RAM Client 连接成功 (RAM Client connection successful)")
			} else {
				log.Fatal(err)
			}
			return client
		}
	}
}
