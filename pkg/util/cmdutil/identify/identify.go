package identify

import (
	"encoding/json"
	"github.com/AlecAivazis/survey/v2"
	alibabaSts "github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	iamModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
	log "github.com/sirupsen/logrus"
	tencentSts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
	"github.com/volcengine/volc-sdk-golang/service/visual"
	"net/url"
	"os"
	"strings"
)

type credential struct {
	AccessKeyId     string
	AccessKeySecret string
	STSToken        string
}

func IdentifyAccessKey() {
	cred := InputAccessKey()
	log.Infoln("AccessKeyId:", cred.AccessKeyId)
	log.Infoln("AccessKeySecret:", cred.AccessKeySecret)
	log.Infoln("STSToken:", cred.STSToken)
	if AliyunIdentity(cred.AccessKeyId, cred.AccessKeySecret, cred.STSToken) {
		os.Exit(0)
	} else if HuaweiIdentity(cred.AccessKeyId, cred.AccessKeySecret, cred.STSToken) {
		os.Exit(0)
	} else if TencentIdentity(cred.AccessKeyId, cred.AccessKeySecret, cred.STSToken) {
		os.Exit(0)
	} else if AwsIdentity(cred.AccessKeyId, cred.AccessKeySecret, cred.STSToken) {
		os.Exit(0)
	} else if QiniuIdentity(cred.AccessKeyId, cred.AccessKeySecret) {
		os.Exit(0)
	} else if HuoshanIdentity(cred.AccessKeyId, cred.AccessKeySecret) {
		os.Exit(0)
	} else {
		log.Errorln("AccessKey 无法识别 (AccessKey cannot be identified)")
	}
}

func AliyunIdentity(accessKey, secretKey, stsToken string) bool {
	request := alibabaSts.CreateGetCallerIdentityRequest()
	request.Scheme = "https"
	_, err := AlibabaSTSClient(accessKey, secretKey, stsToken).GetCallerIdentity(request)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidAccessKeyId.NotFound") {
			log.Errorln("AccessKey 不属于阿里云 (AccessKey does not belong to Aliyun)")
			return false
		} else if strings.Contains(err.Error(), "SignatureDoesNotMatch") {
			log.Infoln("AccessKey 属于阿里云 (AccessKey belongs to Aliyun)")
			log.Errorln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
			return true
		}
	}
	log.Infoln("AccessKey 属于阿里云 (AccessKey belongs to Aliyun)")
	return true
}

func HuaweiIdentity(accessKey, secretKey, stsToken string) bool {
	showPermanentAccessKeyRequestContent := &iamModel.ShowPermanentAccessKeyRequest{}
	showPermanentAccessKeyRequestContent.AccessKey = accessKey
	_, err := IAMClient(accessKey, secretKey, stsToken)
	if err != "" {
		if strings.Contains(err, "ak "+accessKey+" not exist") {
			log.Errorln("AccessKey 不属于华为云 (AccessKey does not belong to Huawei Cloud)")
			return false
		} else if strings.Contains(err, "verify aksk signature fail") {
			log.Infoln("AccessKey 属于华为云 (AccessKey belongs to Huawei Cloud)")
			log.Errorln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
			return true
		}
	} else {
		log.Infoln("AccessKey 属于华为云 (AccessKey belongs to Huawei Cloud)")
		return true
	}
	return true
}

func TencentIdentity(accessKey, secretKey, stsToken string) bool {
	request := tencentSts.NewGetCallerIdentityRequest()
	_, err := TencentSTSClient(accessKey, secretKey, stsToken).GetCallerIdentity(request)
	if err != nil {
		if strings.Contains(err.Error(), "AuthFailure.SecretIdNotFound") {
			log.Errorln("AccessKey 不属于腾讯云 (AccessKey does not belong to Tencent Cloud)")
			return false
		} else if strings.Contains(err.Error(), "AuthFailure.SignatureFailure") {
			log.Infoln("AccessKey 属于腾讯云 (AccessKey belongs to Tencent Cloud)")
			log.Errorln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
			return true
		}
	}
	log.Infoln("AccessKey 属于腾讯云 (AccessKey belongs to Tencent Cloud)")
	return true
}

