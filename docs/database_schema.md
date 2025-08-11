# 資料庫設計文件

## 概述
本系統使用 PostgreSQL 作為主要資料庫，設計重點包括：
- 敏感資料加密儲存
- 完整的 audit trail
- 支援多層級權限管理
- 高效的查詢索引設計

## 主要資料表

### 1. users (系統使用者)
管理後台登入使用者，包含管理人員和助理

| 欄位名 | 類型 | 說明 | 約束 |
|--------|------|------|------|
| id | UUID | 主鍵 | PRIMARY KEY |
| username | VARCHAR(50) | 使用者名稱 | UNIQUE, NOT NULL |
| email | VARCHAR(255) | 電子信件 | UNIQUE, NOT NULL |
| password_hash | VARCHAR(255) | 密碼雜湊 | NOT NULL |
| role | VARCHAR(20) | 角色 (admin/editor/readonly) | NOT NULL |
| is_active | BOOLEAN | 是否啟用 | DEFAULT true |
| last_login_at | TIMESTAMP | 最後登入時間 | |
| created_at | TIMESTAMP | 建立時間 | DEFAULT NOW() |
| updated_at | TIMESTAMP | 更新時間 | DEFAULT NOW() |

### 2. products (產品資料)
管理矽膠產品資訊

| 欄位名 | 類型 | 說明 | 約束 |
|--------|------|------|------|
| id | UUID | 主鍵 | PRIMARY KEY |
| model_number | VARCHAR(100) | 產品型號 | NOT NULL |
| brand | VARCHAR(100) | 品牌 | NOT NULL |
| type | VARCHAR(50) | 產品類型 | NOT NULL |
| size | VARCHAR(50) | 尺寸 | |
| warranty_years | INTEGER | 保固年限 | DEFAULT 10 |
| description | TEXT | 產品描述 | |
| is_active | BOOLEAN | 是否啟用 | DEFAULT true |
| created_at | TIMESTAMP | 建立時間 | DEFAULT NOW() |
| updated_at | TIMESTAMP | 更新時間 | DEFAULT NOW() |

### 3. qr_codes (QR Code 管理)
管理生成的 QR Code

| 欄位名 | 類型 | 說明 | 約束 |
|--------|------|------|------|
| id | UUID | 主鍵 | PRIMARY KEY |
| code | VARCHAR(255) | QR Code 內容 | UNIQUE, NOT NULL |
| product_id | UUID | 關聯產品 | FOREIGN KEY |
| serial_number | VARCHAR(100) | 產品序號 | UNIQUE, NOT NULL |
| generated_by | UUID | 生成者 | FOREIGN KEY |
| is_used | BOOLEAN | 是否已使用 | DEFAULT false |
| used_at | TIMESTAMP | 使用時間 | |
| created_at | TIMESTAMP | 建立時間 | DEFAULT NOW() |

### 4. warranty_registrations (保固登記)
患者保固登記主表，包含診所和醫師資訊

| 欄位名 | 類型 | 說明 | 約束 |
|--------|------|------|------|
| id | UUID | 主鍵 | PRIMARY KEY |
| qr_code_id | UUID | QR Code | FOREIGN KEY |
| patient_name | VARCHAR(100) | 患者姓名 | NOT NULL |
| patient_id_encrypted | TEXT | 加密身分證字號 | NOT NULL |
| patient_birth_date | DATE | 出生年月日 | NOT NULL |
| patient_phone_encrypted | TEXT | 加密手機號碼 | NOT NULL |
| patient_email | VARCHAR(255) | 患者信箱 | NOT NULL |
| hospital_name | VARCHAR(255) | 診所名稱 | NOT NULL |
| doctor_name | VARCHAR(100) | 醫師姓名 | NOT NULL |
| surgery_date | DATE | 手術日期 | NOT NULL |
| product_id | UUID | 產品 | FOREIGN KEY |
| serial_number | VARCHAR(100) | 序號 | NOT NULL |
| warranty_start_date | DATE | 保固開始日 | NOT NULL |
| warranty_end_date | DATE | 保固結束日 | NOT NULL |
| confirmation_email_sent | BOOLEAN | 確認信已發送 | DEFAULT false |
| email_sent_at | TIMESTAMP | 發送時間 | |
| status | VARCHAR(20) | 狀態 (active/expired/cancelled) | DEFAULT 'active' |
| created_at | TIMESTAMP | 建立時間 | DEFAULT NOW() |
| updated_at | TIMESTAMP | 更新時間 | DEFAULT NOW() |

### 5. audit_logs (操作記錄)
系統操作 audit trail

| 欄位名 | 類型 | 說明 | 約束 |
|--------|------|------|------|
| id | UUID | 主鍵 | PRIMARY KEY |
| user_id | UUID | 操作者 | FOREIGN KEY |
| action | VARCHAR(50) | 操作類型 | NOT NULL |
| table_name | VARCHAR(50) | 影響的表 | NOT NULL |
| record_id | UUID | 記錄 ID | |
| old_values | JSONB | 變更前值 | |
| new_values | JSONB | 變更後值 | |
| ip_address | INET | IP 位址 | |
| user_agent | TEXT | 瀏覽器資訊 | |
| created_at | TIMESTAMP | 操作時間 | DEFAULT NOW() |

### 6. email_logs (信件記錄)
信件發送記錄

| 欄位名 | 類型 | 說明 | 約束 |
|--------|------|------|------|
| id | UUID | 主鍵 | PRIMARY KEY |
| warranty_registration_id | UUID | 保固登記 | FOREIGN KEY |
| recipient_email | VARCHAR(255) | 收件者 | NOT NULL |
| subject | VARCHAR(255) | 主旨 | NOT NULL |
| body | TEXT | 內容 | NOT NULL |
| status | VARCHAR(20) | 狀態 (sent/failed/pending) | NOT NULL |
| mailgun_id | VARCHAR(255) | Mailgun ID | |
| error_message | TEXT | 錯誤訊息 | |
| sent_at | TIMESTAMP | 發送時間 | |
| created_at | TIMESTAMP | 建立時間 | DEFAULT NOW() |

## 索引設計

### 主要查詢索引
- `idx_warranty_registrations_patient_name` - 患者姓名查詢
- `idx_warranty_registrations_serial_number` - 序號查詢
- `idx_warranty_registrations_hospital_name` - 診所名稱查詢
- `idx_warranty_registrations_doctor_name` - 醫師姓名查詢
- `idx_warranty_registrations_surgery_date` - 手術日期範圍查詢
- `idx_qr_codes_serial_number` - QR Code 序號查詢
- `idx_audit_logs_created_at` - Audit log 時間查詢
- `idx_audit_logs_user_id` - 使用者操作記錄查詢

### 複合索引
- `idx_warranty_registrations_hospital_doctor` - 診所+醫師查詢
- `idx_warranty_registrations_status_warranty_end` - 狀態+保固到期查詢

## 資料約束

### 外鍵約束
- 所有關聯表都設定適當的外鍵約束
- 使用 CASCADE 或 RESTRICT 根據業務邏輯決定

### 檢查約束
- 電子信件格式驗證
- 手機號碼格式驗證
- 日期邏輯驗證 (手術日期 <= 保固開始日期)

### 唯一約束
- 序號全域唯一
- QR Code 唯一
- 使用者名稱和信箱唯一
