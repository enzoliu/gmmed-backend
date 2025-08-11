package validator

import (
	"reflect"

	"github.com/labstack/echo/v4"
)

const (
	tagTransform = "transform"
	tagBindx     = "bindx" // add x to avoid conflict with framework tag
)

type CallbackFunc func() error

type PayloadCallback struct {
	DefaultValueCallbacks []CallbackFunc
	BindCallbacks         []CallbackFunc
	TransformCallbacks    []CallbackFunc
}

func ScanCallback(ec echo.Context, payload any) *PayloadCallback {
	pc := &PayloadCallback{}

	payloadT := reflect.TypeOf(payload).Elem()
	payloadV := reflect.ValueOf(payload).Elem()

	for _, field := range reflect.VisibleFields(payloadT) {
		fieldValue := payloadV.FieldByName(field.Name)

		if tagValue := field.Tag.Get(tagBindx); tagValue != "" {
			callback := getBindCallback(ec, tagValue, fieldValue)
			if callback != nil {
				pc.BindCallbacks = append(pc.BindCallbacks, callback)
			}
		}

		if tagValue := field.Tag.Get(tagTransform); tagValue != "" {
			callback := getTransformCallback(tagValue, fieldValue)
			if callback != nil {
				pc.TransformCallbacks = append(pc.TransformCallbacks, callback)
			}
		}
	}

	return pc
}

func (pc *PayloadCallback) ExecBind() error {
	for _, callback := range pc.BindCallbacks {
		if err := callback(); err != nil {
			return err
		}
	}
	return nil
}

func (pc *PayloadCallback) ExecTransform() error {
	for _, callback := range pc.TransformCallbacks {
		if err := callback(); err != nil {
			return err
		}
	}
	return nil
}
