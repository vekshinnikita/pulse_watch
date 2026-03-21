package redis_db

import (
	"fmt"

	"github.com/hibiken/asynq"
)

func GetAsynqClientOpt() *asynq.RedisClientOpt {
	config := GetConfig()

	address := fmt.Sprintf("%v:%v/asynq", config.Host, config.Port)
	return &asynq.RedisClientOpt{
		Addr:     address,
		Password: config.Password,
	}
}
