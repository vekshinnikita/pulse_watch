package service

import (
	"context"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

//go:generate mockgen -destination=mocks/mock_auth_service.go -package=mock_service . AuthService
//go:generate mockgen -destination=mocks/mock_app_service.go -package=mock_service . AppService
//go:generate mockgen -destination=mocks/mock_logs_service.go -package=mock_service . LogsService
//go:generate mockgen -destination=mocks/mock_alert_service.go -package=mock_service . AlertService
//go:generate mockgen -destination=mocks/mock_metric_service.go -package=mock_service . MetricService

type AuthService interface {
	// user
	CreateUser(ctx context.Context, createUser *entities.SignUpUser) (int, error)
	GetUserById(ctx context.Context, userId int) (*models.User, error)
	CreateAndGetUser(ctx context.Context, createUser *entities.SignUpUser) (*models.User, error)

	// token
	GenerateTokens(
		ctx context.Context,
		user *models.User,
	) (*entities.AuthTokens, error)
	SignIn(ctx context.Context, signInUser *entities.SignInUser) (*entities.AuthTokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*entities.AuthTokens, error)
	ParseToken(ctx context.Context, accessToken string) (*entities.TokenClaims, error)
	ParseAccessToken(
		ctx context.Context,
		accessToken string,
	) (*entities.AccessTokenClaims, error)

	// permissions
	GetCurrentUser(ctx context.Context) (*models.User, error)
	GetCurrentUserRole(ctx context.Context) (*models.Role, error)
	CheckRolePermission(
		ctx context.Context,
		roleCode string,
		permissionCode string,
	) (bool, error)
	CheckRoleAnyPermission(
		ctx context.Context,
		roleCode string,
		permissionCodes []string,
	) (bool, error)
	CheckCurrentUserPermission(ctx context.Context, permissionCode string) (bool, error)
	CheckCurrentUserAnyPermission(
		ctx context.Context,
		permissionCodes []string,
	) (bool, error)
}

type AppService interface {
	// app
	CreateApp(ctx context.Context, data *entities.CreateAppData) (int, error)
	GetApp(ctx context.Context, appId int) (*models.App, error)
	GetAppIds(ctx context.Context) ([]int, error)
	CreateAndGetApp(ctx context.Context, data *entities.CreateAppData) (*models.App, error)
	DeleteApp(ctx context.Context, appId int) error

	// api key
	CreateApiKey(ctx context.Context, data *entities.CreateApiKeyData) (*models.ApiKeyWithKey, error)
	GetAppApiKeysPaginated(
		ctx context.Context,
		p *entities.PaginationData,
		appId int,
	) (*entities.PaginationResult[models.ApiKey], error)
	RevokeApiKey(ctx context.Context, apiKeyId int) error
	CheckApiKey(ctx context.Context, apiKey string) (int, error)
}

type LogsService interface {
	SendLogs(ctx context.Context, logs entities.AppLogs) error
	SaveLogs(ctx context.Context, logs []entities.EnrichedAppLog) error
	SearchPaginatedLogs(
		ctx context.Context,
		data *entities.SearchLogData,
	) (*entities.PaginationResult[entities.EnrichedAppLog], error)

	GetExistingMetaVars(
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

type AlertService interface {
	GetFullAlertsByIds(
		ctx context.Context,
		alertIds []int,
	) ([]models.AlertFull, error)

	GetRecentAlertsByRules(
		ctx context.Context,
		data *dtos.GetRecentAlerts,
	) ([]models.Alert, error)

	CreateAlertsTask(
		ctx context.Context,
		data []dtos.CreateAlert,
	) error

	SetResolvedAlerts(
		ctx context.Context,
		alertIds []int,
	) error

	// Alert rule

	GetRule(
		ctx context.Context,
		ruleId int,
	) (*models.AlertRule, error)

	GetAppRules(ctx context.Context, appId int) ([]models.AlertRule, error)

	GetAppRulesPaginated(
		ctx context.Context,
		p *entities.PaginationData,
		appId int,
	) (*entities.PaginationResult[models.AlertRule], error)

	CreateAndGetAlertRule(
		ctx context.Context,
		data *entities.CreateAlertRule,
	) (*models.AlertRule, error)

	CreateAlertRule(
		ctx context.Context,
		data *entities.CreateAlertRule,
	) (int, error)

	UpdateAlertRule(
		ctx context.Context,
		ruleId int,
		data *entities.UpdateAlertRule,
	) error

	DeleteAlertRule(
		ctx context.Context,
		ruleId int,
	) error

	GetMetricsByRules(
		ctx context.Context,
		rules []models.AlertRule,
	) (dtos.MetricsByRuleIds, error)
}

type MetricService interface {
	IncrementMetrics(
		ctx context.Context,
		metrics dtos.MetricsMap,
		uniqueMetrics dtos.UniqueMetricsMap,
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

	AggregateOneMinuteMetrics(
		ctx context.Context,
		appIds []int,
		windowStart time.Time,
	) error

	GetAggregatedAppMetricsByPeriod(
		ctx context.Context,
		appIds []int,
		periodType constants.PeriodType,
		startTime time.Time,
		endTime time.Time,
	) ([]dtos.AggregatedAppMetric, error)

	CreateAppMetrics(
		ctx context.Context,
		data []dtos.CreateAppMetric,
	) ([]int, error)

	GetAggregatedUniqueMetricsByApps(
		ctx context.Context,
		appIds []int,
		periodType constants.PeriodType,
		windowStart time.Time,
	) ([]dtos.AggregatedMetric, error)

	GetMetrics(
		ctx context.Context,
		appId int,
		data *entities.GetMetricsData,
	) (*entities.ResultMetricsGroups, error)

	ClearMetrics(
		ctx context.Context,
		data dtos.ClearMetrics,
	) error

	SubscribeChannel(
		ctx context.Context,
		channelId string,
		handler func(channelId string, message string),
	) error

	SendChannelMessage(
		client *entities.WSClient,
		channelId string,
		incomingMessage string,
	) error

	PublishToChannelBulk(ctx context.Context, data []dtos.PublishToChannel) error
	PublishMetricsToChannels(
		ctx context.Context,
		createMetrics []dtos.CreateAppMetric,
	) error
}
