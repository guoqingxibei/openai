package util

import (
	"openai/internal/constant"
	"os"
)

func AccountIsUncle() bool {
	return os.Getenv("ACCOUNT") == constant.Uncle
}

func GetAccount() string {
	if AccountIsUncle() {
		return constant.Uncle
	}
	return constant.Brother
}

func EnvIsProd() bool {
	return GetEnv() == constant.Prod
}

func GetEnv() string {
	return os.Getenv("GO_ENV")
}

func GetPayLink(user string) string {
	payLink := "https://brother.cxyds.top/shop"
	if AccountIsUncle() {
		payLink += "?uncle_openid=" + user
	}
	return payLink
}

func GetInvitationTutorialLink() string {
	if AccountIsUncle() {
		return "https://cxyds.top/2023/10/18/invitation-uncle.html"
	}
	return "https://cxyds.top/2023/10/18/invitation-brother.html"
}
