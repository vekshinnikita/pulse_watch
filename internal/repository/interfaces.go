package repository

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

//go:generate mockgen -destination=mocks/mock_auth_repository.go -package=mock_repository . AuthRepository
//go:generate mockgen -destination=mocks/mock_app_repository.go -package=mock_repository . AppRepository
//go:generate mockgen -destination=mocks/mock_log_repository.go -package=mock_repository . LogRepository
//go:generate mockgen -destination=mocks/mock_alert_repository.go -package=mock_repository . AlertRepository
//go:generate mockgen -destination=mocks/mock_logs_es_repository.go -package=mock_repository . LogsESRepository
//go:generate mockgen -destination=mocks/mock_message_producer.go -package=mock_repository . MessageProducer
//go:generate mockgen -destination=mocks/mock_analytics_redis_repository.go -package=mock_repository . AnalyticsRedisRepository
//go:generate mockgen -destination=mocks/mock_metric_repository.go -package=mock_repository . MetricRepository
//go:generate mockgen -destination=mocks/mock_auth_redis_repository.go -package=mock_repository . AuthRedisRepository

type AuthRepository interface {
	CreateUser(ctx context.Context, createUser *entities.SignUpUser) (int, error)
	GetUserById(ctx context.Context, userId int) (*models.User, error)
	GetRolePermissionsByCode(
		ctx context.Context,
		roleCode string,
	) ([]models.Permission, error)
	GetUserByUsernameAndPassword(
		ctx context.Context,
		username string,
		passwordHash string,
	) (*models.User, error)
	IsRefreshTokenValid(context.Context, int, string) (bool, error)
	SaveRefreshToken(context.Context, int, string, *time.Time) (int, error)
	RevokeRefreshToken(context.Context, string) error
}

type AppRepository interface {
	// app
	CreateApp(ctx context.Context, data *entities.CreateAppData) (int, error)
	GetApp(ctx context.Context, appId int) (*models.App, error)
	GetAppIds(
		ctx context.Context,
	) ([]int, error)
	DeleteApp(ctx context.Context, appId int) error

	// api key
	CreateApiKey(ctx context.Context, data *dtos.CreateApiKeyData) (int, error)
	GetAppApiKeysPaginated(
		ctx context.Context,
		p *entities.PaginationData,
		appId int,
	) (*entities.PaginationResult[models.ApiKey], error)
	RevokeApiKey(ctx context.Context, appId int) error
	CheckApiKey(ctx context.Context, keyHash string) (int, error)
}

type LogRepository interface {
	GetAppLogMetaVarsByCodes(
		ctx context.Context,
		appId int,
		varCodes []string,
	) ([]models.LogMetaVar, error)

	CreateMetaVars(ctx context.Context, data []dtos.CreateMetaVar) ([]int, error)
	GetMetaVarsPaginated(
		ctx context.Context,
		pagination *entities.PaginationData,
		appId *int,
	) (*entities.PaginationResult[entities.LogMetaVarResult], error)
}

type AlertRepository interface {
	GetFullAlertsByIds(
		ctx context.Context,
		alertIds []int,
	) ([]models.AlertFull, error)

	GetRecentAlertsByRules(
		ctx context.Context,
		data *dtos.GetRecentAlerts,
	) ([]models.Alert, error)

	CreateAlerts(
		ctx context.Context,
		data []dtos.CreateAlert,
	) ([]int, error)

	SetResolvedAlerts(
		ctx context.Context,
		alertIds []int,
	) error

	GetAppRules(
		ctx context.Context,
		appId int,
	) ([]models.AlertRule, error)

	GetRule(
		ctx context.Context,
		ruleId int,
	) (*models.AlertRule, error)

	CreateAlertRule(
		ctx context.Context,
		data *entities.CreateAlertRule,
	) (int, error)

	GetAppRulesPaginated(
		ctx context.Context,
		p *entities.PaginationData,
		appId int,
	) (*entities.PaginationResult[models.AlertRule], error)

	UpdateAlertRule(
		ctx context.Context,
		ruleId int,
		data *entities.UpdateAlertRule,
	) error
	DeleteAlertRule(
		ctx context.Context,
		ruleId int,
	) error
}

type LogsESRepository interface {
	BulkSave(ctx context.Context, logs []entities.EnrichedAppLog) error
	SearchPaginated(
		ctx context.Context,
		data *entities.SearchLogData,
	) (*entities.PaginationResult[entities.EnrichedAppLog], error)
}

type MessageProducer interface {
	Publish(
		ctx context.Context,
		options *dtos.MessageOptions,
	) error
	PublishRaw(
		ctx context.Context,
		message *kafka.Message,
	) error

	Close()
}

type AnalyticsRedisRepository interface {
	AddExpire(
		ctx context.Context,
		name string,
		expire time.Duration,
	) error
	AddMetric(
		ctx context.Context,
		name string,
		key string,
		value int,
	) error
	AddUniqueMetric(
		ctx context.Context,
		name string,
		metric *dtos.UniqueMetric,
	) error

	GetKeysCursor(
		ctx context.Context,
		cursor uint64,
		pattern string,
		count int,
	) ([]string, uint64, error)

	TransferLiveMetricsToStreams(
		ctx context.Context,
		metricKeys []string,
	) ([]dtos.TransferredMetrics, error)

	GetAggregatedMetricsByApps(
		ctx context.Context,
		appIds []int,
		windowStart time.Time,
		windowEnd time.Time,
	) ([]dtos.AggregatedMetric, error)

	GetAggregatedUniqueMetricsByApps(
		ctx context.Context,
		appIds []int,
		periodType constants.PeriodType,
		windowStart time.Time,
	) ([]dtos.AggregatedMetric, error)

	GetAggregatedLiveMetrics(
		ctx context.Context,
		appId int,
		start time.Time,
		end time.Time,
	) ([]dtos.AggregatedMetric, error)

	GetLiveMetrics(
		ctx context.Context,
		data *dtos.GetLiveMetricsData,
	) ([]dtos.LiveMetrics, error)

	SubscribePubSub(ctx context.Context, channelId string) (*redis.PubSub, error)
	SubscribeStream(ctx context.Context, channelIds ...string) ([]redis.XStream, error)
	PublishToChannel(ctx context.Context, channelId string, data any) error
}

type MetricRepository interface {
	CreateAppMetrics(
		ctx context.Context,
		data []dtos.CreateAppMetric,
	) ([]int, error)
	GetMetrics(
		ctx context.Context,
		appId int,
		data *entities.GetMetricsData,
	) ([]models.AppMetric, error)
	GetAggregatedAppMetricsByPeriod(
		ctx context.Context,
		appIds []int,
		periodType constants.PeriodType,
		startTime time.Time,
		endTime time.Time,
	) ([]dtos.AggregatedAppMetric, error)
	GetAggregatedMetricsForAlert(
		ctx context.Context,
		appId int,
		data []dtos.AggregateMetricsForAlert,
	) ([]dtos.AggregatedAppMetricForAlert, error)
	ClearMetrics(
		ctx context.Context,
		data dtos.ClearMetrics,
	) error
}

type AuthRedisRepository interface {
	LoadPermissions(
		ctx context.Context,
		roleCode string,
		permissions []models.Permission,
	) error
	CheckPermission(
		ctx context.Context,
		roleCode string,
		permissionCode string,
	) (bool, error)
	CheckAnyPermissions(
		ctx context.Context,
		roleCode string,
		permissionCodes []string,
	) (bool, error)
}
