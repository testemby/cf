package database

import "github.com/teamssix/cf/pkg/util/pubutil"

func InsertImageShareCache(ImageShareCache pubutil.ImageShareCache) {
	CacheDb.Create(&ImageShareCache)
}

func SelectImageShareCache(provider string) []pubutil.ImageShareCache {
	var ImageShareCache []pubutil.ImageShareCache
	AccessKeyId := SelectConfigInUse(provider).AccessKeyId
	CacheDb.Where("access_key_id = ?", AccessKeyId).Find(&ImageShareCache)
	return ImageShareCache
}

func DeleteImageShareCache(ImageID string) {
	var ImageShareCache []pubutil.ImageShareCache
	CacheDb.Where("image_id = ?", ImageID).Delete(&ImageShareCache)
}
