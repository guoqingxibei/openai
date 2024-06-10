package main

import (
	"log/slog"
	"net/http"
	"openai/bootstrap"
	"openai/internal/config"
	"openai/internal/handler"
	"openai/internal/util"
)

func init() {
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

	slog.Info("Server started in port " + config.C.Http.Port)
	err := http.ListenAndServe("127.0.0.1:"+config.C.Http.Port, nil)
	if err != nil {
		panic(err)
	}
}

func logUsefulInfo() {
	slog.Info("[Env]", "env", util.GetEnv(), "account", util.GetAccount())
	slog.Info("[Redis]", "UncleDB", config.C.Redis.UncleDB, "BrotherDB", config.C.Redis.BrotherDB)
}
