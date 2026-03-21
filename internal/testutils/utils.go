package testutils

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
)

func MarshalTestJson[T any](t *testing.T, data T) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("error when marshaling data structure: %s", err.Error())
	}

	return string(jsonBytes)
}

func safeReflectValue(v reflect.Value) any {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return v.Elem().Interface()
	}

	return v.Interface()
}

func extractDBFields(prefix string, v reflect.Value) ([]string, []driver.Value) {

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	var fields []string
	var values []driver.Value

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		dbTag := f.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		column := prefix + dbTag

		// Вложенная структура
		if fv.Kind() == reflect.Struct && f.Anonymous == false {
			subFields, subValues := extractDBFields(column+".", fv)
			fields = append(fields, subFields...)
			values = append(values, subValues...)
			continue
		}

		fields = append(fields, column)
		values = append(values, safeReflectValue(fv))
	}

	return fields, values
}

func ExtractDBFields(data any) ([]string, []driver.Value) {
	v := reflect.ValueOf(data)
	return extractDBFields("", v)
}
