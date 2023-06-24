package identify

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	iamRegion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	log "github.com/sirupsen/logrus"
	"os"
)

func IAMClient(accessKey, secretKey, stsToken string) (client *iam.IamClient, errString string) {
	if accessKey == "" {
		log.Warnln("需要先配置访问密钥 (Access Key need to be configured first)")
		os.Exit(0)
		return nil, ""
	} else {
		// 判断是否已经配置了STS Token
		if stsToken == "" {
			auth := global.NewCredentialsBuilder().
				WithAk(accessKey).
				WithSk(secretKey).
				Build()
			defer func() {
				if err := recover(); err != nil {
					errString = err.(string)
				}
			}()
			client := iam.NewIamClient(
				iam.IamClientBuilder().
					WithRegion(iamRegion.ValueOf("cn-east-3")).
					WithCredential(auth).
					Build())
			return client, ""
		} else {
			auth := global.NewCredentialsBuilder().
				WithAk(accessKey).
				WithSk(secretKey).
				WithSecurityToken(stsToken).
				Build()

			client := iam.NewIamClient(
				iam.IamClientBuilder().
					WithRegion(iamRegion.ValueOf("cn-east-3")).
					WithCredential(auth).
					Build())
			return client, ""
		}
	}
}
