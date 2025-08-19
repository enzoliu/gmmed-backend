package repositories

import (
	"context"
	"fmt"

	"breast-implant-warranty-system/internal/entity"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/pkg/dbutil"

	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
)

// SerialRepository 序號資料庫操作
type SerialRepository struct {
	db dbutil.PgxClientItf
}

// NewSerialRepository 建立新的序號倉庫
func NewSerialRepository(db dbutil.PgxClientItf) *SerialRepository {
	return &SerialRepository{db: db}
}

// Create 建立序號
func (r *SerialRepository) Create(ctx context.Context, req *models.SerialCreateRequest) (string, error) {
	builder := psql.Insert(
		im.Into("serials",
			"serial_number",
			"full_serial_number",
			"product_id",
		),
		im.Values(psql.Arg(
			req.SerialNumber,
			req.FullSerialNumber,
			req.ProductID,
		)),
		im.Returning("id"),
	)

	return dbutil.GetColumn[string](ctx, r.db, builder)
}

// GetByID 根據ID取得序號
func (r *SerialRepository) GetByID(ctx context.Context, id string) (*models.Serial, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"serial_number",
			"full_serial_number",
			"product_id",
			"created_at",
			"updated_at",
		),
		sm.From("serials"),
		sm.Where(psql.And(
			psql.Quote("serials", "id").EQ(psql.Arg(id)),
			psql.Quote("serials", "deleted_at").IsNull(),
		)),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.Serial](ctx, r.db, builder)
}

// GetBySerialNumber 根據序號取得序號資訊
func (r *SerialRepository) GetBySerialNumber(ctx context.Context, serialNumber string) (*models.Serial, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"serial_number",
			"full_serial_number",
			"product_id",
			"created_at",
			"updated_at",
		),
		sm.From("serials"),
		sm.Where(
			psql.And(
				psql.Quote("serials", "serial_number").EQ(psql.Arg(serialNumber)),
				psql.Quote("serials", "deleted_at").IsNull(),
			),
		),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.Serial](ctx, r.db, builder)
}

// GetByFullSerialNumber 根據完整序號取得序號資訊
func (r *SerialRepository) GetByFullSerialNumber(ctx context.Context, fullSerialNumber string) (*models.Serial, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"serial_number",
			"full_serial_number",
			"product_id",
			"created_at",
			"updated_at",
		),
		sm.From("serials"),
		sm.Where(
			psql.And(
				psql.Quote("serials", "full_serial_number").EQ(psql.Arg(fullSerialNumber)),
				psql.Quote("serials", "deleted_at").IsNull(),
			),
		),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.Serial](ctx, r.db, builder)
}

// ExistsBySerialNumber 檢查序號是否存在
func (r *SerialRepository) ExistsBySerialNumber(ctx context.Context, serialNumber string) (string, error) {
	builder := psql.Select(
		sm.Columns("id", "serial_number", "full_serial_number", "product_id", "created_at", "updated_at"),
		sm.From("serials"),
		sm.Where(
			psql.And(
				psql.Quote("serials", "serial_number").EQ(psql.Arg(serialNumber)),
				psql.Quote("serials", "deleted_at").IsNull(),
			),
		),
		sm.Limit(1),
	)

	exists, err := dbutil.GetOne[models.Serial](ctx, r.db, builder)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if exists == nil {
		return "", nil
	}
	if exists.ProductID.Valid {
		return exists.ProductID.String, nil
	}

	return "", nil
}

// ExistsByFullSerialNumber 檢查完整序號是否存在
func (r *SerialRepository) ExistsByFullSerialNumber(ctx context.Context, fullSerialNumber string) (bool, error) {
	builder := psql.Select(
		sm.Columns("1"),
		sm.From("serials"),
		sm.Where(
			psql.And(
				psql.Quote("serials", "full_serial_number").EQ(psql.Arg(fullSerialNumber)),
				psql.Quote("serials", "deleted_at").IsNull(),
			),
		),
		sm.Limit(1),
	)

	exists, err := dbutil.GetOne[int](ctx, r.db, builder)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists != nil, nil
}

