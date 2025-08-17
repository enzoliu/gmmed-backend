package repositories

import (
	"context"
	"fmt"
	"strings"

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

// ProductRepository 產品資料庫操作
type ProductRepository struct {
	db dbutil.PgxClientItf
}

// NewProductRepository 建立新的產品倉庫
func NewProductRepository(db dbutil.PgxClientItf) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create 建立產品
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) (string, error) {
	builder := psql.Insert(
		im.Into("products",
			"model_number",
			"brand",
			"type",
			"size",
			"warranty_years",
			"description",
			"is_active",
		),
		im.Values(psql.Arg(
			product.ModelNumber,
			product.Brand,
			product.Type,
			product.Size,
			product.WarrantyYears,
			product.Description,
			product.IsActive,
		)),
		im.Returning("id"),
	)

	return dbutil.GetColumn[string](ctx, r.db, builder)
}

// GetByID 根據ID取得產品
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"model_number",
			"brand",
			"type",
			"size",
			"warranty_years",
			"description",
			"is_active",
			"created_at",
			"updated_at",
		),
		sm.From("products"),
		sm.Where(
			psql.And(
				psql.Quote("products", "id").EQ(psql.Arg(id)),
				psql.Quote("products", "deleted_at").IsNull(),
			),
		),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.Product](ctx, r.db, builder)
}

// GetOneByCondition 根據條件取得單一產品
func (r *ProductRepository) GetOneByCondition(ctx context.Context, req *models.ProductSearchRequest) (*models.Product, error) {
	conditions := []bob.Expression{
		psql.Quote("products", "is_active").EQ(psql.Arg(true)),
	}
	if req.ID.Valid {
		conditions = append(conditions,
			psql.Quote("products", "id").EQ(psql.Arg(req.ID.String)),
		)
	}
	if req.Brand.Valid {
		conditions = append(conditions,
			psql.Quote("products", "brand").EQ(psql.Arg(req.Brand.String)),
		)
	}
	if req.Type.Valid {
		conditions = append(conditions,
			psql.Quote("products", "type").EQ(psql.Arg(req.Type.String)),
		)
	}
	if req.ModelNumber.Valid {
		conditions = append(conditions,
			psql.Quote("products", "model_number").EQ(psql.Arg(req.ModelNumber.String)),
		)
	}
	if req.Size.Valid {
		conditions = append(conditions,
			psql.Quote("products", "size").EQ(psql.Arg(req.Size.String)),
		)
	}

	// 預設不搜尋已刪除的產品
	deleteCondition := psql.Quote("products", "deleted_at").IsNull()
	if req.SearchDeleted.Valid && req.SearchDeleted.Bool {
		deleteCondition = psql.Quote("products", "deleted_at").IsNotNull()
	}
	conditions = append(conditions, deleteCondition)

	builder := psql.Select(
		sm.Columns(
			"id",
			"model_number",
			"brand",
			"type",
			"size",
			"warranty_years",
			"description",
			"is_active",
			"created_at",
			"updated_at",
		),
		sm.From("products"),
		sm.OrderBy(psql.Quote("products", "created_at")).Desc(),
		sm.Where(psql.And(conditions...)),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.Product](ctx, r.db, builder)
}

// GetByModelNumber 根據型號取得產品
func (r *ProductRepository) GetByModelNumber(ctx context.Context, modelNumber string) (*models.Product, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"model_number",
			"brand",
			"type",
			"size",
			"warranty_years",
			"description",
			"is_active",
			"created_at",
			"updated_at",
		),
		sm.From("products"),
		sm.Where(psql.And(
			psql.Quote("products", "model_number").EQ(psql.Arg(modelNumber)),
			psql.Quote("products", "is_active").EQ(psql.Arg(true)),
			psql.Quote("products", "deleted_at").IsNull(),
		)),
		sm.Limit(1),
	)

	return dbutil.GetOne[models.Product](ctx, r.db, builder)
}

