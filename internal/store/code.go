package store

func buildCodeKey(code string) string {
	return "code:" + code
}

func SetCodeDetail(code string, codeDetail string, useBrotherDB bool) error {
	myClient := client
	if useBrotherDB {
		myClient = brotherClient
	}
	return myClient.Set(ctx, buildCodeKey(code), codeDetail, 0).Err()
}

func GetCodeDetail(code string) (string, error) {
	return client.Get(ctx, buildCodeKey(code)).Result()
}
