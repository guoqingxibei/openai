package store

import "strconv"

func GetBalance(user string, day string) (int, error) {
	balance, err := client.Get(ctx, buildBalanceKey(user, day)).Result()
	cnt, _ := strconv.Atoi(balance)
	return cnt, err
}

func SetBalance(user string, day string, balance int) error {
	return client.Set(ctx, buildBalanceKey(user, day), strconv.Itoa(balance), DAY).Err()
}

func DecrBalance(user string, day string) (int, error) {
	balance, err := client.Decr(ctx, buildBalanceKey(user, day)).Result()
	return int(balance), err
}

func buildBalanceKey(user string, day string) string {
	return "user:" + user + ":day:" + day + ":balance"
}

func SetPaidBalance(user string, balance int) error {
	return SetPaidBalanceWithDB(user, balance, false)
}

func SetPaidBalanceWithDB(user string, balance int, useUncleDB bool) error {
	myClient := client
	if useUncleDB {
		myClient = uncleClient
	}
	return myClient.Set(ctx, buildPaidBalance(user), balance, 0).Err()
}

func GetPaidBalance(user string) (int, error) {
	return GetPaidBalanceWithDB(user, false)
}

func GetPaidBalanceWithDB(user string, useUncleDB bool) (int, error) {
	myClient := client
	if useUncleDB {
		myClient = uncleClient
	}
	balanceStr, err := myClient.Get(ctx, buildPaidBalance(user)).Result()
	if err != nil {
		return 0, err
	}
	balance, err := strconv.Atoi(balanceStr)
	return balance, err
}

func DecrPaidBalance(user string, decrement int64) (int64, error) {
	return client.DecrBy(ctx, buildPaidBalance(user), decrement).Result()
}

func buildPaidBalance(user string) string {
	return "user:" + user + ":paid-balance"
}
