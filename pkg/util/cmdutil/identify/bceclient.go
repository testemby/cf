package identify

import (
	"github.com/baidubce/bce-sdk-go/auth"
	"github.com/baidubce/bce-sdk-go/services/iam"
	log "github.com/sirupsen/logrus"
	"os"
)

func BCEIAMClient(accessKey, secretKey, stsToken string) *iam.Client {
	// 用户的Access Key ID和Secret Access Key
	if accessKey == "" {
		log.Warnln("需要先配置访问密钥 (Access Key need to be configured first)")
		os.Exit(0)
		return nil
	} else {
		if stsToken == "" {
			iamClient, err := iam.NewClient(accessKey, secretKey)
			if err == nil {
				log.Traceln("IAM Client 连接成功 (IAM Client connection successful)")
			} else {
				log.Fatal(err)
			}
			return iamClient
		} else {
			stsCredential, _ := auth.NewSessionBceCredentials(
				accessKey,
				secretKey,
				stsToken)
			iamClient, err := iam.NewClient(accessKey, secretKey)
			if err == nil {
				log.Traceln("IAM Client 连接成功 (IAM Client connection successful)")
			} else {
				log.Fatal(err)
			}
			iamClient.Config.Credentials = stsCredential
			return iamClient
		}
	}
}
