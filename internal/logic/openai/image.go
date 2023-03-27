package openailogic

import (
	"github.com/redis/go-redis/v9"
	"log"
	"openai/internal/service/gptredis"
)

const defaultImageBalance = 5

func FetchImageBalance(user string) int {
	balance, err := gptredis.FetchImageBalance(user)
	if err != nil {
		if err == redis.Nil {
			err := gptredis.SetImageBalance(user, defaultImageBalance)
			if err != nil {
				log.Println("Failed to set ImageBalance to default", err)
				return 0
			}
			return defaultImageBalance
		}
		log.Println("Failed to load image balance", err)
		return 0
	}
	return balance
}