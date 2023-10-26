package store

func getInvitationCodeCursorKey() string {
	return "invitation-code-cursor"
}

func IncInvitationCodeCursor() (int64, error) {
	return client.Incr(ctx, getInvitationCodeCursorKey()).Result()
}

func buildInvitationCodeKey(user string) string {
	return "user:" + user + ":invitation-code"
}

func GetInvitationCode(user string) (string, error) {
	return client.Get(ctx, buildInvitationCodeKey(user)).Result()
}

func SetInvitationCode(user string, code string) error {
	return client.Set(ctx, buildInvitationCodeKey(user), code, 0).Err()
}

func buildUserKey(code string) string {
	return "invitation-code:" + code + ":user"
}

func GetUserByInvitationCode(code string) (string, error) {
	return client.Get(ctx, buildUserKey(code)).Result()
}

func SetUserByInvitationCode(code string, user string) error {
	return client.Set(ctx, buildUserKey(code), user, 0).Err()
}

func buildInviter(user string) string {
	return "user:" + user + ":inviter"
}

func SetInviter(user string, inviter string) error {
	return client.Set(ctx, buildInviter(user), inviter, 0).Err()
}

func GetInviter(user string) (string, error) {
	return client.Get(ctx, buildInviter(user)).Result()
}
