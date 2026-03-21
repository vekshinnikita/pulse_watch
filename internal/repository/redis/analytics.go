package redis_repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	trmRedis "github.com/avito-tech/go-transaction-manager/drivers/goredis8/v2"
	"github.com/go-redis/redis/v8"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type AnalyticsRedis struct {
	client *redis.Client
	getter *trmRedis.CtxGetter
}

func NewAnalyticsRedis(
	client *redis.Client,
	getter *trmRedis.CtxGetter,
) *AnalyticsRedis {
	return &AnalyticsRedis{
		client: client,
		getter: getter,
	}
}

func (r *AnalyticsRedis) AddExpire(
	ctx context.Context,
	name string,
	expire time.Duration,
) error {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	err := executor.Expire(ctx, name, expire).Err()
	if err != nil {
		return fmt.Errorf("run incr: %w", err)
	}

	return nil
}

func (r *AnalyticsRedis) AddMetric(
	ctx context.Context,
	name string,
	key string,
	value int,
) error {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	err := executor.HIncrBy(ctx, name, key, int64(value)).Err()
	if err != nil {
		return fmt.Errorf("run incr: %w", err)
	}

	return nil
}

func (r *AnalyticsRedis) AddUniqueMetric(
	ctx context.Context,
	name string,
	metric *dtos.UniqueMetric,
) error {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	script := `
	local exists = redis.call("EXISTS", KEYS[1])
	local expire = tonumber(ARGV[#ARGV])

	local values = {}
	for i = 1, #ARGV - 1 do
		values[i] = ARGV[i]
	end

	redis.call("PFADD", KEYS[1], unpack(values))

	if exists == 0 then
		redis.call("EXPIRE", KEYS[1], expire)
	end

	return 1
	`

	args := append(metric.Values, metric.Period.Seconds())
	cmd := executor.Eval(
		ctx,
		script,
		[]string{name},
		args...,
	)
	if cmd == nil {
		return nil
	}

	err := cmd.Err()
	if err != nil {
		return fmt.Errorf("run incr: %w", err)
	}

	return nil
}

func (r *AnalyticsRedis) GetKeysCursor(
	ctx context.Context,
	cursor uint64,
	pattern string,
	count int,
) ([]string, uint64, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	keys, nextCursor, err := executor.Scan(ctx, cursor, pattern, int64(count)).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("run scan: %w", err)
	}

	return keys, nextCursor, nil
}

func (r *AnalyticsRedis) TransferLiveMetricsToStreams(
	ctx context.Context,
	metricKeys []string,
) ([]dtos.TransferredMetrics, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	script := `
	local result = {}
	for i, key in ipairs(KEYS) do
		local local_result = {}

		local app_id = string.match(key, "^[^:]+:[^:]+:([^:]+)")
		local time_sec = tonumber(string.match(key, "^[^:]+:[^:]+:[^:]+:([^:]+)"))
		local time_ms = time_sec * 1000
		
		local live_metrics = redis.call("HGETALL", key)
		for i = 1, #live_metrics, 2 do
			local field = live_metrics[i]
			local value = tonumber(live_metrics[i + 1])

			local stream_name = "metric:live_stream:"..app_id..":"..field
			redis.call("XADD", stream_name, "MAXLEN", "~", 20, tostring(time_ms).."-*", "value", value)
			redis.call("EXPIRE", stream_name, 120)

			local_result[#local_result + 1] = {field, value}
		end

		result[#result+1] = {key, local_result}
	end

	return result
	`

	cmd := executor.Eval(
		ctx,
		script,
		metricKeys,
	)

	res, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("run transfer live metrics: %w", err)
	}

	transferred := make([]dtos.TransferredMetrics, 0)

	rows := res.([]interface{})
	for _, row := range rows {
		r := row.([]interface{})

		metrics := make([]dtos.TransferredMetric, 0)
		pairs := r[1].([]interface{})
		for _, pair := range pairs {
			p := pair.([]interface{})
			metrics = append(metrics, dtos.TransferredMetric{
				Name:  p[0].(string),
				Value: int(p[1].(int64)),
			})
		}

		transferred = append(transferred, dtos.TransferredMetrics{
			Key:     r[0].(string),
			Metrics: metrics,
		})
	}

	return transferred, nil
}

