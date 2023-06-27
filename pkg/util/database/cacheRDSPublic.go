package database

import (
	"github.com/teamssix/cf/pkg/util/pubutil"
)

func InsertRDSPublicCache(provider string, DBInstanceId string, engine string, IPAddress string, ConnectionAddress string, Port string, region string) {
	var RDSPublicCache []pubutil.RDSPublicCache
	config := SelectConfigInUse(provider)
	AccessKeyId := config.AccessKeyId
	createTime := pubutil.CurrentTime()
	RDSPublicCache = append(RDSPublicCache, pubutil.RDSPublicCache{
		AccessKeyId:       AccessKeyId,
		DBInstanceId:      DBInstanceId,
		Provider:          provider,
		Region:            region,
		IPAddress:         IPAddress,
		ConnectionAddress: ConnectionAddress,
		Port:              Port,
		Engine:            engine,
		CreateTime:        createTime,
	})
	CacheDb.Create(&RDSPublicCache)
}

func DeleteRDSPublicCache(provider string, DBInstanceId string) {
	var RDSPublicCache []pubutil.RDSPublicCache
	config := SelectConfigInUse(provider)
	CacheDb.Where("access_key_id = ? AND db_instance_id = ?", config.AccessKeyId, DBInstanceId).Delete(&RDSPublicCache)
}

func SelectRDSPublicCache(provider string) []pubutil.RDSPublicCache {
	var RDSPublicCache []pubutil.RDSPublicCache
	AccessKeyId := SelectConfigInUse(provider).AccessKeyId
	CacheDb.Where("access_key_id = ? COLLATE NOCASE", AccessKeyId).Find(&RDSPublicCache)
	return RDSPublicCache
}
