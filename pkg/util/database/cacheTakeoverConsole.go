package database

import (
	"github.com/teamssix/cf/pkg/util/pubutil"
	"time"
)

func InsertTakeoverConsoleCache(provider string, primaryAccountId string, userName string, password string, loginURL string, consoleAccessKeyId string, consoleAccessKeySecret string) {
	var TakeoverConsoleCache []pubutil.TakeoverConsoleCache
	DeleteTakeoverConsoleCache(provider)
	config := SelectConfigInUse(provider)
	accessKeyId := config.AccessKeyId
	createTime := time.Now().Format("2006-01-02 15:04:05")
	TakeoverConsoleCache = append(TakeoverConsoleCache, pubutil.TakeoverConsoleCache{
		Provider:               provider,
		AccessKeyId:            accessKeyId,
		PrimaryAccountId:       primaryAccountId,
		UserName:               userName,
		Password:               password,
		LoginUrl:               loginURL,
		ConsoleAccessKeyId:     consoleAccessKeyId,
		ConsoleAccessKeySecret: consoleAccessKeySecret,
		CreateTime:             createTime,
	})
	CacheDb.Create(&TakeoverConsoleCache)
}

func DeleteTakeoverConsoleCache(provider string) {
	var TakeoverConsoleCache []pubutil.TakeoverConsoleCache
	config := SelectConfigInUse(provider)
	CacheDb.Where("access_key_id = ?", config.AccessKeyId).Delete(&TakeoverConsoleCache)
}

func SelectTakeoverConsoleCache(provider string) []pubutil.TakeoverConsoleCache {
	var TakeoverConsoleCache []pubutil.TakeoverConsoleCache
	AccessKeyId := SelectConfigInUse(provider).AccessKeyId
	CacheDb.Where("access_key_id = ? COLLATE NOCASE", AccessKeyId).Find(&TakeoverConsoleCache)
	return TakeoverConsoleCache
}
