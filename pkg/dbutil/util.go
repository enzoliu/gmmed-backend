package dbutil

import (
	"fmt"
	"strings"

	"github.com/stephenafamo/bob/dialect/psql"
)

func EscapeWildcard(s string) string {
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "%", "\\%")
	return s
}

func Column(table, column string) string {
	if table != "" {
		column = fmt.Sprintf("%s.%s", table, column)
	}
	return column
}

func Rawf(format string, args ...any) psql.Expression {
	return psql.Raw(fmt.Sprintf(format, args...))
}

type NoOpMod[T any] struct{}

func (NoOpMod[T]) Apply(T) {}