// IsDuplicate 檢查是否重複
func (r *ProductRepository) IsDuplicate(ctx context.Context, product *models.Product) (bool, error) {
	conditions := []bob.Expression{
		psql.Quote("products", "model_number").EQ(psql.Arg(product.ModelNumber)),
		psql.Quote("products", "brand").EQ(psql.Arg(product.Brand)),
		psql.Quote("products", "type").EQ(psql.Arg(product.Type)),
		psql.Quote("products", "size").EQ(psql.Arg(product.Size)),
	}
	if product.ID != "" {
		conditions = append(conditions, psql.Quote("products", "id").NE(psql.Arg(product.ID)))
	}
	conditions = append(conditions,
		psql.Quote("products", "deleted_at").IsNull(),
	)

	builder := psql.Select(
		sm.Columns(
			"id",
		),
		sm.From("products"),
		sm.Where(psql.And(conditions...)),
		sm.Limit(1),
	)

	prod, err := dbutil.GetOne[models.Product](ctx, r.db, builder)

	if err != nil && err != pgx.ErrNoRows {
		return true, err
	}

	if prod != nil {
		return true, nil
	}

	return false, nil
}

// Update 更新產品
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	builder := psql.Update(
		um.Table("products"),
		um.SetCol("model_number").ToArg(product.ModelNumber),
		um.SetCol("brand").ToArg(product.Brand),
		um.SetCol("type").ToArg(product.Type),
		um.SetCol("size").ToArg(product.Size),
		um.SetCol("warranty_years").ToArg(product.WarrantyYears),
		um.SetCol("description").ToArg(product.Description),
		um.SetCol("is_active").ToArg(product.IsActive),
		um.SetCol("updated_at").To("NOW()"),
		um.Where(psql.Quote("products", "id").EQ(psql.Arg(product.ID))),
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

// Delete 軟性刪除產品
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	builder := psql.Update(
		um.Table("products"),
		um.SetCol("deleted_at").To("NOW()"),
		um.Where(psql.Quote("products", "id").EQ(psql.Arg(id))),
	)

	return dbutil.ShouldExec(ctx, r.db, builder)
}

// Search 搜尋產品
func (r *ProductRepository) Search(ctx context.Context, req *models.ProductSearchRequest, page *entity.Pagination) ([]*models.Product, int, error) {
	conditions := []bob.Expression{}
	if req.Brand.Valid {
		conditions = append(conditions,
			psql.Quote("products", "brand").ILike(psql.Arg("%"+req.Brand.String+"%")),
		)
	}
	if req.Type.Valid {
		conditions = append(conditions,
			psql.Quote("products", "type").ILike(psql.Arg("%"+req.Type.String+"%")),
		)
	}
	if req.Active.Valid {
		conditions = append(conditions,
			psql.Quote("products", "is_active").EQ(psql.Arg(req.Active.Bool)),
		)
	}
	if req.ModelNumber.Valid {
		conditions = append(conditions,
			psql.Quote("products", "model_number").ILike(psql.Arg("%"+req.ModelNumber.String+"%")),
		)
	}

	if req.WarrantyYears.Valid {
		conditions = append(conditions,
			psql.Quote("products", "warranty_years").EQ(psql.Arg(req.WarrantyYears.Int64)),
		)
	}
	if req.Size.Valid {
		conditions = append(conditions,
			psql.Quote("products", "size").ILike(psql.Arg("%"+req.Size.String+"%")),
		)
	}
	conditions = append(conditions,
		psql.Quote("products", "deleted_at").IsNull(),
	)

	builder := psql.Select(
		sm.Columns(
			"id",
			"model_number",
			"brand",
			"type",
			"size",
			"warranty_years",
			"description",
			"is_active",
			"created_at",
			"updated_at",
			"COUNT(id) OVER() AS total_count",
		),
		sm.From("products"),
		sm.OrderBy(psql.Quote("products", "created_at")).Desc(),
		page.OffsetClause(),
	)

	if len(conditions) > 0 {
		builder.Apply(sm.Where(psql.And(conditions...)))
	}

	type ProductWithTotalCount struct {
		models.Product
		TotalCount int `db:"total_count"`
	}
	pwtts, _, err := dbutil.GetPage[ProductWithTotalCount](ctx, r.db, builder, page.Limit)
	if err != nil {
		return nil, 0, err
	}

	products := make([]*models.Product, len(pwtts))
	for i, pwt := range pwtts {
		products[i] = &pwt.Product
	}

	total := 0
	if len(pwtts) > 0 {
		total = pwtts[0].TotalCount
	}

	return products, total, nil
}

