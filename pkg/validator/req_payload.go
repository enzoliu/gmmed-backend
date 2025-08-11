package validator

import (
	"context"
	"reflect"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
)

const (
	requestPayloadKey = "request-payload"
)

var (
	defaultBinder = &echo.DefaultBinder{}
)

type ValidatableItf interface {
	Validate() error
}

type ValidatableWithContextItf interface {
	ValidateWithContext(ctx context.Context) error
}

type DefaultValueSetterItf interface {
	SetDefaultValue()
}

func PayloadMiddleware[T any]() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			var payload T
			if err := Load(c, &payload); err != nil {
				return err
			}

			c.Set(requestPayloadKey, &payload)
			return next(c)
		}
	}
}

func GetPayload[T any](c echo.Context) *T {
	return c.Get(requestPayloadKey).(*T)
}

func Load(c echo.Context, payload any) error {
	payloadCallback := ScanCallback(c, payload)

	// fill default value
	defaultSetter, ok := payload.(DefaultValueSetterItf)
	if ok {
		defaultSetter.SetDefaultValue()
	}

	// bind
	if err := defaultBinder.Bind(payload, c); err != nil {
		return err
	}
	if err := defaultBinder.BindHeaders(c, payload); err != nil {
		return err
	}

	if err := payloadCallback.ExecBind(); err != nil {
		return errors.Wrap(err, "failed to exec bind callbacks")
	}

	// transform
	if err := payloadCallback.ExecTransform(); err != nil {
		return errors.Wrap(err, "failed to exec transform callbacks")
	}

	// validate
	var err error
	switch val := payload.(type) {
	case ValidatableItf:
		err = val.Validate()
	case ValidatableWithContextItf:
		err = val.ValidateWithContext(c.Request().Context())
	default:
		err = Struct(payload)
	}
	if err != nil {
		return err
	}

	return nil
}

func getFieldNameByJSONTag(payload any, fieldName string) string {
	if fieldName == "" {
		return ""
	}

	fieldType, ok := reflect.TypeOf(payload).Elem().FieldByName(fieldName)
	if !ok {
		return fieldName
	}

	tagValue := fieldType.Tag.Get("json")
	if tagValue == "" || tagValue == "-" {
		return fieldName
	}
	return tagValue
}
