package utils

import "reflect"

func ToPtr[T any](a T) *T {
	return &a
}

func IsPointer(input any) bool {
	rv := reflect.ValueOf(input)

	// Если указатель — разыменуем
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return rv.Kind() == reflect.Struct
}

func ToInterface(input any) any {
	v := reflect.ValueOf(input)

	// Разыменовываем указатель, если это Ptr
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			panic("nil pointer passed")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("input is not a struct or pointer to struct")
	}

	return v.Interface()
}

func CopyPtr[T any](src *T) *T {
	if src == nil {
		return nil
	}
	dst := *src
	return &dst
}
