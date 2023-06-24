package identify

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	log "github.com/sirupsen/logrus"
	"os"
)

func AwsIamClient(accessKey, secretKey, stsToken string) *iam.IAM {
	if accessKey == "" {
		log.Warnln("需要先配置访问密钥 (Access Key need to be configured first)")
		os.Exit(0)
		return nil
	} else {

		cfg := &aws.Config{
			Region:      aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(accessKey, secretKey, stsToken),
		}
		sess := session.Must(session.NewSession(cfg))
		svc := iam.New(sess)
		log.Traceln("IAM Client 连接成功 (IAM Client connection successful)")
		return svc
	}
}
