package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"openai/bootstrap"
	"openai/internal/config"
	"openai/internal/handler"
	"os"
)

func init() {

}

func main() {
	r := bootstrap.New()

	// 微信消息处理
	r.POST("/wx", handler.ReceiveMsg)
	// 用于公众号自动验证
	r.GET("/wx", handler.WechatCheck)

	// 用于测试
	r.GET("/test", handler.Test)
	r.GET("/test_reply", handler.TestReplyToText)

	// Use webpage to display timeout or long message
	r.GET("/index", handler.Index)
	// For webpage
	r.GET("/reply", handler.GetReply)

	// 设置日志
	SetLog()

	fmt.Printf("启动服务，使用 curl 'http://127.0.0.1:%s/test?msg=你好' 测试一下吧\n", config.C.Http.Port)
	if err := http.ListenAndServe("127.0.0.1:"+config.C.Http.Port, r); err != nil {
		panic(err)
	}
}

func SetLog() {
	dir := "./log"
	path := dir + "/data.log"
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	fmt.Println("查看日志请使用 tail -f " + path)
}
