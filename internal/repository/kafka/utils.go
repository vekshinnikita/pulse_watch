package kafka_repository

import (
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func getHeader(headers []kafka.Header, key string) []byte {
	for _, h := range headers {
		if h.Key == key {
			return h.Value
		}
	}
	return nil
}

func GetHeaderString(headers []kafka.Header, key string) string {
	return string(getHeader(headers, key))
}

func GetHeaderInt(headers []kafka.Header, key string) (int, bool) {
	value, err := strconv.Atoi(GetHeaderString(headers, key))
	if err != nil {
		return 0, false
	}

	return value, true
}
