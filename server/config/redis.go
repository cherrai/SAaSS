package conf

import (
	"time"

	"github.com/cherrai/nyanyago-utils/nredis"
)

var Redisdb *nredis.NRedis

var BaseKey = "meow-whisper"

var RedisCacheKeys = map[string]*nredis.RedisCacheKeysType{
	"AppToken": {
		Key:        "AppToken",
		Expiration: 5 * 60 * time.Second,
	},
	"ParentFolderId": {
		Key:        "ParentFolderId",
		Expiration: 5 * 60 * time.Second,
	},
	"UserToken": {
		Key:        "UserToken",
		Expiration: 5 * 60 * time.Second,
	},
	"TemporaryAccessToken": {
		Key:        "TemporaryAccessToken",
		Expiration: 30 * 60 * time.Second,
	},
}
