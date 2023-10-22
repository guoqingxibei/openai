package store

import (
	"github.com/bsm/redislock"
)

var locker *redislock.Client

func GetLocker() *redislock.Client {
	if locker == nil {
		locker = redislock.New(client)
	}
	return locker
}
