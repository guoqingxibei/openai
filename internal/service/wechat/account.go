package wechat

import (
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/sirupsen/logrus"
	"openai/internal/config"
	"openai/internal/util"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
}

func GetAccount() *officialaccount.OfficialAccount {
	wc := wechat.NewWechat()
	db := config.C.Redis.BrotherDB
	if util.AccountIsUncle() {
		db = config.C.Redis.UncleDB
	}
	redisOpts := &cache.RedisOpts{
		Host:        config.C.Redis.Addr,
		Database:    db,
		MaxActive:   10,
		MaxIdle:     10,
		IdleTimeout: 60, //second
	}
	redisCache := cache.NewRedis(ctx, redisOpts)
	cfg := &offConfig.Config{
		AppID:     config.C.Wechat.AppId,
		AppSecret: config.C.Wechat.AppSecret,
		Token:     config.C.Wechat.Token,
		//EncodingAESKey: "xxxx",
		Cache: redisCache,
	}
	return wc.GetOfficialAccount(cfg)
}