// Update 更新序號
func (r *SerialRepository) Update(ctx context.Context, id string, req *models.SerialUpdateRequest) error {
	builder := psql.Update(
		um.Table("serials"),
		um.SetCol("updated_at").To("NOW()"),
		um.Where(psql.Quote("serials", "id").EQ(psql.Arg(id))),
	)

	if req.SerialNumber.Valid {
		builder.Apply(um.SetCol("serial_number").ToArg(req.SerialNumber.String))
	}
	if req.FullSerialNumber.Valid {
		builder.Apply(um.SetCol("full_serial_number").ToArg(req.FullSerialNumber.String))
	}
	if req.ProductID.Valid {
		builder.Apply(um.SetCol("product_id").ToArg(req.ProductID.String))
	}

	result, err := dbutil.Exec(ctx, r.db, builder)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete 軟性刪除序號
func (r *SerialRepository) Delete(ctx context.Context, id string) error {
	builder := psql.Update(
		um.Table("serials"),
		um.SetCol("deleted_at").To("NOW()"),
		um.Where(psql.Quote("serials", "id").EQ(psql.Arg(id))),
	)

	return dbutil.ShouldExec(ctx, r.db, builder)
}

// BulkCreate 大量創建序號
func (r *SerialRepository) BulkCreate(ctx context.Context, req *models.SerialBulkImportRequest, response *models.SerialBulkImportResponse) error {
	if len(req.Serials) == 0 {
		return nil
	}

	// 移除已經存在的序號
	existingSerials, err := r.ListDuplicateSerials(ctx, req)
	if err != nil {
		return fmt.Errorf("處理失敗，原因: %s", err.Error())
	}
	failedCnt := len(response.FailedItems)
	keepSerials := []models.SerialImportItem{}
	for _, reqSerial := range req.Serials {
		exist := false
		for _, existingSerial := range existingSerials {
			if reqSerial.SerialNumber == existingSerial.SerialNumber || reqSerial.FullSerialNumber == existingSerial.FullSerialNumber {
				response.FailedItems = append(response.FailedItems, models.SerialImportErrorItem{
					Index:            reqSerial.Index,
					ProductID:        reqSerial.ProductID,
					SerialNumber:     reqSerial.SerialNumber,
					FullSerialNumber: reqSerial.FullSerialNumber,
					Error:            "該序號已存在。",
				})
				failedCnt++
				exist = true
				break
			}
		}
		if !exist {
			keepSerials = append(keepSerials, reqSerial)
		}
	}

	// 開始事務
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("處理失敗，原因: %s", err.Error())
	}
	defer tx.Rollback(ctx)

	// 添加所有值
	for _, serial := range keepSerials {
		// 準備批量插入
		builder := psql.Insert(
			im.Into("serials",
				"serial_number",
				"full_serial_number",
				"product_id",
			),
			im.Values(psql.Arg(
				serial.SerialNumber,
				serial.FullSerialNumber,
				serial.ProductID,
			)),
		)
		// 執行批量插入
		_, err = dbutil.Exec(ctx, tx, builder)
		if err != nil {
			response.FailedItems = append(response.FailedItems, models.SerialImportErrorItem{
				Index:            serial.Index,
				ProductID:        serial.ProductID,
				SerialNumber:     serial.SerialNumber,
				FullSerialNumber: serial.FullSerialNumber,
				Error:            err.Error(),
			})
			failedCnt++
		}
	}
	// 提交事務
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("處理失敗，原因: %s", err.Error())
	}

	// 全部成功
	response.SuccessCount = int64(len(keepSerials))
	response.FailedCount += int64(len(existingSerials))

	return nil
}

