package auth_usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type CheckPermissionUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCheckPermissionUseCase(trm repository.TransactionManager, repo *repository.Repository) *CheckPermissionUseCase {
	return &CheckPermissionUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *CheckPermissionUseCase) loadPerms(ctx context.Context, roleCode string) error {
	perms, err := s.repo.Auth.GetRolePermissionsByCode(ctx, roleCode)
	if err != nil {
		return fmt.Errorf("get role permissions: %w", err)
	}

	err = s.repo.AuthRedis.LoadPermissions(ctx, roleCode, perms)
	if err != nil {
		return fmt.Errorf("load permissions: %w", err)
	}

	return nil
}

func (s *CheckPermissionUseCase) loadAndCheck(ctx context.Context, roleCode, permissionCode string) (bool, error) {
	// загружаем список прав в кэш
	err := s.loadPerms(ctx, roleCode)
	if err != nil {
		return false, fmt.Errorf("load perms: %w", err)
	}

	// Повторная проверка списка прав
	ok, err := s.repo.AuthRedis.CheckPermission(ctx, roleCode, permissionCode)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	return ok, nil
}

func (s *CheckPermissionUseCase) loadAndCheckAny(ctx context.Context, roleCode string, permissionCodes []string) (bool, error) {
	// загружаем список прав в кэш
	err := s.loadPerms(ctx, roleCode)
	if err != nil {
		return false, fmt.Errorf("load perms: %w", err)
	}

	// Повторная проверка списка прав
	ok, err := s.repo.AuthRedis.CheckAnyPermissions(ctx, roleCode, permissionCodes)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	return ok, nil
}

func (s *CheckPermissionUseCase) Check(ctx context.Context, roleCode, permissionCode string) (bool, error) {
	// Админу можно все
	if roleCode == "admin" {
		return true, nil
	}

	// Проверка списка прав
	ok, err := s.repo.AuthRedis.CheckPermission(ctx, roleCode, permissionCode)
	if err != nil {

		// Список не загружен
		var nek *errs.NotExistKeyRedisError
		if errors.As(err, &nek) {
			return s.loadAndCheck(ctx, roleCode, permissionCode)
		}

		return false, fmt.Errorf("check permission: %w", err)
	}

	return ok, nil
}

func (s *CheckPermissionUseCase) CheckAny(ctx context.Context, roleCode string, permissionCodes []string) (bool, error) {
	// Админу можно все
	if roleCode == "admin" {
		return true, nil
	}

	// Проверка списка прав
	ok, err := s.repo.AuthRedis.CheckAnyPermissions(ctx, roleCode, permissionCodes)
	if err != nil {

		// Список не загружен
		var nek *errs.NotExistKeyRedisError
		if errors.As(err, &nek) {
			return s.loadAndCheckAny(ctx, roleCode, permissionCodes)
		}

		return false, fmt.Errorf("check permission: %w", err)
	}

	return ok, nil
}
