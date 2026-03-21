package redis_db

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

func NewClient() (*redis.Client, error) {
	config := GetConfig()

	address := fmt.Sprintf("%v:%v", config.Host, config.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: config.Password,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return client, nil
}
