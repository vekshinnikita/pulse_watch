package logs_service

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	logs_usecases "github.com/vekshinnikita/pulse_watch/internal/service/logs/usecases"
)

type LogsUseCases struct {
	sendLogs            SendLogsUseCase
	saveLogs            SaveLogsUseCase
	getExistingMetaVars GetExistingMetaVars
	createMetaVars      CreateMetaVars
}

type LogsService struct {
	trm  repository.TransactionManager
	repo *repository.Repository
	uc   *LogsUseCases
}

func NewDefaultLogsService(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *LogsService {
	return &LogsService{
		repo: repo,
		trm:  trm,
		uc: &LogsUseCases{
			sendLogs:            logs_usecases.NewSendLogsUseCase(trm, repo),
			saveLogs:            logs_usecases.NewSaveLogsUseCase(trm, repo),
			getExistingMetaVars: logs_usecases.NewGetExistingMetaVarsUseCase(trm, repo),
			createMetaVars:      logs_usecases.NewCreateMetaVarsUseCase(trm, repo),
		},
	}
}

func (s *LogsService) SendLogs(ctx context.Context, logs entities.AppLogs) error {
	return s.uc.sendLogs.Send(ctx, logs)
}

func (s *LogsService) SaveLogs(ctx context.Context, logs []entities.EnrichedAppLog) error {
	return s.uc.saveLogs.Save(ctx, logs)
}

func (s *LogsService) SearchPaginatedLogs(
	ctx context.Context,
	data *entities.SearchLogData,
) (*entities.PaginationResult[entities.EnrichedAppLog], error) {
	return s.repo.LogsES.SearchPaginated(ctx, data)
}

func (s *LogsService) GetExistingMetaVars(
	ctx context.Context,
	appId int,
	varCodes []string,
) ([]models.LogMetaVar, error) {
	return s.uc.getExistingMetaVars.Get(ctx, appId, varCodes)
}

func (s *LogsService) CreateMetaVars(ctx context.Context, data []dtos.CreateMetaVar) ([]int, error) {
	return s.uc.createMetaVars.Create(ctx, data)
}

func (s *LogsService) GetMetaVarsPaginated(
	ctx context.Context,
	pagination *entities.PaginationData,
	appId *int,
) (*entities.PaginationResult[entities.LogMetaVarResult], error) {
	return s.repo.Log.GetMetaVarsPaginated(ctx, pagination, appId)
}
