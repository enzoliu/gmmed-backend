package validator

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
)

var (
	transformerRegistry = map[string]TransformFunc{
		"case_insensitive": CaseInsensitiveTransformer,
	}
)

type TransformFunc func(fieldValue reflect.Value) error

type Nullable[T any] interface {
	IsZero() bool
	ValueOrZero() T
}

type UnmarshalNullable[T any] interface {
	encoding.TextUnmarshaler
	Nullable[T]
}

func RegisterTransformer(name string, transformer TransformFunc) {
	transformerRegistry[name] = transformer
}

func getTransformCallback(tagValue string, fieldValue reflect.Value) CallbackFunc {
	transformFunc, ok := transformerRegistry[tagValue]
	if !ok {
		panic(fmt.Sprintf("transform function not found for tag '%s'", tagValue))
	}

	return func() error {
		return transformFunc(fieldValue)
	}
}

func CaseInsensitiveTransformer(fieldValue reflect.Value) error {
	if fieldValue.Kind() == reflect.String {
		newStr := strings.ToLower(fieldValue.String())
		fieldValue.SetString(newStr)

	} else if fieldValue.Kind() == reflect.Ptr && fieldValue.Elem().Kind() == reflect.String {
		newStr := strings.ToLower(fieldValue.Elem().String())
		fieldValue.Elem().SetString(newStr)

	} else if nullable, ok := fieldValue.Addr().Interface().(UnmarshalNullable[string]); ok {
		if !nullable.IsZero() {
			newStr := strings.ToLower(nullable.ValueOrZero())
			return nullable.UnmarshalText([]byte(newStr))
		}

	} else {
		panic(fmt.Sprintf("unsupported type '%T'", fieldValue.Interface()))
	}

	return nil
}
