package util

import "os"

func AccountIsUncle() bool {
	return os.Getenv("ACCOUNT") == "uncle"
}

func GetPayLink(user string) string {
	payLink := "https://brother.cxyds.top/shop"
	if AccountIsUncle() {
		payLink += "?uncle_openid=" + user
	}
	return payLink
}
