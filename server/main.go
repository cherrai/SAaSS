package main

import (
	"context"
	"os"

	conf "github.com/cherrai/SAaSS/config"
	mongodb "github.com/cherrai/SAaSS/db/mongo"
	redisdb "github.com/cherrai/SAaSS/db/redis"
	"github.com/cherrai/SAaSS/services/encryption"
	"github.com/cherrai/SAaSS/services/methods"
	"github.com/cherrai/SAaSS/services/socketio_service"

	"github.com/cherrai/nyanyago-utils/nlog"
	"github.com/cherrai/nyanyago-utils/nredis"
	sso "github.com/cherrai/saki-sso-go"

	// sfu "github.com/pion/ion-sfu/pkg/sfu"

	"github.com/go-redis/redis/v8"
)

var (
	log = nlog.New()
)

// 文件到期后根据时间进行删除 未做
func main() {
	nlog.SetPrefixTemplate("[{{Timer}}] [{{Type}}] [{{Date}}] [{{File}}]@{{Name}}")
	nlog.SetName("SAaSS")
	// 正式代码
	defer func() {
		log.Info("=========Error=========")
		if err := recover(); err != nil {
			log.Error(err)
		}
		log.Info("=========Error=========")
	}()

	configPath := ""
	for k, v := range os.Args {
		switch v {
		case "--config":
			if os.Args[k+1] != "" {
				configPath = os.Args[k+1]
			}
			break

		}
	}
	if configPath == "" {
		log.Error("Config file does not exist.")
		return
	}
	conf.GetConfig(configPath)
	conf.Init()

	// Connect to redis.
	redisdb.ConnectRedis(&redis.Options{
		Addr:     conf.Config.Redis.Addr,
		Password: conf.Config.Redis.Password, // no password set
		DB:       conf.Config.Redis.DB,       // use default DB
	})

	conf.Redisdb = nredis.New(context.Background(), &redis.Options{
		Addr:     conf.Config.Redis.Addr,
		Password: conf.Config.Redis.Password, // no password set
		DB:       conf.Config.Redis.DB,       // use default DB
	}, conf.BaseKey)
	conf.Redisdb.CreateKeys(conf.RedisCacheKeys)

	// Connect to mongodb.
	mongodb.ConnectMongoDB(conf.Config.Mongodb.Currentdb.Uri, conf.Config.Mongodb.Currentdb.Name)
	mongodb.ConnectMongoDB(conf.Config.Mongodb.Ssodb.Uri, conf.Config.Mongodb.Ssodb.Name)
	// SSO Init
	conf.SSO = sso.New(&sso.SakiSsoOptions{
		AppId:  conf.Config.SSO.AppId,
		AppKey: conf.Config.SSO.AppKey,
		Host:   conf.Config.SSO.Host,
		RedisOptions: &redis.Options{
			Addr:     conf.Config.Redis.Addr,
			Password: conf.Config.Redis.Password,
			DB:       conf.Config.Redis.DB,
		},
	})
	// Test()
	conf.EncryptionClient = encryption.New(encryption.NewOption{
		RedisClient:     redisdb.Rdb,
		RsaKeyDelayDays: 10,
		UserAesKeyMins:  10,
		TempDataMins:    1,
	})

	methods.DeleteFile()
	socketio_service.Init()

}