// GetAllBrands 取得所有品牌
func (r *ProductRepository) GetAllBrands(ctx context.Context) ([]string, error) {
	builder := psql.Select(
		sm.Columns("products.brand"),
		sm.From("products"),
		sm.Where(
			psql.And(
				psql.Quote("products", "is_active").EQ(psql.Arg(true)),
				psql.Quote("products", "deleted_at").IsNull(),
			),
		),
		sm.GroupBy(psql.Quote("products", "brand")),
	)
	return dbutil.GetAllColumns[string](ctx, r.db, builder)
}

// GetAllTypes 取得所有類型
func (r *ProductRepository) GetAllTypes(ctx context.Context, req *models.ProductSearchRequest) ([]string, error) {
	constraints := []bob.Expression{}
	if req.Brand.Valid {
		constraints = append(constraints, psql.Quote("products", "brand").EQ(psql.Arg(req.Brand.String)))
	}
	constraints = append(constraints, psql.Quote("products", "is_active").EQ(psql.Arg(true)))
	constraints = append(constraints, psql.Quote("products", "deleted_at").IsNull())
	builder := psql.Select(
		sm.Columns("products.type"),
		sm.From("products"),
		sm.GroupBy(psql.Quote("products", "type")),
		sm.OrderBy(psql.Quote("products", "type")),
	)
	if len(constraints) > 0 {
		builder.Apply(sm.Where(psql.And(constraints...)))
	}
	return dbutil.GetAllColumns[string](ctx, r.db, builder)
}

// GetAllModelNumbers 取得所有型號
func (r *ProductRepository) GetAllModelNumbers(ctx context.Context, req *models.ProductSearchRequest) ([]string, error) {
	constraints := []bob.Expression{}
	if req.Brand.Valid {
		constraints = append(constraints, psql.Quote("products", "brand").EQ(psql.Arg(req.Brand.String)))
	}
	if req.Type.Valid {
		constraints = append(constraints, psql.Quote("products", "type").EQ(psql.Arg(req.Type.String)))
	}
	constraints = append(constraints, psql.Quote("products", "is_active").EQ(psql.Arg(true)))
	constraints = append(constraints, psql.Quote("products", "deleted_at").IsNull())
	builder := psql.Select(
		sm.Columns("products.model_number"),
		sm.From("products"),
		sm.GroupBy(psql.Quote("products", "model_number")),
		sm.OrderBy(psql.Quote("products", "model_number")),
	)
	if len(constraints) > 0 {
		builder.Apply(sm.Where(psql.And(constraints...)))
	}
	return dbutil.GetAllColumns[string](ctx, r.db, builder)
}

// GetAllSizes 取得所有尺寸
func (r *ProductRepository) GetAllSizes(ctx context.Context, req *models.ProductSearchRequest) ([]string, error) {
	constraints := []bob.Expression{}
	if req.Brand.Valid {
		constraints = append(constraints, psql.Quote("products", "brand").EQ(psql.Arg(req.Brand.String)))
	}
	if req.Type.Valid {
		constraints = append(constraints, psql.Quote("products", "type").EQ(psql.Arg(req.Type.String)))
	}
	if req.ModelNumber.Valid {
		constraints = append(constraints, psql.Quote("products", "model_number").EQ(psql.Arg(req.ModelNumber.String)))
	}
	constraints = append(constraints, psql.Quote("products", "is_active").EQ(psql.Arg(true)))
	constraints = append(constraints, psql.Quote("products", "deleted_at").IsNull())
	builder := psql.Select(
		sm.Columns("products.size"),
		sm.From("products"),
		sm.GroupBy(psql.Quote("products", "size")),
	)
	if len(constraints) > 0 {
		builder.Apply(sm.Where(psql.And(constraints...)))
	}
	return dbutil.GetAllColumns[string](ctx, r.db, builder)
}

