package constants

import (
	"fmt"
	"time"
)

const liveMetricHashFormatString = "metric:live:%d:%d" // Вставляем по порядку app_id, time(unix)

const uniqueMetricFormatString = "metric:unique:%s:%s:%d:%d" // Вставляем по порядку name, period, app_id, time(unix)

func FormatLiveMetricName(appId int, t time.Time) string {
	return fmt.Sprintf(liveMetricHashFormatString, appId, t.Unix())
}

func FormatUniqueMetricName(name string, period string, appId int, time time.Time) string {
	return fmt.Sprintf(uniqueMetricFormatString, period, name, appId, time.Unix())
}

const (
	MetricExpire          = 2 * time.Minute
	BaseMetricWindowTime  = 5  // Секунды
	BaseMetricCommitDelay = 10 // Секунды
)
