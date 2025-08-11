package entity

import (
	"breast-implant-warranty-system/pkg/dbutil"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/guregu/null/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type Search struct {
	SearchParam null.String `query:"q"`
	// might need to add this in the future,
	// if there had more than one column supported search, also need to make sure the column index are created
	//SearchParamBy  null.String `query:"qBy"`

	// allen: below might need to create new struct to separate from Search,
	// it could be
	// rangeFrom null.Time
	// rangeTo 	 null.Time
	// rangeBy   string
	CreatedAtFrom null.Time `query:"createdAtFrom"` // TODO: sync the time format with the frontend
	CreatedAtTo   null.Time `query:"createdAtTo"`
}

func (req Search) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.SearchParam, validation.NilOrNotEmpty),
		validation.Field(&req.CreatedAtFrom, validation.NilOrNotEmpty),
		validation.Field(&req.CreatedAtTo, validation.NilOrNotEmpty),
	)
}

type Pagination struct {
	Page        int    `query:"page"`
	PageSize    int    `query:"page_size"`
	Offset      int    `query:"offset"`
	Limit       int    `query:"limit"`
	SortBy      string `query:"sortBy"`
	SortOrder   string `query:"sortOrder"`
	DisableSort bool
}

func (p *Pagination) SetDefaultValue() {
	p.Page = 1
	p.PageSize = 10
	p.Offset = 0
	p.Limit = 10
	p.SortBy = "updated_at"
	p.SortOrder = "DESC"
}

func (p Pagination) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Page, validation.Min(1)),
		validation.Field(&p.PageSize, validation.Min(1), validation.Max(100)),
		validation.Field(&p.Offset, validation.Min(0)),
		validation.Field(&p.Limit, validation.Min(1), validation.Max(100)),
		validation.Field(&p.SortOrder, validation.In("DESC", "ASC", "desc", "asc")),
	)
}

func (p Pagination) OrderByClause(table string) bob.Mod[*dialect.SelectQuery] {
	if p.DisableSort {
		return dbutil.NoOpMod[*dialect.SelectQuery]{}
	}

	orderClause := sm.OrderBy(psql.Quote(table, p.SortBy))
	if strings.ToUpper(p.SortOrder) == "ASC" {
		orderClause = orderClause.Asc()
	} else {
		orderClause = orderClause.Desc()
	}
	return orderClause
}

func (p Pagination) OffsetClause() bob.Mod[*dialect.SelectQuery] {
	// turn page and page_size to offset
	if p.Page > 1 || p.PageSize != 10 {
		p.Offset = (p.Page - 1) * p.PageSize
	}
	return sm.Offset(p.Offset)
}

func (p Pagination) LimitClause() bob.Mod[*dialect.SelectQuery] {
	if p.Limit == -1 || p.PageSize == -1 {
		return sm.Limit("ALL")
	} else if p.PageSize != 10 {
		return sm.Limit(p.PageSize)
	}
	return sm.Limit(p.Limit)
}