func (r *AnalyticsRedis) GetAggregatedMetricsByApps(
	ctx context.Context,
	appIds []int,
	windowStart time.Time,
	windowEnd time.Time,
) ([]dtos.AggregatedMetric, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	script := `
	local window_start_ms = tonumber(ARGV[1])*1000
	local window_end_ms = tonumber(ARGV[2])*1000

	local result = {}
	for i, app_id in ipairs(KEYS) do
		

		local keys = redis.call("KEYS", "metric:live_stream:"..tostring(app_id)..":*")
		for _, key in ipairs(keys) do
			local metric_name = string.match(key, "metric:live_stream:[^:]+:(.+)$")

			local sum = 0
			local entries = redis.call("XRANGE", key, window_start_ms, window_end_ms)
			for i, entry in ipairs(entries) do
				local value = tonumber(entry[2][2])
				sum = sum + value
			end

			result[#result+1] = {tonumber(app_id), metric_name, sum}
		end
		
	end

	return result
	`

	idsStr := utils.Map(appIds, func(v int) string { return strconv.Itoa(v) })
	cmd := executor.Eval(
		ctx,
		script,
		idsStr,
		windowStart.Unix(),
		windowEnd.Unix(),
	)
	res, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("run aggregate alert metrics: %w", err)
	}

	metrics := make([]dtos.AggregatedMetric, 0)

	rows := res.([]interface{})
	for _, row := range rows {
		r := row.([]interface{})
		metrics = append(metrics, dtos.AggregatedMetric{
			AppId:      int(r[0].(int64)),
			MetricName: r[1].(string),
			Value:      int(r[2].(int64)),
		})
	}

	return metrics, nil
}

func (r *AnalyticsRedis) GetAggregatedUniqueMetricsByApps(
	ctx context.Context,
	appIds []int,
	periodType constants.PeriodType,
	windowStart time.Time,
) ([]dtos.AggregatedMetric, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	script := `
	local period_type = ARGV[1]
	local window_start = tonumber(ARGV[2])

	local result = {}
	for i, app_id in ipairs(KEYS) do

		local keys = redis.call("KEYS", "metric:unique:"..period_type..":*:"..tostring(app_id)..":"..window_start)
		for _, key in ipairs(keys) do
			local count = redis.call("PFCOUNT", key)
			if count and count ~= 0 then     
				local metric_name = string.match(key, "metric:unique:[^:]+:([^:]+)")
				result[#result+1] = {tonumber(app_id), metric_name, count}
			end    

		end   
	end

	return result
	`

	idsStr := utils.Map(appIds, func(v int) string { return strconv.Itoa(v) })
	cmd := executor.Eval(
		ctx,
		script,
		idsStr,
		string(periodType),
		windowStart.Unix(),
	)
	res, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("run aggregate alert metrics: %w", err)
	}

	metrics := make([]dtos.AggregatedMetric, 0)

	rows := res.([]interface{})
	for _, row := range rows {
		r := row.([]interface{})
		metrics = append(metrics, dtos.AggregatedMetric{
			AppId:      int(r[0].(int64)),
			MetricName: r[1].(string),
			Value:      int(r[2].(int64)),
		})
	}

	return metrics, nil
}

