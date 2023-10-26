package store

import "strconv"

func buildQuotaKey(user string, day string) string {
	return "user:" + user + ":day:" + day + ":quota"
}

func SetQuota(user string, day string, quota int) error {
	return client.Set(ctx, buildQuotaKey(user, day), quota, DAY).Err()
}

func GetQuota(user string, day string) (int, error) {
	quotaStr, err := client.Get(ctx, buildQuotaKey(user, day)).Result()
	if err != nil {
		return 0, err
	}
	quota, err := strconv.Atoi(quotaStr)
	return quota, err
}
