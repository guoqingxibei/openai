package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"openai/bootstrap"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/handler"
	"openai/internal/util"
	"os"
	"path/filepath"
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

	slog.Info("Server started in port " + config.C.Http.Port)
	err := http.ListenAndServe("127.0.0.1:"+config.C.Http.Port, nil)
	if err != nil {
		panic(err)
	}
}

func configLog() {
	if util.GetEnv() != constant.Dev {
		logPath := "logs/app.log"
		err := os.MkdirAll(filepath.Dir(logPath), os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating log file directory: %v\n", err)
			return
		}

		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Error opening log file: %v\n", err)
			return
		}
		log.SetOutput(logFile)
	}
}

func logUsefulInfo() {
	slog.Info("[Env]", "env", util.GetEnv(), "account", util.GetAccount())
	slog.Info("[Redis]", "UncleDB", config.C.Redis.UncleDB, "BrotherDB", config.C.Redis.BrotherDB)
}
