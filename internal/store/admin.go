package store

func getEmailNotificationKey() string {
	return "email-notification"
}

func SetEmailNotificationStatus(status string) error {
	return client.Set(ctx, getEmailNotificationKey(), status, 0).Err()
}

func GetEmailNotificationStatus() (status string, err error) {
	return client.Get(ctx, getEmailNotificationKey()).Result()
}
