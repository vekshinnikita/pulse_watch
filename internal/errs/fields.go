package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type StructuredError interface {
	error
	GetStructuredError() *StructuredFieldError
	GetStructuredErrors() StructuredFieldErrors
}

type StructuredFieldError struct {
	Field   string         `json:"field"`
	Type    FieldErrorType `json:"type"`
	Message string         `json:"message"`
}

type StructuredFieldErrors []StructuredFieldError

type FieldErrorType string

const (
	UniqueFieldErrorType   FieldErrorType = "unique"
	NotFoundFieldErrorType FieldErrorType = "not_found"
	TypeFieldErrorType     FieldErrorType = "type"
	RequiredFieldErrorType FieldErrorType = "required"
	EmailFieldErrorType    FieldErrorType = "email"
	LteFieldErrorType      FieldErrorType = "gte"
	GteFieldErrorType      FieldErrorType = "lte"
	MinFieldErrorType      FieldErrorType = "min"
	MaxFieldErrorType      FieldErrorType = "max"
	PasswordFieldErrorType FieldErrorType = "password"
	FutureFieldErrorType   FieldErrorType = "future"
	EnumFieldErrorType     FieldErrorType = "enum"
	OneOfFieldErrorType    FieldErrorType = "oneof"
	DatetimeFieldErrorType FieldErrorType = "datetime"
)
const (
	InvalidValueFieldErrorMessage    string = "Invalid value"
	UniqueFieldErrorMessage          string = "This %s already exists"
	TypeFieldErrorMessage            string = "Value must be %s type"
	RequiredFieldErrorMessage        string = "This field is required"
	EmailFieldErrorMessage           string = "Invalid email format"
	LteFieldErrorMessage             string = "Value must be greater than or equal to %v"
	GteFieldErrorMessage             string = "Value must be less than or equal to %v"
	MinFieldErrorMessage             string = "Value length must be greater %v"
	MaxFieldErrorMessage             string = "Value length must be less %v"
	PasswordFieldErrorMessage        string = "The password must contain at least 8 characters, including letters and numbers."
	FutureFieldErrorMessage          string = "The date must be greater then current time"
	EnumFieldErrorMessage            string = "Value must be one of values: %v"
	EnumUnknownEnumFieldErrorMessage string = "Value must be one of value enum: %v"
	DatetimeFieldErrorMessage        string = "Value must be in '%s' format"
)

var errorTypeTranslation = map[string]FieldErrorType{}

type UniqueFieldError struct {
	Field string
}

func (e *UniqueFieldError) Error() string {
	return e.Field + " must be unique"
}

func (e *UniqueFieldError) GetStructuredError() *StructuredFieldError {
	return &StructuredFieldError{
		Field:   e.Field,
		Type:    UniqueFieldErrorType,
		Message: fmt.Sprintf(UniqueFieldErrorMessage, e.Field),
	}
}

func (e *UniqueFieldError) GetStructuredErrors() StructuredFieldErrors {
	return StructuredFieldErrors{*e.GetStructuredError()}
}

type NotFoundFieldError struct {
	Message string
	Field   string
}

func (e *NotFoundFieldError) Error() string {
	return e.Message
}

func (e *NotFoundFieldError) GetStructuredError() *StructuredFieldError {
	return &StructuredFieldError{
		Field:   e.Field,
		Type:    NotFoundFieldErrorType,
		Message: e.Message,
	}
}

func (e *NotFoundFieldError) GetStructuredErrors() StructuredFieldErrors {
	return StructuredFieldErrors{*e.GetStructuredError()}
}

type TypeFieldError struct {
	Message string
	Field   string
}

func (e *TypeFieldError) Error() string {
	return e.Message
}

func (e *TypeFieldError) GetStructuredError() *StructuredFieldError {
	return &StructuredFieldError{
		Field:   e.Field,
		Type:    TypeFieldErrorType,
		Message: e.Message,
	}
}

func (e *TypeFieldError) GetStructuredErrors() StructuredFieldErrors {
	return StructuredFieldErrors{*e.GetStructuredError()}
}