func AwsIdentity(accessKey, secretKey, stsToken string) bool {
	svc := AwsIamClient(accessKey, secretKey, stsToken)
	input := &iam.GetAccountSummaryInput{}

	_, err := svc.GetAccountSummary(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if strings.Contains(aerr.Code(), "InvalidClientTokenId") {
				log.Errorln("AccessKey 不属于 AWS (AccessKey does not belong to AWS)")
				return false
			} else if strings.Contains(aerr.Code(), "SignatureDoesNotMatch") {
				log.Infoln("AccessKey 属于 AWS (AccessKey belongs to AWS)")
				log.Errorln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
				return true
			}
		}

	}
	log.Infoln("AccessKey 属于 AWS (AccessKey belongs to AWS)")
	return true
}

func QiniuIdentity(accessKey, secretKey string) bool {
	bucket := "teamssix"
	mac := auth.New(accessKey, secretKey)

	cfg := storage.Config{
		UseHTTPS: false,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	prefix, delimiter, marker := "", "", ""
	_, err := bucketManager.ListBucket(bucket, prefix, delimiter, marker)
	if err != nil {
		if strings.Contains(err.Error(), "query region error, app/accesskey is not found") {
			log.Errorln("AccessKey 不属于七牛云 (AccessKey does not belong to Qiniu Cloud)")
			return false
		} else if strings.Contains(err.Error(), "query region error, no such bucket") {
			log.Infoln("AccessKey 属于七牛云 (AccessKey belongs to Qiniu Cloud)")
			log.Errorln("AccessKeySecret 似乎输入有误或者 BucketName teamssix 不存在 (AccessKeySecret appears to be incorrect or BucketName does not exist)")
			return true
		} else if strings.Contains(err.Error(), "bad token") {
			log.Infoln("AccessKey 属于七牛云 (AccessKey belongs to Qiniu Cloud)")
			log.Errorln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect")
			return true
		}
	}
	log.Infoln("AccessKey 属于七牛云 (AccessKey belongs to Qiniu Cloud)")
	return true
}

func HuoshanIdentity(accessKey, secretKey string) bool {
	visual.DefaultInstance.Client.SetAccessKey(accessKey)
	visual.DefaultInstance.Client.SetSecretKey(secretKey)

	form := url.Values{}
	form.Add("image_base64", "")

	resp, _, _ := visual.DefaultInstance.BankCard(form)
	b, _ := json.Marshal(resp)
	if strings.Contains(string(b), "InvalidAccessKey") {
		log.Errorln("AccessKey 不属于火山引擎 (AccessKey does not belong to Huoshan Engine)")
	} else if strings.Contains(string(b), "SignatureDoesNotMatch") {
		log.Infoln("AccessKey 属于火山引擎 (AccessKey belongs to Huoshan Engine)")
		log.Errorln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
		return true
	}
	log.Infoln("AccessKey 属于火山引擎 (AccessKey belongs to Huoshan Engine)")
	return true

}

func InputAccessKey() credential {
	var qs = []*survey.Question{
		{
			Name:   "AccessKeyId",
			Prompt: &survey.Input{Message: "输入访问密钥 ID (Input Access Key Id) (必须 Required):"},
			Validate: func(val interface{}) error {
				str := val.(string)
				if len(strings.TrimSpace(str)) < 7 {
					log.Warnln("访问凭证似乎输入有误 (This access credential appears to be incorrect.)")
				}
				return nil
			},
		},
		{
			Name:   "AccessKeySecret",
			Prompt: &survey.Password{Message: "输入访问密钥密钥 (Input Access Key Secret) (必须 Required):"},
			Validate: func(val interface{}) error {
				str := val.(string)
				if len(strings.TrimSpace(str)) < 7 {
					log.Warnln("访问凭证似乎输入有误 (This access credential appears to be incorrect.)")
				}
				return nil
			},
		},
		{
			Name:   "STSToken",
			Prompt: &survey.Input{Message: "输入临时凭证的 Token (Input STS Token) (可选 Optional):"},
		},
	}
	cred := credential{}
	err := survey.Ask(qs, &cred)
	if err != nil {
		log.Fatal(err)
	}
	return cred
}
