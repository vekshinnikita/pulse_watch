package constants

type KafkaTopic string

const (
	LogsTopic       KafkaTopic = "logs"
	LogsDLQTopic    KafkaTopic = "logs.dlq"
	AlertsSendTopic KafkaTopic = "alerts.send"
)

const (
	GroupIdCtxKey = "GroupId"
)
