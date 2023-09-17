package wechat

import "openai/internal/util"

func init() {
	initToken()
	initMedias()
	if !util.AccountIsUncle() {
		createOrUpdateMenu()
	}
}
