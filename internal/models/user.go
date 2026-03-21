package models

import "time"

type Role struct {
	Id   int    `db:"id" json:"id" binding:"required"`
	Code string `db:"code" json:"code" binding:"required"`
	Name string `db:"name" json:"name" binding:"required"`
}

type Permission struct {
	Id   int    `db:"id" json:"id" binding:"required"`
	Code string `db:"code" json:"code" binding:"required"`
	Name string `db:"name" json:"name" binding:"required"`
}

type User struct {
	Id        int        `db:"id" json:"id" binding:"required"`
	Name      string     `db:"name" json:"name" binding:"required"`
	Username  string     `db:"username" json:"username" binding:"required"`
	Email     *string    `db:"email" json:"email"`
	TgId      *int       `db:"tg_id" json:"tg_id"`
	Role      Role       `db:"role" json:"role"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
	Deleted   *bool      `db:"deleted" json:"deleted"`
}

type UserWithPermissions struct {
	User
	PermissionsMap map[string]Permission `json:"permissions"`
}
