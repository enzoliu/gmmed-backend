package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"breast-implant-warranty-system/internal/entity"
	"breast-implant-warranty-system/internal/models"
	"breast-implant-warranty-system/internal/utils"
	"breast-implant-warranty-system/pkg/dbutil"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
)

// WarrantyRepository 保固資料庫操作
type WarrantyRepository struct {
	db            dbutil.PgxClientItf
	encryptionKey string
}

// NewWarrantyRepository 建立新的保固倉庫
func NewWarrantyRepository(db dbutil.PgxClientItf, encryptionKey string) *WarrantyRepository {
	return &WarrantyRepository{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

// Create 建立保固登記
func (r *WarrantyRepository) Create(ctx context.Context, warranty *models.WarrantyRegistration) (string, error) {
	builder := psql.Insert(
		im.Into("warranty_registrations",
			"patient_name",
			"patient_id_encrypted",
			"patient_birth_date",
			"patient_phone_encrypted",
			"patient_email",
			"hospital_name",
			"doctor_name",
			"surgery_date",
			"product_id",
			"product_serial_number",
			"serial_number_2",
			"warranty_start_date",
			"warranty_end_date",
			"confirmation_email_sent",
			"status",
		),
		im.Values(psql.Arg(
			warranty.PatientName,
			warranty.PatientIDEncrypted,
			warranty.PatientBirthDate,
			warranty.PatientPhoneEncrypted,
			warranty.PatientEmail,
			warranty.HospitalName,
			warranty.DoctorName,
			warranty.SurgeryDate,
			warranty.ProductID,
			warranty.ProductSerialNumber,
			warranty.ProductSerialNumber2,
			warranty.WarrantyStartDate,
			warranty.WarrantyEndDate,
			warranty.ConfirmationEmailSent,
			warranty.Status,
		)),
		im.Returning("id"),
	)

	return dbutil.GetColumn[string](ctx, r.db, builder)
}

// GetByID 根據ID取得保固登記
func (r *WarrantyRepository) GetByID(ctx context.Context, id string) (*models.WarrantyRegistration, error) {
	builder := psql.Select(
		sm.Columns(
			"warranty_registrations.id",
			"warranty_registrations.patient_name",
			"warranty_registrations.patient_id_encrypted",
			"warranty_registrations.patient_birth_date",
			"warranty_registrations.patient_phone_encrypted",
			"warranty_registrations.patient_email",
			"warranty_registrations.hospital_name",
			"warranty_registrations.doctor_name",
			"warranty_registrations.surgery_date",
			"warranty_registrations.product_id",
			"warranty_registrations.product_serial_number",
			"warranty_registrations.serial_number_2",
			"warranty_registrations.warranty_start_date",
			"warranty_registrations.warranty_end_date",
			"warranty_registrations.confirmation_email_sent",
			"warranty_registrations.email_sent_at",
			"warranty_registrations.status",
			"warranty_registrations.created_at",
			"warranty_registrations.updated_at",
			"warranty_registrations.step",
			"products.model_number",
			"products.brand",
			"products.type",
			"products.size",
			"products.warranty_years",
			"products.description",
			"products.is_active",
		),
		sm.From("warranty_registrations"),
		sm.LeftJoin("products").On(
			psql.Raw("warranty_registrations.product_id = products.id"),
		),
		sm.Where(
			psql.And(
				psql.Quote("warranty_registrations", "id").EQ(psql.Arg(id)),
				psql.Quote("warranty_registrations", "deleted_at").IsNull(),
			),
		),
		sm.Limit(1),
	)
	warranty, err := dbutil.GetOne[models.WarrantyRegistration](ctx, r.db, builder)
	if err != nil {
		return nil, err
	}

	// 解密敏感資料
	if warranty.PatientIDEncrypted.Valid && warranty.PatientIDEncrypted.String != "" {
		if decryptedID, err := utils.DecryptPatientID(warranty.PatientIDEncrypted.String, r.encryptionKey); err == nil {
			warranty.PatientID = null.StringFrom(decryptedID)
		}
	}

	if warranty.PatientPhoneEncrypted.Valid && warranty.PatientPhoneEncrypted.String != "" {
		if decryptedPhone, err := utils.DecryptPatientPhone(warranty.PatientPhoneEncrypted.String, r.encryptionKey); err == nil {
			warranty.PatientPhone = null.StringFrom(decryptedPhone)
		}
	}

	return warranty, nil
}

// GetByProductSerialNumber 根據產品序號取得保固登記
func (r *WarrantyRepository) GetByProductSerialNumber(ctx context.Context, serialNumber string) (*models.WarrantyRegistration, error) {
	builder := psql.Select(
		sm.Columns(
			"id",
			"patient_name",
			"patient_id_encrypted",
			"patient_birth_date",
			"patient_phone_encrypted",
			"patient_email",
			"hospital_name",
			"doctor_name",
			"surgery_date",
			"product_id",
			"product_serial_number",
			"serial_number_2",
			"warranty_start_date",
			"warranty_end_date",
			"confirmation_email_sent",
			"email_sent_at",
			"status",
			"created_at",
			"updated_at",
			"step",
		),
		sm.From("warranty_registrations"),
		sm.Where(
			psql.And(
				psql.Or(
					psql.Quote("warranty_registrations", "product_serial_number").EQ(psql.Arg(serialNumber)),
					psql.Quote("warranty_registrations", "serial_number_2").EQ(psql.Arg(serialNumber)),
				),
				psql.Quote("warranty_registrations", "deleted_at").IsNull(),
			),
		),
		sm.Limit(1),
	)
	return dbutil.GetOne[models.WarrantyRegistration](ctx, r.db, builder)
}

// Update 更新保固登記
func (r *WarrantyRepository) Update(ctx context.Context, warranty *models.WarrantyRegistration) error {
	builder := psql.Update(
		um.Table("warranty_registrations"),
		um.SetCol("patient_name").ToArg(warranty.PatientName),
		um.SetCol("patient_id_encrypted").ToArg(warranty.PatientIDEncrypted),
		um.SetCol("patient_birth_date").ToArg(warranty.PatientBirthDate),
		um.SetCol("patient_phone_encrypted").ToArg(warranty.PatientPhoneEncrypted),
		um.SetCol("patient_email").ToArg(warranty.PatientEmail),
		um.SetCol("hospital_name").ToArg(warranty.HospitalName),
		um.SetCol("doctor_name").ToArg(warranty.DoctorName),
		um.SetCol("surgery_date").ToArg(warranty.SurgeryDate),
		um.SetCol("product_id").ToArg(warranty.ProductID),
		um.SetCol("warranty_start_date").ToArg(warranty.WarrantyStartDate),
		um.SetCol("warranty_end_date").ToArg(warranty.WarrantyEndDate),
		um.SetCol("status").ToArg(warranty.Status),
		um.SetCol("updated_at").To("NOW()"),
		um.SetCol("product_serial_number").ToArg(warranty.ProductSerialNumber),
		um.SetCol("serial_number_2").ToArg(warranty.ProductSerialNumber2),
		um.SetCol("step").ToArg(warranty.Step),
		um.Where(psql.Quote("warranty_registrations", "id").EQ(psql.Arg(warranty.ID))),
	)

	return dbutil.ShouldExec(ctx, r.db, builder)
}

// Delete 軟性刪除保固登記
func (r *WarrantyRepository) Delete(ctx context.Context, id string) error {
	builder := psql.Update(
		um.Table("warranty_registrations"),
		um.SetCol("deleted_at").To("NOW()"),
		um.Where(psql.Quote("warranty_registrations", "id").EQ(psql.Arg(id))),
	)

	return dbutil.ShouldExec(ctx, r.db, builder)
}

// UpdateEmailSent 更新信件發送狀態
func (r *WarrantyRepository) UpdateEmailSent(ctx context.Context, id string, sent bool) error {
	builder := psql.Update(
		um.Table("warranty_registrations"),
		um.SetCol("confirmation_email_sent").ToArg(sent),
		um.Where(psql.Quote("warranty_registrations", "id").EQ(psql.Arg(id))),
	)
	if sent {
		builder.Apply(
			um.SetCol("email_sent_at").To("NOW()"),
		)
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

// Search 搜尋保固登記
func (r *WarrantyRepository) Search(ctx context.Context, req *models.WarrantySearchRequest, page *entity.Pagination) ([]*models.WarrantyRegistration, int, error) {
	conditions := []bob.Expression{}

	conditions = append(conditions,
		psql.Quote("warranty_registrations", "created_at").NE(psql.Quote("warranty_registrations", "updated_at")),
	)

	// 檢查是否可能是身分證字號搜尋
	isPossiblePatientID := req.GeneralSearch.Valid && len(req.GeneralSearch.String) == 10 &&
		req.GeneralSearch.String[0] >= 'A' && req.GeneralSearch.String[0] <= 'Z'

	// 檢查是否可能是手機號碼搜尋
	isPossibleCellphone := req.GeneralSearch.Valid && utils.ValidateTaiwanPhone(req.GeneralSearch.String)

	// 處理通用搜尋 - 使用 OR 條件搜尋多個欄位
	if req.GeneralSearch.Valid && !isPossiblePatientID && !isPossibleCellphone {
		searchPattern := "%" + req.GeneralSearch.String + "%"
		conditions = append(conditions,
			psql.Or(
				psql.Quote("warranty_registrations", "patient_name").ILike(psql.Arg(searchPattern)),
				psql.Quote("warranty_registrations", "product_serial_number").ILike(psql.Arg(searchPattern)),
				psql.Quote("warranty_registrations", "hospital_name").ILike(psql.Arg(searchPattern)),
				psql.Quote("warranty_registrations", "doctor_name").ILike(psql.Arg(searchPattern)),
			),
		)
	} else {
		// 如果沒有通用搜尋，則使用具體欄位搜尋（AND 條件）
		if req.PatientName.Valid {
			conditions = append(conditions,
				psql.Quote("warranty_registrations", "patient_name").ILike(psql.Arg("%"+req.PatientName.String+"%")),
			)
		}

		if req.SerialNumber.Valid {
			conditions = append(conditions,
				psql.Or(
					psql.Quote("warranty_registrations", "product_serial_number").ILike(psql.Arg("%"+req.SerialNumber.String+"%")),
					psql.Quote("warranty_registrations", "serial_number_2").ILike(psql.Arg("%"+req.SerialNumber.String+"%")),
				),
			)

		}
	}

	if req.HospitalName.Valid {
		conditions = append(conditions,
			psql.Quote("warranty_registrations", "hospital_name").ILike(psql.Arg("%"+req.HospitalName.String+"%")),
		)
	}

	if req.DoctorName.Valid {
		conditions = append(conditions,
			psql.Quote("warranty_registrations", "doctor_name").ILike(psql.Arg("%"+req.DoctorName.String+"%")),
		)
	}

	if req.Status.Valid {
		conditions = append(conditions,
			psql.Quote("warranty_registrations", "status").EQ(psql.Arg(req.Status.String)),
		)
	}

	if req.StartDate.Valid {
		parsedStartDate, err := utils.ParseTaiwanDateToUTC(req.StartDate.String)
		if err != nil {
			return nil, 0, err
		}
		conditions = append(conditions,
			psql.Quote("warranty_registrations", "surgery_date").GTE(psql.Arg(parsedStartDate)),
		)
	}

	if req.EndDate.Valid {
		parsedEndDate, err := utils.ParseTaiwanDateToUTC(req.EndDate.String + " 23:59:59")
		if err != nil {
			return nil, 0, err
		}
		conditions = append(conditions,
			psql.Quote("warranty_registrations", "surgery_date").LTE(psql.Arg(parsedEndDate)),
		)
	}

	// 預設不搜尋已刪除的保固記錄
	deleteCondition := psql.Quote("warranty_registrations", "deleted_at").IsNull()
	if req.SearchDeleted.Valid && req.SearchDeleted.Bool {
		deleteCondition = psql.Quote("warranty_registrations", "deleted_at").IsNotNull()
	}
	conditions = append(conditions, deleteCondition)

	// 如果Product ID 是null，表示是空白保固記錄，不應該被搜尋到
	conditions = append(conditions, psql.Quote("warranty_registrations", "product_id").IsNotNull())

	builder := psql.Select(
		sm.Columns(
			"warranty_registrations.id",
			"warranty_registrations.patient_name",
			"warranty_registrations.patient_id_encrypted",
			"warranty_registrations.patient_birth_date",
			"warranty_registrations.patient_phone_encrypted",
			"warranty_registrations.patient_email",
			"warranty_registrations.hospital_name",
			"warranty_registrations.doctor_name",
			"warranty_registrations.surgery_date",
			"warranty_registrations.product_id",
			"warranty_registrations.product_serial_number",
			"warranty_registrations.serial_number_2",
			"warranty_registrations.warranty_start_date",
			"warranty_registrations.warranty_end_date",
			"warranty_registrations.confirmation_email_sent",
			"warranty_registrations.email_sent_at",
			"warranty_registrations.status",
			"warranty_registrations.created_at",
			"warranty_registrations.updated_at",
			"warranty_registrations.step",
			"products.model_number",
			"products.brand",
			"products.type",
			"products.size",
			"products.warranty_years",
			"products.description",
			"products.is_active",
			"COUNT(warranty_registrations.id) OVER() AS total_count",
		),
		sm.From("warranty_registrations"),
		sm.LeftJoin("products").On(
			psql.Raw("warranty_registrations.product_id = products.id"),
		),
		sm.OrderBy(psql.Quote("warranty_registrations", "created_at")).Desc(),
		page.OffsetClause(),
	)

	if len(conditions) > 0 {
		builder.Apply(sm.Where(psql.And(conditions...)))
	}

	type WarrantyWithTotalCount struct {
		models.WarrantyRegistration
		TotalCount int `db:"total_count"`
	}

	wwtcs, _, err := dbutil.GetPage[WarrantyWithTotalCount](ctx, r.db, builder, page.Limit)
	if err != nil {
		return nil, 0, err
	}

	warranties := make([]*models.WarrantyRegistration, len(wwtcs))
	for i, wwtc := range wwtcs {
		warranties[i] = &wwtc.WarrantyRegistration
		// 解密敏感資料
		if warranties[i].PatientIDEncrypted.Valid && warranties[i].PatientIDEncrypted.String != "" {
			if decryptedID, err := utils.DecryptPatientID(warranties[i].PatientIDEncrypted.String, r.encryptionKey); err == nil {
				warranties[i].PatientID = null.StringFrom(decryptedID)
			}
		}

		if warranties[i].PatientPhoneEncrypted.Valid && warranties[i].PatientPhoneEncrypted.String != "" {
			if decryptedPhone, err := utils.DecryptPatientPhone(warranties[i].PatientPhoneEncrypted.String, r.encryptionKey); err == nil {
				warranties[i].PatientPhone = null.StringFrom(decryptedPhone)
			}
		}
	}

	// 如果是身分證字號搜尋，進行後處理過濾
	if isPossiblePatientID && r.encryptionKey != "" {
		filteredWarranties := r.filterByPatientID(warranties, req.GeneralSearch.String)
		return filteredWarranties, len(filteredWarranties), nil
	}

	// 如果是手機號碼搜尋，進行後處理過濾
	if isPossibleCellphone && r.encryptionKey != "" {
		filteredWarranties := r.filterByCellphone(warranties, req.GeneralSearch.String)
		return filteredWarranties, len(filteredWarranties), nil
	}

	total := 0
	if len(warranties) > 0 {
		total = wwtcs[0].TotalCount
	}

	return warranties, total, nil
}

// GetStatistics 取得保固統計資料
func (r *WarrantyRepository) GetStatistics(ctx context.Context) (*models.WarrantyStatistics, error) {
	builder := psql.Select(
		sm.Columns(
			"COUNT(*) as total_registrations",
			psql.Raw("COUNT(CASE WHEN status = 'active' THEN 1 END) as active_warranties"),
			psql.Raw("COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_warranties"),
			psql.Raw("COUNT(CASE WHEN status = 'active' AND warranty_end_date <= NOW() + INTERVAL '30 days' AND warranty_end_date < '9999-01-01'::timestamp THEN 1 END) as expiring_soon"),
		),
		sm.From("warranty_registrations"),
	)

	result, err := dbutil.GetOne[models.WarrantyStatistics](ctx, r.db, builder)
	if err != nil {
		return nil, err
	}

	builder = psql.Select(
		sm.Columns(
			"hospital_name",
			"COUNT(*) as total_registrations",
			psql.Raw("COUNT(CASE WHEN status = 'active' THEN 1 END) as active_registrations"),
			psql.Raw("COUNT(DISTINCT doctor_name) as total_doctors"),
		),
		sm.From("v_hospital_statistics"),
		sm.GroupBy(psql.Quote("v_hospital_statistics", "hospital_name")),
		sm.OrderBy(psql.Quote("v_hospital_statistics", "total_registrations")).Desc(),
		sm.Limit(10),
	)
	result.HospitalStats, err = dbutil.GetAll[models.HospitalStatistic](ctx, r.db, builder)
	if err != nil {
		return nil, err
	}

	builder = psql.Select(
		sm.Columns(
			"p.id",
			"p.model_number",
			"p.brand",
			"COUNT(wr.id) as total_usage",
			psql.Raw("COUNT(CASE WHEN wr.status = 'active' THEN 1 END) as active_usage"),
		),
		sm.From("products").As("p"),
		sm.LeftJoin("warranty_registrations").As("wr").On(psql.Raw("p.id = wr.product_id")),
		sm.Where(psql.Quote("p", "is_active").EQ(psql.Arg(true))),
		sm.GroupBy(psql.Quote("p", "id")),
		sm.OrderBy(psql.Quote("total_usage")).Desc(),
		sm.Limit(10),
	)
	result.ProductStats, err = dbutil.GetAll[models.ProductStatistic](ctx, r.db, builder)
	if err != nil {
		return nil, err
	}

	builder = psql.Select(
		sm.Columns(
			"EXTRACT(YEAR FROM created_at) as year",
			"EXTRACT(MONTH FROM created_at) as month",
			"COUNT(*) as count",
		),
		sm.From("warranty_registrations"),
		sm.Where(psql.Quote("warranty_registrations", "created_at").GTE(psql.Arg(time.Now().AddDate(0, -12, 0)))),
		sm.GroupBy(psql.Quote("warranty_registrations", "created_at")),
		sm.OrderBy(psql.Quote("warranty_registrations", "created_at")).Desc(),
	)
	result.MonthlyRegistrations, err = dbutil.GetAll[models.MonthlyRegistration](ctx, r.db, builder)
	if err != nil {
		return nil, err
	}

	return result, nil

	// 月度統計
	// monthlyQuery := `
	// 	SELECT
	// 		EXTRACT(YEAR FROM created_at) as year,
	// 		EXTRACT(MONTH FROM created_at) as month,
	// 		COUNT(*) as count
	// 	FROM warranty_registrations
	// 	WHERE created_at >= NOW() - INTERVAL '12 months'
	// 	GROUP BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at)
	// 	ORDER BY year DESC, month DESC
	// `
}

// filterByPatientID 根據身分證字號過濾保固記錄
func (r *WarrantyRepository) filterByPatientID(warranties []*models.WarrantyRegistration, searchID string) []*models.WarrantyRegistration {
	var filtered []*models.WarrantyRegistration

	for _, warranty := range warranties {
		if warranty.PatientIDEncrypted.Valid && warranty.PatientIDEncrypted.String != "" {
			// 嘗試解密身分證字號
			if decryptedID, err := utils.DecryptPatientID(warranty.PatientIDEncrypted.String, r.encryptionKey); err == nil {
				// 檢查是否匹配搜尋的身分證字號
				if decryptedID == searchID {
					filtered = append(filtered, warranty)
				}
			}
		}
	}

	return filtered
}

// filterByCellphone 根據手機號碼過濾保固記錄
func (r *WarrantyRepository) filterByCellphone(warranties []*models.WarrantyRegistration, searchPhone string) []*models.WarrantyRegistration {
	var filtered []*models.WarrantyRegistration

	for _, warranty := range warranties {
		if warranty.PatientPhoneEncrypted.Valid && warranty.PatientPhoneEncrypted.String != "" {
			// 嘗試解密手機號碼
			if decryptedPhone, err := utils.DecryptPatientPhone(warranty.PatientPhoneEncrypted.String, r.encryptionKey); err == nil {
				// 檢查是否匹配搜尋的手機號碼
				if decryptedPhone == searchPhone {
					filtered = append(filtered, warranty)
				}
			}
		}
	}

	return filtered
}

// UpdateExpiredWarranties 批次更新所有過期的保固狀態
func (r *WarrantyRepository) UpdateExpiredWarranties(ctx context.Context) (int, error) {
	builder := psql.Update(
		um.Table("warranty_registrations"),
		um.SetCol("status").ToArg("expired"),
		um.SetCol("updated_at").To("NOW()"),
		um.Where(
			psql.And(
				psql.Quote("warranty_registrations", "status").EQ(psql.Arg("active")),
				psql.Quote("warranty_registrations", "warranty_end_date").LT(psql.Arg(time.Now())),
			),
		),
	)
	result, err := dbutil.Exec(ctx, r.db, builder)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// BatchCreateEmpty 批次創建空白保固記錄
func (r *WarrantyRepository) BatchCreateEmpty(ctx context.Context, count int) ([]string, error) {
	var ids []string

	// 準備批次插入的值
	valueStrings := make([]string, 0, count)
	valueArgs := make([]interface{}, 0, count)

	for i := 0; i < count; i++ {
		id := uuid.New()
		ids = append(ids, id.String())
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, 'pending', NOW(), NOW())", i+1))
		valueArgs = append(valueArgs, id)
	}

	query := fmt.Sprintf(`
		INSERT INTO warranty_registrations (id, status, created_at, updated_at)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := r.db.Exec(ctx, query, valueArgs...)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
