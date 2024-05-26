package store

import "time"

func buildReceivedImageUrlsKey(user string) string {
	return "user:" + user + ":received-image-urls"
}

func GetReceivedImageUrls(user string) ([]string, error) {
	return client.LRange(ctx, buildReceivedImageUrlsKey(user), 0, -1).Result()
}

func AppendReceivedImageUrl(user string, imageUrl string) error {
	err := client.RPush(ctx, buildReceivedImageUrlsKey(user), imageUrl).Err()
	if err != nil {
		return err
	}

	return client.Expire(ctx, buildReceivedImageUrlsKey(user), 10*time.Minute).Err()
}

func DelReceivedImageUrls(user string) error {
	return client.Del(ctx, buildReceivedImageUrlsKey(user)).Err()
}