func (r *AnalyticsRedis) GetAggregatedLiveMetrics(
	ctx context.Context,
	appId int,
	start time.Time,
	end time.Time,
) ([]dtos.AggregatedMetric, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	script := `
	local start_sec = tonumber(ARGV[1])
	local end_sec = tonumber(ARGV[2])

	local app_id = KEYS[1]

	local aggregation = {}

	local keys = redis.call("KEYS", "metric:live:"..app_id..":*")
	for _, key in ipairs(keys) do
		local time_sec = tonumber(string.match(key, "metric:live:[^:]+:([^:]+)"))

		if time_sec <= end_sec and time_sec >= start_sec then

			local live_metrics = redis.call("HGETALL", key)
			for i = 1, #live_metrics, 2 do
				local key = live_metrics[i]
				local value = tonumber(live_metrics[i + 1])

				aggregation[key] = (aggregation[key] or 0) + value
			end

		end
	end

	local result = {}
	for key, value in pairs(aggregation) do
		result[#result+1] = {key, value}
	end

	return result
	`

	cmd := executor.Eval(
		ctx,
		script,
		[]string{strconv.Itoa(appId)},
		start.Unix(),
		end.Unix(),
	)
	res, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("run script: %w", err)
	}

	metrics := make([]dtos.AggregatedMetric, 0)

	rows := res.([]interface{})
	for _, row := range rows {
		r := row.([]interface{})
		metrics = append(metrics, dtos.AggregatedMetric{
			AppId:      appId,
			MetricName: r[0].(string),
			Value:      int(r[1].(int64)),
		})
	}

	return metrics, nil
}

func (r *AnalyticsRedis) GetLiveMetrics(
	ctx context.Context,
	data *dtos.GetLiveMetricsData,
) ([]dtos.LiveMetrics, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	script := `
	local app_id = KEYS[1]
	local desired_metric = KEYS[2]
	local start_ms = tonumber(ARGV[1]) * 1000
	local end_ms = tonumber(ARGV[2]) * 1000

	local result = {}

	local keys = redis.call("KEYS", "metric:live_stream:"..app_id..":"..desired_metric)
	for _, key in ipairs(keys) do
		local metric_name = string.match(key, "metric:live_stream:[^:]+:(.+)$")
		local local_result = {}

		local entries = redis.call("XRANGE", key, start_ms, end_ms)
		for _, entry in ipairs(entries) do
			local time = tonumber(string.match(entry[1], "^(%d+)%-%d+$")) / 1000

			local value = tonumber(entry[2][2])

			local_result[#local_result + 1] = {time, value}
		end

		result[#result + 1] = {metric_name, local_result}
	end

	return result
	`

	cmd := executor.Eval(
		ctx,
		script,
		[]string{strconv.Itoa(data.AppId), data.MetricName},
		data.Start.Unix(),
		data.End.Unix(),
	)
	res, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("run script: %w", err)
	}

	gropedMetrics := make([]dtos.LiveMetrics, 0)

	fmt.Println(data.Start.Unix(), data.End.Unix())

	rows := res.([]interface{})
	for _, row := range rows {
		r := row.([]interface{})
		gropedMetric := dtos.LiveMetrics{
			AppId:      data.AppId,
			MetricName: r[0].(string),
			Metrics:    make([]dtos.LiveMetric, 0),
		}

		metrics := r[1].([]interface{})
		for _, metric := range metrics {
			m := metric.([]interface{})
			gropedMetric.Metrics = append(gropedMetric.Metrics, dtos.LiveMetric{
				PeriodStart: time.Unix(m[0].(int64), 0),
				Value:       int(m[1].(int64)),
			})
		}

		gropedMetrics = append(gropedMetrics, gropedMetric)
	}

	return gropedMetrics, nil
}

func (r *AnalyticsRedis) SubscribePubSub(ctx context.Context, channelId string) (*redis.PubSub, error) {
	pubsub := r.client.Subscribe(ctx, channelId)

	// Дожидаемся подтверждения подписки
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("subscribe receive: %w", err)
	}

	return pubsub, nil
}

func (r *AnalyticsRedis) PublishToChannel(ctx context.Context, channelId string, data any) error {
	err := r.client.Publish(ctx, channelId, data).Err()
	if err != nil {
		return fmt.Errorf("channel publish: %w", err)
	}

	return nil
}

func (r *AnalyticsRedis) SubscribeStream(ctx context.Context, channelIds ...string) ([]redis.XStream, error) {
	streams, err := r.client.XRead(ctx, &redis.XReadArgs{
		Streams: channelIds,
		Block:   0,
		Count:   1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("subscribe streams: %w", err)
	}

	return streams, nil
}
