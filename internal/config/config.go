package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type conf struct {
	Http struct {
		Port  string `json:"port"`
		Proxy string `json:"proxy"`
	} `json:"http"`
	Ohmygpt struct {
		Key      string `json:"key"`
		BaseURL  string `json:"base_url"`
		UseAzure bool   `json:"use_azure"`
	} `json:"ohmygpt"`
	GptApiUs struct {
		Key     string `json:"key"`
		BaseURL string `json:"base_url"`
	} `json:"gpt_api_us"`
	Openai struct {
		Key     string `json:"key"`
		BaseURL string `json:"base_url"`
	} `json:"openai"`
	Wechat struct {
		Token            string `json:"token"`
		MessageUrlPrefix string `json:"message_url_prefix"`
		AppId            string `json:"app_id"`
		AppSecret        string `json:"app_secret"`
		MchId            string `json:"mch_id"`
		SerialNo         string `json:"serial_no"`
		APIv3Key         string `json:"api_v3_key"`
		PrivateKey       string `json:"private_key"`
		NotifyUrl        string `json:"notify_url"`
	} `json:"wechat"`
	Redis struct {
		Addr      string `json:"addr"`
		BrotherDB int    `json:"brother_db"`
		UncleDB   int    `json:"uncle_db"`
	}
	Baidu struct {
		ApiKey    string `json:"api_key"`
		SecretKey string `json:"secret_key"`
	}
	Email struct {
		SmtpServer string `json:"smtp_server"`
		From       string `json:"from"`
		Pass       string `json:"pass"`
		To         string `json:"to"`
	} `json:"email"`
}

var (
	C   conf
	env = os.Getenv("GO_ENV")
)

func init() {
	// 尝试加载配置文件，否则使用参数
	if err := parseConfigFile(); err != nil {
		slog.Error("缺少配置文件 config-" + env + ".json")
		os.Exit(0)
	}

	if C.Ohmygpt.Key == "" || C.GptApiUs.Key == "" || C.Openai.Key == "" {
		slog.Error("OpenAI的Key不能为空")
		os.Exit(0)
	}

	if C.Http.Port == "" {
		C.Http.Port = "9001"
	}

	if C.Wechat.Token == "" {
		slog.Error("未设置公众号token，公众号功能不可用")
		os.Exit(0)
	}
}

func parseConfigFile() error {
	filename := fmt.Sprint("./config-", env, ".json")
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	bs, _ := io.ReadAll(f)
	err = json.Unmarshal(bs, &C)
	return err
}
