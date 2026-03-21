package dtos

import "github.com/vekshinnikita/pulse_watch/internal/constants"

type CreateMetaVar struct {
	AppId int
	Name  string
	Code  string
	Type  constants.MetaVarType
}
