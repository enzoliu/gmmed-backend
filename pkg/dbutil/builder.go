package dbutil

import "context"

var _ QueryBuilderItf = (*RawQuery)(nil)

type RawQuery struct {
	Query string
	Args  []any
}

func (rq RawQuery) Build(ctx context.Context) (string, []any, error) {
	return rq.Query, rq.Args, nil
}

func NewRawQuery(query string, args ...any) *RawQuery {
	return &RawQuery{
		Query: query,
		Args:  args,
	}
}
