package testutils

import (
	"fmt"
	"strings"

	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

func NewTestUser(name string) *models.User {
	return &models.User{
		Id:       1,
		Name:     name,
		Username: strings.ToLower(name),
		Email:    utils.ToPtr(fmt.Sprintf("%s@example.ru", name)),
		TgId:     utils.ToPtr(1231232),
		Role: models.Role{
			Id:   1,
			Code: "admin",
			Name: "Админ",
		},
		CreatedAt: &FixedTime,
		UpdatedAt: &FixedTime,
		Deleted:   utils.ToPtr(false),
	}
}
