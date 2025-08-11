package repositories

import (
	"context"

	"breast-implant-warranty-system/internal/entity"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/pkg/dbutil"

	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
)

// UserRepository 使用者資料庫操作
type UserRepository struct {
	db dbutil.PgxClientItf
}

// NewUserRepository 建立新的使用者倉庫
func NewUserRepository(db dbutil.PgxClientItf) *UserRepository {
	return &UserRepository{db: db}
}

// Create 建立使用者
func (r *UserRepository) Create(ctx context.Context, user *models.User) (string, error) {
	builder := psql.Insert(
		im.Into("users",
			"username",
			"email",
			"password_hash",
			"role",
			"is_active",
		),
		im.Values(psql.Arg(
			user.Username,
			user.Email,
			user.PasswordHash,
			user.Role,
			user.IsActive,
		)),
		im.Returning("id"),
	)

	return dbutil.GetColumn[string](ctx, r.db, builder)
}

// GetByID 根據ID取得使用者
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"username",
			"email",
			"password_hash",
			"role",
			"is_active",
			"last_login_at",
			"created_at",
			"updated_at",
		),
		sm.From("users"),
		sm.Where(psql.Quote("users", "id").EQ(psql.Arg(id))),
		sm.Limit(1),
	)

	return dbutil.GetOne[models.User](ctx, r.db, builder)
}

// GetByUsername 根據使用者名稱取得使用者
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"username",
			"email",
			"password_hash",
			"role",
			"is_active",
			"last_login_at",
			"created_at",
			"updated_at",
		),
		sm.From("users"),
		sm.Where(psql.Quote("users", "username").EQ(psql.Arg(username))),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.User](ctx, r.db, builder)
}

// GetByEmail 根據電子信箱取得使用者
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"username",
			"email",
			"password_hash",
			"role",
			"is_active",
			"last_login_at",
			"created_at",
			"updated_at",
		),
		sm.From("users"),
		sm.Where(psql.Quote("users", "email").EQ(psql.Arg(email))),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.User](ctx, r.db, builder)
}

// Update 更新使用者
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	builder := psql.Update(
		um.Table("users"),
		um.SetCol("username").ToArg(user.Username),
		um.SetCol("email").ToArg(user.Email),
		um.SetCol("password_hash").ToArg(user.PasswordHash),
		um.SetCol("role").ToArg(user.Role),
		um.SetCol("is_active").ToArg(user.IsActive),
		um.SetCol("updated_at").To("NOW()"),
		um.Where(psql.Quote("users", "id").EQ(psql.Arg(user.ID))),
	)
	result, err := dbutil.Exec(ctx, r.db, builder)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete 刪除使用者
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	builder := psql.Delete(
		dm.From("users"),
		dm.Where(psql.Quote("users", "id").EQ(psql.Arg(id))),
	)
	result, err := dbutil.Exec(ctx, r.db, builder)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Search 搜尋使用者
func (r *UserRepository) Search(ctx context.Context, req *models.UserSearchRequest, page *entity.Pagination) ([]*models.User, int, error) {
	conditions := []bob.Expression{}
	if req.Username.Valid {
		conditions = append(conditions,
			psql.Quote("users", "username").ILike(psql.Arg("%"+req.Username.String+"%")),
		)
	}
	if req.Email.Valid {
		conditions = append(conditions,
			psql.Quote("users", "email").ILike(psql.Arg("%"+req.Email.String+"%")),
		)
	}
	if req.Role.Valid {
		conditions = append(conditions,
			psql.Quote("users", "role").EQ(psql.Arg(req.Role.String)),
		)
	}
	if req.IsActive.Valid {
		conditions = append(conditions,
			psql.Quote("users", "is_active").EQ(psql.Arg(req.IsActive.Bool)),
		)
	}

	builder := psql.Select(
		sm.Columns(
			"id",
			"username",
			"email",
			"password_hash",
			"role",
			"is_active",
			"last_login_at",
			"created_at",
			"updated_at",
			"COUNT(id) OVER() AS total_count",
		),
		sm.From("users"),
		sm.OrderBy(psql.Quote("users", "created_at")).Desc(),
		page.OffsetClause(),
	)

	if len(conditions) > 0 {
		builder.Apply(sm.Where(psql.And(conditions...)))
	}

	type UserWithTotalCount struct {
		models.User
		TotalCount int `db:"total_count"`
	}

	ut, _, err := dbutil.GetPage[UserWithTotalCount](ctx, r.db, builder, page.Limit)
	if err != nil {
		return nil, 0, err
	}
	users := make([]*models.User, len(ut))
	for i, u := range ut {
		users[i] = &u.User
	}

	total := 0
	if len(ut) > 0 {
		total = ut[0].TotalCount
	}

	return users, total, nil

}

// UpdateLastLogin 更新最後登入時間
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	builder := psql.Update(
		um.Table("users"),
		um.SetCol("last_login_at").To("NOW()"),
		um.Where(psql.Quote("users", "id").EQ(psql.Arg(id))),
	)
	_, err := dbutil.Exec(ctx, r.db, builder)
	return err
}
