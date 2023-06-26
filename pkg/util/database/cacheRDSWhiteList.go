package database

import "github.com/teamssix/cf/pkg/util/pubutil"

func InsertRDSWhiteListCache(provider string, DBInstanceId string, engine string, IPArrayName string, IPType string, IPList string, WhiteList string, region string) {
	var RDSWhiteListCache []pubutil.RDSWhiteListCache
	config := SelectConfigInUse(provider)
	AccessKeyId := config.AccessKeyId
	createTime := pubutil.CurrentTime()
	RDSWhiteListCache = append(RDSWhiteListCache, pubutil.RDSWhiteListCache{
		Provider:     provider,
		AccessKeyId:  AccessKeyId,
		DBInstanceId: DBInstanceId,
		IPArrayName:  IPArrayName,
		IPType:       IPType,
		WhiteList:    WhiteList,
		IPList:       IPList,
		Engine:       engine,
		Region:       region,
		CreateTime:   createTime,
	})
	CacheDb.Create(&RDSWhiteListCache)
}

func DeleteRDSWhiteListCache(provider string, DBInstanceId string, WhiteList string) {
	var RDSWhiteListCache []pubutil.RDSWhiteListCache
	config := SelectConfigInUse(provider)
	CacheDb.Where("access_key_id = ? AND db_instance_id = ? AND white_list = ?", config.AccessKeyId, DBInstanceId, WhiteList).Delete(&RDSWhiteListCache)
}

func SelectRDSWhiteListCache(provider string) []pubutil.RDSWhiteListCache {
	var RDSWhiteListCache []pubutil.RDSWhiteListCache
	AccessKeyId := SelectConfigInUse(provider).AccessKeyId
	CacheDb.Where("access_key_id = ? COLLATE NOCASE", AccessKeyId).Find(&RDSWhiteListCache)
	return RDSWhiteListCache
}
