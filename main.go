package main

import (
	"log"
	"net/http"
	"openai/bootstrap"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/handler"
	"openai/internal/util"
	"os"
)

func init() {
	configLog()
	logUsefulInfo()
}

func main() {
	engine := bootstrap.New()
	// 公众号消息处理
	engine.POST("/talk", handler.ServeWechat)
	// 用于公众号自动验证
	engine.GET("/talk", handler.ServeWechat)
	// Provide reply content for the webpage
	engine.GET("/reply-stream", handler.GetReplyStream)
	engine.GET("/openid", handler.GetOpenId)
	if !util.AccountIsUncle() && util.EnvIsProd() {
		engine.POST("/transactions", handler.Transaction)
		engine.POST("/notify-transaction-result", handler.NotifyTransactionResult)
		engine.GET("/trade-result", handler.GetTradeResult)
	}
	handlerWithRequestLog := bootstrap.LogRequestHandler(engine)

	http.Handle("/", handlerWithRequestLog)

	log.Println("Server started in port " + config.C.Http.Port)
	err := http.ListenAndServe("127.0.0.1:"+config.C.Http.Port, nil)
	if err != nil {
		panic(err)
	}
}

func configLog() {
	if util.GetEnv() == constant.Dev {
		log.SetOutput(os.Stdout)
	} else {
		dir := "./log"
		path := dir + "/data.log"
		_, err := os.Stat(dir)
		if err != nil && os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				panic(err)
			}
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0755)
		if err != nil {
			panic(err)
		}
		log.SetOutput(file)
	}
}

func logUsefulInfo() {
	log.Printf("[Env] env: %s, account: %s", util.GetEnv(), util.GetAccount())
	log.Printf("[Redis] uncle db: %d, brother db: %d", config.C.Redis.UncleDB, config.C.Redis.BrotherDB)
}
