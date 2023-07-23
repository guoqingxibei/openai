package main

import (
	"log"
	"net/http"
	"openai/bootstrap"
	"openai/internal/config"
	"openai/internal/handler"
	"os"
)

var (
	env = os.Getenv("GO_ENV")
)

func main() {
	ConfigLog()

	engine := bootstrap.New()
	// 公众号消息处理
	engine.POST("/talk", handler.Talk)
	// 用于公众号自动验证
	engine.GET("/talk", handler.Check)
	// Provide reply content for the webpage
	engine.GET("/reply-stream", handler.GetReplyStream)
	engine.GET("/openid", handler.GetOpenId)
	engine.POST("/transactions", handler.Transaction)
	engine.POST("/notify-transaction-result", handler.NotifyTransactionResult)

	handlerWithRequestLog := bootstrap.LogRequestHandler(engine)
	http.Handle("/talk", handlerWithRequestLog)
	http.Handle("/reply-stream", handlerWithRequestLog)
	http.Handle("/openid", handlerWithRequestLog)
	http.Handle("/transactions", handlerWithRequestLog)
	http.Handle("/notify-transaction-result", handlerWithRequestLog)
	http.Handle("/answer/", http.StripPrefix("/answer", http.FileServer(http.Dir("./public"))))
	http.Handle("/images/", http.FileServer(http.Dir("./public")))
	http.Handle("/shop/", http.FileServer(http.Dir("./public/")))

	log.Println("Server started in port " + config.C.Http.Port)
	err := http.ListenAndServe("127.0.0.1:"+config.C.Http.Port, nil)
	if err != nil {
		panic(err)
	}
}

func ConfigLog() {
	if env == "dev" {
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
