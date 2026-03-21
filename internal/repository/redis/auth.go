package redis_repository

import (
	"context"
	"fmt"
	"time"

	trmRedis "github.com/avito-tech/go-transaction-manager/drivers/goredis8/v2"
	"github.com/go-redis/redis/v8"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type AuthRedis struct {
	client *redis.Client
	getter *trmRedis.CtxGetter
}

func NewAuthRedis(
	client *redis.Client,
	getter *trmRedis.CtxGetter,
) *AuthRedis {
	return &AuthRedis{
		client: client,
		getter: getter,
	}
}

func (r *AuthRedis) LoadPermissions(
	ctx context.Context,
	roleCode string,
	permissions []models.Permission,
) error {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	pipe := executor.Pipeline()

	key := fmt.Sprintf("permissions:%s", roleCode)

	// Формируем список значений
	members := make([]interface{}, len(permissions))
	for i, perm := range permissions {
		members[i] = interface{}(perm.Code)
	}

	// Добавляем элемент заглушку
	if len(members) == 0 {
		members = append(members, interface{}(0))
	}

	// Добавляем ключи в set
	pipe.SAdd(ctx, key, members...)

	// Установка TTL
	pipe.Expire(ctx, key, time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("run pipe: %w", err)
	}

	return nil
}

func (r *AuthRedis) CheckPermission(
	ctx context.Context,
	roleCode string,
	permissionCode string,
) (bool, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	pipe := executor.Pipeline()

	key := fmt.Sprintf("permissions:%s", roleCode)
	// Проверка список есть в redis
	exists := pipe.Exists(ctx, key)

	// Проверка, право есть в списке
	sismember := pipe.SIsMember(ctx, key, "delete_user")

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("run pipe: %w", err)
	}

	hasPerm := sismember.Val()
	keyExists := exists.Val()

	if keyExists == 0 {
		return false, &errs.NotExistKeyRedisError{Message: fmt.Sprintf("redis key %s doesn't exists", key)}
	}

	return hasPerm, nil
}

func (r *AuthRedis) CheckAnyPermissions(
	ctx context.Context,
	roleCode string,
	permissionCodes []string,
) (bool, error) {
	executor := r.getter.DefaultTrOrDB(ctx, r.client)

	script := `
	local key = KEYS[1]

	local is_exists = redis.call("EXISTS", key)
	if is_exists == 0 then
		return -1
	end

	for _, perm_code in ipairs(ARGV) do
		local has_perm = redis.call("SISMEMBER", key, perm_code)

		if has_perm == 1 then
			return 1
		end
	end

	return 0
	`

	key := fmt.Sprintf("permissions:%s", roleCode)
	args := make([]interface{}, len(permissionCodes))
	for i, v := range permissionCodes {
		args[i] = v
	}

	cmd := executor.Eval(
		ctx,
		script,
		[]string{key},
		args...,
	)
	res, err := cmd.Result()
	if err != nil {
		return false, fmt.Errorf("run script: %w", err)
	}

	val, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("can't cast result to int")
	}

	// нет списка прав
	if val == -1 {
		return false, &errs.NotExistKeyRedisError{Message: fmt.Sprintf("redis key %s doesn't exists", key)}
	}

	// права есть
	if val == 1 {
		return true, nil
	}

	// прав нет
	return false, nil
}
