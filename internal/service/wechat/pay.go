package wechat

import (
	"context"
	"errors"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"log"
	"openai/internal/config"
	"openai/internal/service/recorder"
	"openai/internal/util"
	"time"
)

var client *wechat.ClientV3
var ctx = context.Background()
var wcCfg = config.C.Wechat

func init() {
	if !util.AccountIsUncle() && util.EnvIsProd() {
		initPayClient()
	}
}

func initPayClient() {
	// NewClientV3 初始化微信客户端 v3
	// mchid：商户ID 或者服务商模式的 sp_mchid
	// serialNo：商户证书的证书序列号
	// apiV3Key：apiV3Key，商户平台获取
	// privateKey：私钥 apiclient_key.pem 读取后的内容
	var err error
	client, err = wechat.NewClientV3(wcCfg.MchId, wcCfg.SerialNo, wcCfg.APIv3Key, wcCfg.PrivateKey)
	if err != nil {
		recorder.RecordError("wechat.NewClientV3() failed", err)
		return
	}

	// 设置微信平台API证书和序列号（推荐开启自动验签，无需手动设置证书公钥等信息）
	//client.SetPlatformCert([]byte(""), "")

	// 启用自动同步返回验签，并定时更新微信平台API证书（开启自动验签时，无需单独设置微信平台API证书和序列号）
	err = client.AutoVerifySign()
	if err != nil {
		recorder.RecordError("client.AutoVerifySign() failed", err)
		return
	}

	// 自定义配置http请求接收返回结果body大小，默认 10MB
	//client.SetBodySize() // 没有特殊需求，可忽略此配置

	// 打开Debug开关，输出日志，默认是关闭的
	client.DebugSwitch = gopay.DebugOn
}

func InitiateTransaction(openid string, tradeNo string, total int, description string) (string, error) {
	expire := time.Now().Add(10 * time.Minute).Format(time.RFC3339)
	bm := make(gopay.BodyMap)
	bm.Set("appid", wcCfg.AppId).
		Set("description", description).
		Set("out_trade_no", tradeNo).
		Set("time_expire", expire).
		Set("notify_url", wcCfg.NotifyUrl).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", total).
				Set("currency", "CNY")
		}).
		SetBodyMap("payer", func(bm gopay.BodyMap) {
			bm.Set("openid", openid)
		})

	wxRsp, err := client.V3TransactionJsapi(ctx, bm)
	if err != nil {
		recorder.RecordError("client.V3TransactionJsapi() failed", err)
		return "", err
	}
	if wxRsp.Code != wechat.Success {
		recorder.RecordError("client.V3TransactionJsapi() failed", errors.New(wxRsp.Error))
		return "", errors.New("wxRsp error")
	}
	log.Printf("wxRsp: %+v", wxRsp.Response)
	return wxRsp.Response.PrepayId, nil
}

func VerifySignAndDecrypt(notifyReq *wechat.V3NotifyReq) (*wechat.V3DecryptResult, error) {
	// 获取微信平台证书
	certMap := client.WxPublicKeyMap()
	// 验证异步通知的签名
	err := notifyReq.VerifySignByPKMap(certMap)
	if err != nil {
		return nil, err
	}

	return notifyReq.DecryptCipherText(wcCfg.APIv3Key)
}

func GeneratePaySignParams(prepayid string) (*wechat.JSAPIPayParams, error) {
	return client.PaySignOfJSAPI(wcCfg.AppId, prepayid)
}
