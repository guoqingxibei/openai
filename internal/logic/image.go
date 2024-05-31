package logic

import (
	"openai/internal/constant"
	"openai/internal/store"
	"openai/internal/util"
	"strings"
)

const (
	wechatImageDir = constant.Temp + "/wechat-images"
)

func CalAndStoreImageTokens(imageUrl string) (err error) {
	name := strings.ReplaceAll(imageUrl, ":", "")
	name = strings.ReplaceAll(name, "/", "_")
	name += ".jpg"

	path := wechatImageDir + "/" + name
	err = util.DownloadFile(imageUrl, path)
	if err != nil {
		return
	}

	width, height, err := util.GetImageSize(path)
	if err != nil {
		return
	}

	tokens := util.CountImageTokens(width, height)
	_ = store.SetImageTokens(imageUrl, tokens)
	return
}
