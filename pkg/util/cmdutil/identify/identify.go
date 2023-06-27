package identify

import (
	"encoding/json"
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
	"strings"
)

func AlibabaIdentity(accessKey, secretKey, stsToken string) bool {
	request := alibabaSts.CreateGetCallerIdentityRequest()
	request.Scheme = "https"
	_, err := AlibabaSTSClient(accessKey, secretKey, stsToken).GetCallerIdentity(request)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidAccessKeyId.NotFound") {
			log.Debugln("AccessKey 不属于阿里云 (AccessKey does not belong to Aliyun)")
			return false
		} else if strings.Contains(err.Error(), "SignatureDoesNotMatch") {
			log.Debugln("AccessKey 属于阿里云 (AccessKey belongs to Aliyun)")
			log.Debugln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
			return true
		}
	}
	log.Debugln("AccessKey 属于阿里云 (AccessKey belongs to Aliyun)")
	return true
}

func HuaweiIdentity(accessKey, secretKey, stsToken string) bool {
	showPermanentAccessKeyRequestContent := &iamModel.ShowPermanentAccessKeyRequest{}
	showPermanentAccessKeyRequestContent.AccessKey = accessKey
	_, err := IAMClient(accessKey, secretKey, stsToken)
	if err != "" {
		if strings.Contains(err, "ak "+accessKey+" not exist") {
			log.Debugln("AccessKey 不属于华为云 (AccessKey does not belong to Huawei Cloud)")
			return false
		} else if strings.Contains(err, "verify aksk signature fail") {
			log.Debugln("AccessKey 属于华为云 (AccessKey belongs to Huawei Cloud)")
			log.Debugln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
			return true
		}
	} else {
		log.Debugln("AccessKey 属于华为云 (AccessKey belongs to Huawei Cloud)")
		return true
	}
	return true
}

func TencentIdentity(accessKey, secretKey, stsToken string) bool {
	request := tencentSts.NewGetCallerIdentityRequest()
	_, err := TencentSTSClient(accessKey, secretKey, stsToken).GetCallerIdentity(request)
	if err != nil {
		if strings.Contains(err.Error(), "AuthFailure.SecretIdNotFound") {
			log.Debugln("AccessKey 不属于腾讯云 (AccessKey does not belong to Tencent Cloud)")
			return false
		} else if strings.Contains(err.Error(), "AuthFailure.SignatureFailure") {
			log.Debugln("AccessKey 属于腾讯云 (AccessKey belongs to Tencent Cloud)")
			log.Debugln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
			return true
		}
	}
	log.Debugln("AccessKey 属于腾讯云 (AccessKey belongs to Tencent Cloud)")
	return true
}

func AwsIdentity(accessKey, secretKey, stsToken string) bool {
	svc := AwsIamClient(accessKey, secretKey, stsToken)
	input := &iam.GetAccountSummaryInput{}
	_, err := svc.GetAccountSummary(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if strings.Contains(aerr.Code(), "InvalidClientTokenId") {
				log.Debugln("AccessKey 不属于 AWS (AccessKey does not belong to AWS)")
				return false
			} else if strings.Contains(aerr.Code(), "SignatureDoesNotMatch") {
				log.Debugln("AccessKey 属于 AWS (AccessKey belongs to AWS)")
				log.Debugln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
				return true
			}
		}
	}
	log.Debugln("AccessKey 属于 AWS (AccessKey belongs to AWS)")
	return true
}

func QiniuIdentity(accessKey, secretKey string) bool {
	bucket := "crossfire"
	mac := auth.New(accessKey, secretKey)

	cfg := storage.Config{
		UseHTTPS: false,
	}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	prefix, delimiter, marker := "", "", ""
	_, err := bucketManager.ListBucket(bucket, prefix, delimiter, marker)
	if err != nil {
		if strings.Contains(err.Error(), "query region error, app/accesskey is not found") {
			log.Debugln("AccessKey 不属于七牛云 (AccessKey does not belong to Qiniu Cloud)")
			return false
		} else if strings.Contains(err.Error(), "query region error, no such bucket") {
			log.Debugln("AccessKey 属于七牛云 (AccessKey belongs to Qiniu Cloud)")
			log.Debugln("AccessKeySecret 似乎输入有误或者 BucketName crossfire 不存在 (AccessKeySecret appears to be incorrect or BucketName does not exist)")
			return true
		} else if strings.Contains(err.Error(), "bad token") {
			log.Debugln("AccessKey 属于七牛云 (AccessKey belongs to Qiniu Cloud)")
			log.Debugln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect")
			return true
		}
	}
	log.Debugln("AccessKey 属于七牛云 (AccessKey belongs to Qiniu Cloud)")
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
		log.Debugln("AccessKey 不属于火山引擎 (AccessKey does not belong to Huoshan Engine)")
		return false
	} else if strings.Contains(string(b), "SignatureDoesNotMatch") {
		log.Debugln("AccessKey 属于火山引擎 (AccessKey belongs to Huoshan Engine)")
		log.Debugln("AccessKeySecret 似乎输入有误 (AccessKeySecret appears to be incorrect)")
		return true
	}
	log.Debugln("AccessKey 属于火山引擎 (AccessKey belongs to Huoshan Engine)")
	return true
}

func BaiduIdentity(accessKey, secretKey, stsToken string) bool {
	client := BCEIAMClient(accessKey, secretKey, stsToken)
	_, err := client.ListUser()
	if err != nil {
		if strings.Contains(err.Error(), "AccessDenied") {
			log.Debugln("AccessKey 不属于百度云 (AccessKey does not belong to Baidu Cloud)")
			return false
		}
	}
	log.Debugln("AccessKey 属于百度云 (AccessKey belongs to Baidu Cloud)")
	return true
}
