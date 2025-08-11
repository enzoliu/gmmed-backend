package repositories

import (
	"breast-implant-warranty-system/internal/entity"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/dbutil"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type AuditRepository struct {
	db dbutil.PgxClientItf
}

func NewAuditRepository(db dbutil.PgxClientItf) *AuditRepository {
	return &AuditRepository{db: db}
}

// List 取得審計日誌列表
func (r *AuditRepository) List(ctx context.Context, req *models.AuditSearchLogRequest, page *entity.Pagination) ([]*models.AuditLog, int, error) {
	condition := []bob.Expression{}

	if req.UserName.Valid {
		condition = append(condition,
			psql.Quote("users", "username").ILike(psql.Arg("%"+req.UserName.String+"%")),
		)
	}
	if req.Action.Valid {
		condition = append(condition,
			psql.Quote("audit_logs", "action").EQ(psql.Arg(req.Action.String)),
		)
	}
	if req.TableName.Valid {
		condition = append(condition,
			psql.Quote("audit_logs", "table_name").EQ(psql.Arg(req.TableName.String)),
		)
	}
	if req.IPAddress.Valid {
		condition = append(condition,
			psql.Quote("audit_logs", "ip_address").EQ(psql.Arg(req.IPAddress.String)),
		)
	}
	if req.StartDate.Valid {
		parsedStartDate, err := utils.ParseTaiwanDateToUTC(req.StartDate.String)
		if err != nil {
			return nil, 0, err
		}
		condition = append(condition,
			psql.Quote("audit_logs", "created_at").GTE(psql.Arg(parsedStartDate)),
		)
	}
	if req.EndDate.Valid {
		parsedEndDate, err := utils.ParseTaiwanDateToUTC(req.EndDate.String + " 23:59:59")
		if err != nil {
			return nil, 0, err
		}
		condition = append(condition,
			psql.Quote("audit_logs", "created_at").LTE(psql.Arg(parsedEndDate)),
		)
	}

	builder := psql.Select(
		sm.Columns(
			"audit_logs.id",
			"audit_logs.user_id",
			"audit_logs.action",
			"audit_logs.table_name",
			"audit_logs.record_id",
			"audit_logs.old_values",
			"audit_logs.new_values",
			"audit_logs.ip_address",
			"audit_logs.user_agent",
			"audit_logs.created_at",
			"users.username",
			"users.email",
			"COUNT(audit_logs.id) OVER() AS total_count",
		),
		sm.From("audit_logs"),
		sm.LeftJoin("users").On(psql.Raw("audit_logs.user_id = users.id")),
		sm.OrderBy(psql.Quote("audit_logs", "created_at")).Desc(),
		page.OffsetClause(),
	)
	if len(condition) > 0 {
		builder.Apply(sm.Where(psql.And(condition...)))
	}

	type AuditLogWithTotalCount struct {
		models.AuditLog
		TotalCount int `db:"total_count"`
	}

	logs, _, err := dbutil.GetPage[AuditLogWithTotalCount](ctx, r.db, builder, page.Limit)
	if err != nil {
		return nil, 0, err
	}

	auditLogs := make([]*models.AuditLog, len(logs))
	for i, log := range logs {
		auditLogs[i] = &log.AuditLog
	}

	total := 0
	if len(logs) > 0 {
		total = logs[0].TotalCount
	}

	return auditLogs, total, nil
}

// GetByID 根據ID取得審計日誌
func (r *AuditRepository) GetByID(ctx context.Context, id string) (*models.AuditLog, error) {
	builder := psql.Select(
		sm.Columns(
			"audit_logs.id",
			"audit_logs.user_id",
			"audit_logs.action",
			"audit_logs.table_name",
			"audit_logs.record_id",
			"audit_logs.old_values",
			"audit_logs.new_values",
			"audit_logs.ip_address",
			"audit_logs.user_agent",
			"audit_logs.created_at",
			"users.username",
			"users.email",
		),
		sm.From("audit_logs"),
		sm.LeftJoin("users").On(psql.Raw("audit_logs.user_id = users.id")),
		sm.Where(psql.Quote("audit_logs", "id").EQ(psql.Arg(id))),
		sm.Limit(1),
	)
	al, err := dbutil.GetOne[models.AuditLog](ctx, r.db, builder)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// 將表格名稱替換為中文
	al.TableName = models.TableNameMapping[al.TableName]

	return al, nil
}

// Create 建立審計日誌
func (r *AuditRepository) Create(ctx context.Context, req *models.CreateAuditLogRequest) error {
	builder := psql.Insert(
		im.Into("audit_logs",
			"user_id",
			"action",
			"table_name",
			"record_id",
			"old_values",
			"new_values",
			"ip_address",
			"user_agent",
		),
		im.Values(psql.Arg(
			req.UserID,
			req.Action,
			req.TableName,
			req.RecordID,
			req.OldValues,
			req.NewValues,
			req.IPAddress,
			req.UserAgent,
		)),
	)
	return dbutil.ShouldExec(ctx, r.db, builder)
}
