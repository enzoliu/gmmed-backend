package validator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
)

type bindType string

const (
	bindTypeMultiQuery bindType = "multi_query"
)

func getBindCallback(ec echo.Context, tagValue string, fieldValue reflect.Value) CallbackFunc {
	segs := strings.Split(tagValue, "=")
	if len(segs) != 2 {
		panic("invalid value for bindx tag")
	}
	bindType := bindType(segs[0])
	paramName := segs[1]

	if paramName == "" {
		return nil
	}

	switch bindType {
	case bindTypeMultiQuery:
		return getBindMultiQueryCallback(ec, paramName, fieldValue)
	default:
		panic("invalid bind type for bindx tag")
	}
}

func getBindMultiQueryCallback(ec echo.Context, paramName string, fieldValue reflect.Value) CallbackFunc {
	return func() error {
		fieldPtr := fieldValue.Addr().Interface()

		values := ec.QueryParam(paramName)

		if values == "" {
			return nil
		}

		if err := json.Unmarshal([]byte(values), fieldPtr); err != nil {
			slog.WarnContext(ec.Request().Context(), "failed to unmarshal multi query", "err", err.Error(), "key", paramName, "value", values)
			return fmt.Errorf("failed to unmarshal multi query '%s': %s", paramName, err.Error())
		}
		return nil
	}
}
