package models

type LogMetaVar struct {
	Id      int    `db:"id" json:"id" binding:"required"`
	AppId   int    `db:"app_id" json:"app_id" binding:"required"`
	Name    string `db:"name" json:"name" binding:"required"`
	Code    string `db:"code" json:"code" binding:"required"`
	Type    string `db:"type" json:"type" binding:"required"`
	Deleted bool   `db:"deleted" json:"deleted"`
}