// GetMetadataAll 取得所有產品元資料（品牌、型號、類型、尺寸）
func (r *ProductRepository) GetMetadataAll(ctx context.Context) (*models.ProductMetadataAllResponse, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"model_number",
			"brand",
			"type",
			"size",
			"warranty_years",
			"description",
			"is_active",
			"created_at",
			"updated_at",
		),
		sm.From("products"),
		sm.Where(
			psql.And(
				psql.Quote("products", "is_active").EQ(psql.Arg(true)),
				psql.Quote("products", "deleted_at").IsNull(),
			),
		),
	)
	res, err := dbutil.GetAll[models.Product](ctx, r.db, builder)
	if err != nil {
		return nil, err
	}
	data := make([]models.ProductMetadata, 0, len(res))
	for _, p := range res {
		data = append(data, models.ProductMetadata{
			Brand:       p.Brand,
			Type:        p.Type,
			ModelNumber: p.ModelNumber,
			Size:        p.Size.String,
		})
	}

	return &models.ProductMetadataAllResponse{
		Data: data,
	}, nil
}

// GetAllProducts 取得所有產品
func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]*models.Product, error) {
	builder := psql.Select(
		sm.Columns("id",
			"model_number",
			"brand",
			"type",
			"size",
			"warranty_years",
			"description",
			"is_active",
			"created_at",
			"updated_at",
		),
		sm.From("products"),
		sm.Where(psql.Quote("products", "deleted_at").IsNull()),
	)
	return dbutil.GetAll[models.Product](ctx, r.db, builder)
}

// CheckManyProductIDExists 檢查多個產品ID是否存在，只返回不存在的ids
func (r *ProductRepository) CheckManyProductIDExists(ctx context.Context, productIDs []string) ([]string, error) {
	if len(productIDs) == 0 {
		return []string{}, nil
	}

	const batchSize = 100 // 每批處理 100 個型號

	// 分批處理，使用真正的 NOT IN 查詢
	nonExistingProductIDs := make([]string, 0)

	for i := 0; i < len(productIDs); i += batchSize {
		end := i + batchSize
		if end > len(productIDs) {
			end = len(productIDs)
		}

		batch := productIDs[i:end]
		batchArgs := make([]any, 0, len(batch))
		for _, productID := range batch {
			batchArgs = append(batchArgs, productID)
		}

		if len(batchArgs) > 0 {
			// 使用 LEFT JOIN + IS NULL 查詢，這是最高效的方法
			// 資料庫優化器對 JOIN 的處理通常比 NOT IN 更優化

			// 構建 VALUES 子句
			valuesClause := "VALUES "
			placeholders := make([]string, len(batch))
			for j := range batch {
				placeholders[j] = fmt.Sprintf("($%d::uuid)", j+1)
			}
			valuesClause += strings.Join(placeholders, ", ")

			// 使用 LEFT JOIN + IS NULL 查詢
			rawSQL := fmt.Sprintf(`
				SELECT m.id 
				FROM (%s) AS m(id)
				LEFT JOIN products p ON m.id = p.id
				WHERE p.id IS NULL
			`, valuesClause)

			// 執行查詢
			rows, err := r.db.Query(ctx, rawSQL, batchArgs...)
			if err != nil {
				return nil, fmt.Errorf("failed to query non-existing model numbers batch %d-%d: %w", i, end-1, err)
			}
			defer rows.Close()

			// 讀取結果
			for rows.Next() {
				var productID string
				if err := rows.Scan(&productID); err != nil {
					return nil, fmt.Errorf("failed to scan model number: %w", err)
				}
				nonExistingProductIDs = append(nonExistingProductIDs, productID)
			}

			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("error iterating rows: %w", err)
			}
		}
	}

	return nonExistingProductIDs, nil
}
