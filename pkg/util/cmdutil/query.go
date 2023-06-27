package cmdutil

import (
	log "github.com/sirupsen/logrus"
	"github.com/teamssix/cf/pkg/util/cmdutil/identify"
	"github.com/teamssix/cf/pkg/util/pubutil"
	"regexp"
	"strings"
)

func IdentifyProvider(AccessKeyId, SecretAccessKeyId, SessionToken string) pubutil.Provider {
	log.Debugf("\nAccessKeyId: %s\nSecretAccessKeyId: %s\nSessionToken: %s", AccessKeyId, SecretAccessKeyId, SessionToken)
	var provider pubutil.Provider
	switch {
	case (regexp.MustCompile("^LTAI[0-9a-zA-Z]{20}$").MatchString(AccessKeyId) || strings.HasPrefix(AccessKeyId, "STS")):
		// 正则已验证完全正确
		if SecretAccessKeyId == "" || identify.AlibabaIdentity(AccessKeyId, SecretAccessKeyId, SessionToken) {
			provider.CN = "阿里云"
			provider.EN = "Alibaba Cloud"
		}
	case regexp.MustCompile("^AKID[0-9a-zA-Z]{32}$").MatchString(AccessKeyId):
		// 正则已验证完全正确
		if SecretAccessKeyId == "" || identify.TencentIdentity(AccessKeyId, SecretAccessKeyId, SessionToken) {
			provider.CN = "腾讯云"
			provider.EN = "Tencent Cloud"
		}
	case (regexp.MustCompile("^[A-Z0-9]*$").MatchString(AccessKeyId) && (len(AccessKeyId) == 20 || len(AccessKeyId) == 40)):
		// 正则已验证完全正确
		if SecretAccessKeyId == "" || identify.HuaweiIdentity(AccessKeyId, SecretAccessKeyId, SessionToken) {
			provider.CN = "华为云"
			provider.EN = "Huawei Cloud"
		}
	case regexp.MustCompile("(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}").MatchString(AccessKeyId):
		// 正则源自 RExpository
		if SecretAccessKeyId == "" || identify.AwsIdentity(AccessKeyId, SecretAccessKeyId, SessionToken) {
			provider.CN = "亚马逊"
			provider.EN = "AWS"
		}
	case regexp.MustCompile("^ALTAK[0-9a-zA-Z]{21}$").MatchString(AccessKeyId):
		// 正则已验证完全正确
		if SecretAccessKeyId == "" || identify.BaiduIdentity(AccessKeyId, SecretAccessKeyId, SessionToken) {
			provider.CN = "百度云"
			provider.EN = "Baidu Cloud"
		}
	case (strings.HasPrefix(AccessKeyId, "AKL") || strings.HasPrefix(AccessKeyId, "AKTP")):
		if SecretAccessKeyId == "" || identify.HuoshanIdentity(AccessKeyId, SecretAccessKeyId) {
			provider.CN = "火山引擎"
			provider.EN = "Volcano Engine"
		}
	case (regexp.MustCompile(`^[a-zA-Z0-9-_]{40}$`).MatchString(AccessKeyId)):
		if SecretAccessKeyId == "" || identify.QiniuIdentity(AccessKeyId, SecretAccessKeyId) {
			provider.CN = "七牛云"
			provider.EN = "Qiniu Cloud"
		}
	case strings.HasPrefix(AccessKeyId, "UCLOUD"):
		provider.CN = "优刻得"
		provider.EN = "UCloud"
	case regexp.MustCompile("^AKLT[\\w-]{20}$").MatchString(AccessKeyId):
		// 正则已验证完全正确
		provider.CN = "金山云"
		provider.EN = "Kingsoft Cloud"
	case regexp.MustCompile("^JDC_[0-9A-Z]{28}$").MatchString(AccessKeyId):
		// 正则已验证完全正确
		provider.CN = "京东云"
		provider.EN = "JD Cloud"

	case regexp.MustCompile("AIza[0-9A-Za-z_\\-]{35}").MatchString(AccessKeyId):
		// 正则源自 RExpository
		provider.CN = "谷歌云"
		provider.EN = "GCP"
	default:
		provider.CN = ""
		provider.EN = ""
	}
	return provider
}
