package wechat

import "openai/internal/util"

func init() {
	initMedias()
	if !util.AccountIsUncle() || !util.EnvIsProd() {
		createOrUpdateMenu()
	}
}
