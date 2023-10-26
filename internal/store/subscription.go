package store

import "strconv"

func buildSubscribeTimestampKey(user string) string {
	return "user:" + user + ":subscribe-timestamp"
}

func SetSubscribeTimestamp(user string, timestamp int64) error {
	return client.Set(ctx, buildSubscribeTimestampKey(user), strconv.FormatInt(timestamp, 10), 0).Err()
}

func GetSubscribeTimestamp(user string) (int64, error) {
	timestampStr, err := client.Get(ctx, buildSubscribeTimestampKey(user)).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(timestampStr, 10, 64)
}