// ListDuplicateSerials 列出req中已經存在於資料庫中的序號
func (r *SerialRepository) ListDuplicateSerials(ctx context.Context, req *models.SerialBulkImportRequest) ([]*models.Serial, error) {
	if len(req.Serials) == 0 {
		return []*models.Serial{}, nil
	}

	const batchSize = 100 // 每批處理 100 個序號

	// 查詢已存在的序號
	duplicateSerials := make([]*models.Serial, 0)

	// 分批處理 serial_number
	for i := 0; i < len(req.Serials); i += batchSize {
		end := i + batchSize
		if end > len(req.Serials) {
			end = len(req.Serials)
		}

		batch := req.Serials[i:end]
		serialNumbers := make([]any, 0, len(batch))
		for _, serial := range batch {
			serialNumbers = append(serialNumbers, serial.SerialNumber)
		}

		if len(serialNumbers) > 0 {
			serialBuilder := psql.Select(
				sm.Columns(
					"id",
					"serial_number",
					"full_serial_number",
					"product_id",
					"created_at",
					"updated_at",
				),
				sm.From("serials"),
				sm.Where(psql.Quote("serials", "serial_number").In(psql.Arg(serialNumbers...))),
			)

			existingBySerial, err := dbutil.GetAll[models.Serial](ctx, r.db, serialBuilder)
			if err != nil {
				return nil, fmt.Errorf("failed to query existing serial numbers batch %d-%d: %w", i, end-1, err)
			}
			duplicateSerials = append(duplicateSerials, existingBySerial...)
		}
	}

	// 分批處理 full_serial_number
	for i := 0; i < len(req.Serials); i += batchSize {
		end := i + batchSize
		if end > len(req.Serials) {
			end = len(req.Serials)
		}

		batch := req.Serials[i:end]
		fullSerialNumbers := make([]any, 0, len(batch))
		for _, serial := range batch {
			fullSerialNumbers = append(fullSerialNumbers, serial.FullSerialNumber)
		}

		if len(fullSerialNumbers) > 0 {
			fullSerialBuilder := psql.Select(
				sm.Columns(
					"id",
					"serial_number",
					"full_serial_number",
					"product_id",
					"created_at",
					"updated_at",
				),
				sm.From("serials"),
				sm.Where(psql.Quote("serials", "full_serial_number").In(psql.Arg(fullSerialNumbers...))),
			)

			existingByFullSerial, err := dbutil.GetAll[models.Serial](ctx, r.db, fullSerialBuilder)
			if err != nil {
				return nil, fmt.Errorf("failed to query existing full serial numbers batch %d-%d: %w", i, end-1, err)
			}
			duplicateSerials = append(duplicateSerials, existingByFullSerial...)
		}
	}

	// 去重（因為同一個序號可能同時在 serial_number 和 full_serial_number 中出現）
	uniqueDuplicates := make(map[string]*models.Serial)
	for _, serial := range duplicateSerials {
		uniqueDuplicates[serial.ID] = serial
	}

	// 轉換回 slice
	result := make([]*models.Serial, 0, len(uniqueDuplicates))
	for _, serial := range uniqueDuplicates {
		result = append(result, serial)
	}

	return result, nil
}

// Search 搜尋序號
func (r *SerialRepository) Search(ctx context.Context, req *models.SerialSearchRequest, page *entity.Pagination) ([]*models.Serial, int, error) {
	conditions := []bob.Expression{}

	if req.ID.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "id").EQ(psql.Arg(req.ID.String)),
		)
	}
	if req.SerialNumber.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "serial_number").ILike(psql.Arg("%"+req.SerialNumber.String+"%")),
		)
	}
	if req.FullSerialNumber.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "full_serial_number").ILike(psql.Arg("%"+req.FullSerialNumber.String+"%")),
		)
	}
	if req.ProductID.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "product_id").EQ(psql.Arg(req.ProductID.String)),
		)
	}

	// 預設不搜尋已刪除的序號
	deleteCondition := psql.Quote("serials", "deleted_at").IsNull()
	if req.SearchDeleted.Valid && req.SearchDeleted.Bool {
		deleteCondition = psql.Quote("serials", "deleted_at").IsNotNull()
	}
	conditions = append(conditions, deleteCondition)

	builder := psql.Select(
		sm.Columns(
			"id",
			"serial_number",
			"full_serial_number",
			"product_id",
			"created_at",
			"updated_at",
			"COUNT(id) OVER() AS total_count",
		),
		sm.From("serials"),
	)

	if req.IsUsedByWarranty.Valid {
		builder.Apply(sm.LeftJoin("warranty_registrations").As("w").On(
			psql.Or(
				psql.Quote("w", "product_serial_number").EQ(psql.Quote("serials", "serial_number")),
				psql.Quote("w", "serial_number_2").EQ(psql.Quote("serials", "serial_number")),
			),
		))
		if req.IsUsedByWarranty.Bool {
			builder.Apply(sm.Where(psql.Quote("w", "id").IsNotNull()))
		} else {
			builder.Apply(sm.Where(psql.Quote("w", "id").IsNull()))
		}
	}

	if len(conditions) > 0 {
		builder.Apply(sm.Where(psql.And(conditions...)))
	}
	builder.Apply(
		page.OrderByClause("serials"),
		page.OffsetClause(),
		page.LimitClause(),
	)

	type SerialWithTotalCount struct {
		models.Serial
		TotalCount int `db:"total_count"`
	}

	swtts, _, err := dbutil.GetPage[SerialWithTotalCount](ctx, r.db, builder, page.Limit)
	if err != nil {
		return nil, 0, err
	}

	serials := make([]*models.Serial, len(swtts))
	for i, swt := range swtts {
		serials[i] = &swt.Serial
	}

	total := 0
	if len(swtts) > 0 {
		total = swtts[0].TotalCount
	}

	return serials, total, nil
}