func getJSONFieldName(input interface{}, fe validator.FieldError) string {
	re := regexp.MustCompile(`(\w+)(\[(\d+)\])?`)
	parts := strings.Split(fe.StructNamespace(), ".")

	// Пропускаем верхний уровень (SendLogs)
	parts = parts[1:]

	var jsonParts []string
	objType := reflect.TypeOf(utils.ToInterface(input))

	for _, part := range parts {
		matches := re.FindStringSubmatch(part)
		fieldName := matches[1]
		index := matches[3]

		// Получаем JSON-тег
		if objType.Kind() == reflect.Slice {
			objType = objType.Elem()
		}
		field, _ := objType.FieldByName(fieldName)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]

		jsonParts = append(jsonParts, jsonTag)
		if index != "" {
			jsonParts = append(jsonParts, index)
		}

		objType = field.Type
	}

	return strings.Join(jsonParts, ".")
}

func GetTypeError(fe validator.FieldError) FieldErrorType {
	value, ok := errorTypeTranslation[fe.Tag()]
	if ok {
		return value
	}

	return FieldErrorType(fe.Tag())
}

func ValidationErrorToText(fe validator.FieldError) string {
	switch GetTypeError(fe) {
	case RequiredFieldErrorType:
		return RequiredFieldErrorMessage
	case EmailFieldErrorType:
		return EmailFieldErrorMessage
	case LteFieldErrorType:
		return fmt.Sprintf(LteFieldErrorMessage, fe.Param())
	case GteFieldErrorType:
		return fmt.Sprintf(GteFieldErrorMessage, fe.Param())
	case MinFieldErrorType:
		return fmt.Sprintf(MinFieldErrorMessage, fe.Param())
	case MaxFieldErrorType:
		return fmt.Sprintf(MaxFieldErrorMessage, fe.Param())
	case TypeFieldErrorType:
		return fmt.Sprintf(TypeFieldErrorMessage, fe.Param())
	case PasswordFieldErrorType:
		return PasswordFieldErrorMessage
	case FutureFieldErrorType:
		return FutureFieldErrorMessage

	case OneOfFieldErrorType:
		lambda := func(v string) string { return fmt.Sprintf(`'%s'`, v) }
		params := strings.Join(utils.Map(strings.Fields(fe.Param()), lambda), ", ")
		return fmt.Sprintf(EnumFieldErrorMessage, params)

	case EnumFieldErrorType:
		param := fe.Param()
		enumSet, ok := constants.EnumSetsMap[param]
		if !ok {
			return fmt.Sprintf(EnumUnknownEnumFieldErrorMessage, param)
		}

		lambda := func(v any) string { return fmt.Sprintf(`'%s'`, v) }
		enumValues := utils.MapKeys(enumSet)

		valuesString := strings.Join(utils.Map(enumValues, lambda), ", ")
		return fmt.Sprintf(EnumFieldErrorMessage, valuesString)

	case DatetimeFieldErrorType:
		return fmt.Sprintf(DatetimeFieldErrorMessage, fe.Param())
	}

	return InvalidValueFieldErrorMessage
}

func makeErrorsMapFromValidationErrors(input interface{}, errors validator.ValidationErrors) StructuredFieldErrors {
	structuredErrors := make(StructuredFieldErrors, 0)
	for _, fe := range errors {

		structuredErrors = append(structuredErrors, StructuredFieldError{
			Field:   getJSONFieldName(input, fe),
			Type:    GetTypeError(fe),
			Message: ValidationErrorToText(fe),
		})
	}

	return structuredErrors
}

func makeErrorsFromUnmarshalTypeError(input interface{}, err *json.UnmarshalTypeError) StructuredFieldErrors {
	structuredErrors := make(StructuredFieldErrors, 0)

	structuredErrors = append(structuredErrors, StructuredFieldError{
		Field:   err.Field,
		Type:    TypeFieldErrorType,
		Message: fmt.Sprintf(TypeFieldErrorMessage, err.Type),
	})

	return structuredErrors
}

func IsValidationError(err error) bool {

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return true
	}

	var ue *json.UnmarshalTypeError
	if errors.As(err, &ue) {
		return true
	}

	return false
}

func MakeFieldsErrorMap(input interface{}, err error) StructuredFieldErrors {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return makeErrorsMapFromValidationErrors(input, ve)
	}

	var ue *json.UnmarshalTypeError
	if errors.As(err, &ue) {
		return makeErrorsFromUnmarshalTypeError(input, ue)
	}

	return nil
}
