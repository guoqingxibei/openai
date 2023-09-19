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

func GetPayLink(user string) string {
	payLink := "https://brother.cxyds.top/shop"
	if AccountIsUncle() {
		payLink += "?uncle_openid=" + user
	}
	return payLink
}
