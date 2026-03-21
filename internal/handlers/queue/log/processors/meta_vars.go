package log_queue_processors

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type MetaVarsProcessor struct {
	services *service.Service
}

type metaVarInfoMap map[string]constants.MetaVarType

func NewMetaVarsProcessor(services *service.Service) *MetaVarsProcessor {
	return &MetaVarsProcessor{
		services: services,
	}
}

func (uc *MetaVarsProcessor) getMetaVarType(value any) constants.MetaVarType {
	strValue, ok := value.(string)
	if ok {
		// Попытка спарсить дату и время
		_, err := time.Parse(time.RFC3339Nano, strValue)
		if err == nil {
			return constants.MetaVarTypeDatetime
		}

		// Попытка спарсить дату
		_, err = time.Parse(time.DateOnly, strValue)
		if err == nil {
			return constants.MetaVarTypeDate
		}

		return constants.MetaVarTypeString
	}

	_, ok = value.(int)
	if ok {
		return constants.MetaVarTypeNumber
	}

	_, ok = value.(float64)
	if ok {
		return constants.MetaVarTypeNumber
	}

	// по умолчанию
	return constants.MetaVarTypeString
}

func (uc *MetaVarsProcessor) parseUniqueMeta(logs []entities.EnrichedAppLog) metaVarInfoMap {
	// создание списка уникальных кодов
	uniqueTypeMap := make(metaVarInfoMap)
	for _, log := range logs {
		for key, value := range log.Meta {
			_, ok := uniqueTypeMap[key]
			if !ok {
				uniqueTypeMap[key] = uc.getMetaVarType(value)
			}
		}
	}

	return uniqueTypeMap
}

func (uc *MetaVarsProcessor) getNotExistingMetaMap(
	ctx context.Context,
	logs []entities.EnrichedAppLog,
) (metaVarInfoMap, error) {
	appId := logs[0].AppId
	uniqueMetaMap := uc.parseUniqueMeta(logs)
	existsMetaVars, err := uc.services.Logs.GetExistingMetaVars(ctx, appId, utils.MapKeys(uniqueMetaMap))
	if err != nil {
		return nil, fmt.Errorf("get meta vars by codes: %w", err)
	}

	// Создание множества существующих кодов
	existingCodeSet := make(map[string]struct{}, 0)
	for _, metaVar := range existsMetaVars {
		existingCodeSet[metaVar.Code] = struct{}{}
	}

	// Создание map не существующих кодов
	notExistsMetaMap := make(metaVarInfoMap, 0)
	for code, metaType := range uniqueMetaMap {
		_, ok := existingCodeSet[code]
		if !ok {
			notExistsMetaMap[code] = metaType
		}
	}

	return notExistsMetaMap, nil
}

func (uc *MetaVarsProcessor) makeCreateMetaVars(
	appId int,
	metaMap metaVarInfoMap,
) []dtos.CreateMetaVar {
	data := make([]dtos.CreateMetaVar, 0)
	for code, metaType := range metaMap {
		data = append(data, dtos.CreateMetaVar{
			AppId: appId,
			Name:  code,
			Code:  code,
			Type:  metaType,
		})
	}

	return data
}

func (uc *MetaVarsProcessor) storeNewMetaVars(
	ctx context.Context,
	appId int,
	metaMap metaVarInfoMap,
) error {
	data := uc.makeCreateMetaVars(appId, metaMap)

	_, err := uc.services.Logs.CreateMetaVars(ctx, data)
	if err != nil {
		return fmt.Errorf("create meta vars: %w", err)
	}

	return nil
}

// Process обрабатывает новые мета переменные и сохраняет их в бд
// для понимания того какие мета переменные могут быть у логов.
func (uc *MetaVarsProcessor) Process(ctx context.Context, logs []entities.EnrichedAppLog) error {
	if len(logs) == 0 {
		return nil
	}

	notExistingMetaMap, err := uc.getNotExistingMetaMap(ctx, logs)
	if err != nil {
		return fmt.Errorf("get not exists meta vars code: %w", err)
	}

	if len(notExistingMetaMap) == 0 {
		return nil
	}

	slog.InfoContext(ctx, fmt.Sprintf("find %v new log meta vars", len(notExistingMetaMap)))

	// Сохранение новых мета переменных
	appId := logs[0].AppId
	err = uc.storeNewMetaVars(ctx, appId, notExistingMetaMap)
	if err != nil {
		return fmt.Errorf("store new meta vars: %w", err)
	}

	slog.InfoContext(ctx, "new meta vars is saved")
	return err
}