// GetByProductID 根據產品ID取得所有序號
func (r *SerialRepository) GetByProductID(ctx context.Context, productID string) ([]*models.Serial, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"serial_number",
			"full_serial_number",
			"product_id",
			"created_at",
			"updated_at",
		),
		sm.From("serials"),
		sm.Where(
			psql.And(
				psql.Quote("serials", "product_id").EQ(psql.Arg(productID)),
				psql.Quote("serials", "deleted_at").IsNull(),
			),
		),
		sm.OrderBy(psql.Quote("serials", "created_at")).Desc(),
	)
	return dbutil.GetAll[models.Serial](ctx, r.db, builder)
}

// GetSerialsWithProduct 取得序號及其產品資訊
func (r *SerialRepository) GetSerialsWithProduct(ctx context.Context, req *models.SerialSearchRequest, page *entity.Pagination) ([]*models.SerialDetailResponse, int, error) {
	conditions := []bob.Expression{}

	if req.ID.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "id").EQ(psql.Arg(req.ID.String)),
		)
	}
	if req.SerialNumber.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "serial_number").ILike(psql.Arg("%"+req.SerialNumber.String+"%")),
		)
	}
	if req.FullSerialNumber.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "full_serial_number").ILike(psql.Arg("%"+req.FullSerialNumber.String+"%")),
		)
	}
	if req.ProductID.Valid {
		conditions = append(conditions,
			psql.Quote("serials", "product_id").EQ(psql.Arg(req.ProductID.String)),
		)
	}

	// 預設不搜尋已刪除的序號
	deleteCondition := psql.Quote("serials", "deleted_at").IsNull()
	if req.SearchDeleted.Valid && req.SearchDeleted.Bool {
		deleteCondition = psql.Quote("serials", "deleted_at").IsNotNull()
	}
	conditions = append(conditions, deleteCondition)

	builder := psql.Select(
		sm.Columns(
			"serials.id",
			"serials.serial_number",
			"serials.full_serial_number",
			"serials.product_id",
			"serials.created_at",
			"serials.updated_at",
			"products.id AS product_id",
			"products.model_number",
			"products.brand",
			"products.type",
			"products.size",
			"products.warranty_years",
			"products.description",
			"products.is_active",
			"COUNT(serials.id) OVER() AS total_count",
		),
		sm.From("serials"),
		sm.LeftJoin("products").On(psql.Quote("serials", "product_id").EQ(psql.Quote("products", "id"))),
		page.OrderByClause("serials"),
		page.OffsetClause(),
		page.LimitClause(),
	)

	if len(conditions) > 0 {
		builder.Apply(sm.Where(psql.And(conditions...)))
	}

	type SerialWithProductAndTotal struct {
		models.SerialDetailResponse
		TotalCount int `db:"total_count"`
	}

	swpts, _, err := dbutil.GetPage[SerialWithProductAndTotal](ctx, r.db, builder, page.Limit)
	if err != nil {
		return nil, 0, err
	}

	serials := make([]*models.SerialDetailResponse, len(swpts))
	for i, swpt := range swpts {
		serials[i] = &swpt.SerialDetailResponse
	}

	total := 0
	if len(swpts) > 0 {
		total = swpts[0].TotalCount
	}

	return serials, total, nil
}

// ListSerialsUsedByWarranty 列出已經被保固使用的序號
func (r *SerialRepository) ListSerialsUsedByWarranty(ctx context.Context, page *entity.Pagination) ([]*models.Serial, int, error) {
	// 構建查詢：找出有保固記錄的序號
	builder := psql.Select(
		sm.Columns(
			"s.id",
			"s.serial_number",
			"s.full_serial_number",
			"s.product_id",
			"s.created_at",
			"s.updated_at",
			"COUNT(*) OVER() AS total_count",
		),
		sm.From("serials").As("s"),
		sm.InnerJoin("warranty_registrations").As("w").On(
			psql.Quote("s", "serial_number").EQ(psql.Quote("w", "product_serial_number")),
		),
		sm.OrderBy(psql.Quote("w", "created_at")).Desc().Asc(),
	)

	// 添加分頁
	if page != nil {
		builder.Apply(sm.Limit(page.Limit))
		builder.Apply(sm.Offset(page.Offset))
	}

	// 執行查詢
	type SerialWithTotal struct {
		models.Serial
		TotalCount int `db:"total_count"`
	}

	serialsWithTotal, err := dbutil.GetAll[SerialWithTotal](ctx, r.db, builder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get serials used by warranty: %w", err)
	}

	// 提取序號和總數
	serials := make([]*models.Serial, len(serialsWithTotal))
	total := 0
	for i, swt := range serialsWithTotal {
		serials[i] = &swt.Serial
		if i == 0 {
			total = swt.TotalCount
		}
	}

	return serials, total, nil
}
