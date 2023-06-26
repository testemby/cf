package database

import (
	"github.com/teamssix/cf/pkg/util/pubutil"
)

func InsertRDSAccountsCache(provider string, DBInstanceId string, engine string, userName string, password string, region string) {
	var RDSAccountsCache []pubutil.RDSAccountsCache
	config := SelectConfigInUse(provider)
	AccessKeyId := config.AccessKeyId
	createTime := pubutil.CurrentTime()
	RDSAccountsCache = append(RDSAccountsCache, pubutil.RDSAccountsCache{
		Provider:     provider,
		AccessKeyId:  AccessKeyId,
		DBInstanceId: DBInstanceId,
		UserName:     userName,
		Password:     password,
		Engine:       engine,
		Region:       region,
		CreateTime:   createTime,
	})
	CacheDb.Create(&RDSAccountsCache)
}

func DeleteRDSAccountCache(provider string, DBInstanceId string) {
	var RDSAccountsCache []pubutil.RDSAccountsCache
	config := SelectConfigInUse(provider)
	CacheDb.Where("access_key_id = ? AND db_instance_id = ?", config.AccessKeyId, DBInstanceId).Delete(&RDSAccountsCache)
}

func SelectRDSAccountCache(provider string) []pubutil.RDSAccountsCache {
	var RDSAccountsCache []pubutil.RDSAccountsCache
	AccessKeyId := SelectConfigInUse(provider).AccessKeyId
	CacheDb.Where("access_key_id = ? COLLATE NOCASE", AccessKeyId).Find(&RDSAccountsCache)
	return RDSAccountsCache
}
