package repository

import (
	trmRedis "github.com/avito-tech/go-transaction-manager/drivers/goredis8/v2"
	trmSqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	elasticsearch_repository "github.com/vekshinnikita/pulse_watch/internal/repository/elasticsearch"
	kafka_repository "github.com/vekshinnikita/pulse_watch/internal/repository/kafka"
	postgres_repository "github.com/vekshinnikita/pulse_watch/internal/repository/postgres"
	redis_repository "github.com/vekshinnikita/pulse_watch/internal/repository/redis"
)

type RepositoryParams struct {
	Database     *sqlx.DB
	SQLTRMGetter *trmSqlx.CtxGetter

	Producer kafka_repository.MessageProducer
	ESClient *elasticsearch.TypedClient

	RedisClient    *redis.Client
	RedisTRMGetter *trmRedis.CtxGetter
}

type Repository struct {
	Auth   AuthRepository
	App    AppRepository
	Log    LogRepository
	Alert  AlertRepository
	Metric MetricRepository

	LogsES         LogsESRepository
	AnalyticsRedis AnalyticsRedisRepository
	AuthRedis      AuthRedisRepository

	Producer kafka_repository.MessageProducer
}

func NewRepository(params *RepositoryParams) *Repository {
	return &Repository{
		Auth:   postgres_repository.NewAuthPostgres(params.Database, params.SQLTRMGetter),
		App:    postgres_repository.NewAppPostgres(params.Database, params.SQLTRMGetter),
		Log:    postgres_repository.NewLogPostgres(params.Database, params.SQLTRMGetter),
		Alert:  postgres_repository.NewAlertPostgres(params.Database, params.SQLTRMGetter),
		Metric: postgres_repository.NewMetricPostgres(params.Database, params.SQLTRMGetter),

		LogsES:         elasticsearch_repository.NewLogsES(params.ESClient),
		AnalyticsRedis: redis_repository.NewAnalyticsRedis(params.RedisClient, params.RedisTRMGetter),
		AuthRedis:      redis_repository.NewAuthRedis(params.RedisClient, params.RedisTRMGetter),

		Producer: params.Producer,
	}
}
