package main

import (
	"io"
	"log"
	"net/http"
	"openai/bootstrap"
	"openai/internal/config"
	"openai/internal/handler"
	"os"
)

func main() {
	engine := bootstrap.New()

	// 公众号消息处理
	engine.POST("/talk", handler.Talk)
	// 用于公众号自动验证
	engine.GET("/talk", handler.Check)
	// Use webpage to display timeout or long message
	engine.GET("/index", handler.Index)
	// Provide reply content for the webpage
	engine.GET("/reply", handler.GetReply)

	SetLog()
	if err := http.ListenAndServe("127.0.0.1:"+config.C.Http.Port, engine); err != nil {
		panic(err)
	}
}

func SetLog() {
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
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
}
