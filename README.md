# 豐胸隆乳矽膠保固查詢系統 (繁體中文版)

## 說明

這是一個豐胸隆乳矽膠保固查詢系統 (繁體中文版)，預計實現以下功能：
1. 系統使用者登入
2. 醫院醫美中心管理
3. 系統使用者人員管理(管理人員/助理)
4. 產品資料管理(手動新增/EXCEL匯入)
5. 產品使用及保固紀錄
6. 權限設定

### 患者使用：
必填欄位: 姓名、身分證字號、出生年月日、手機、Email、診所、醫師、手術日、產品型號、序號(SN)
欄位驗證: 格式檢查、SN 不可重複
送出後回饋: 成功頁 + 自動確認信

信件服務預計使用mailgun

### 自動 Email
收件人: 患者 Email + 公司指定 Email
主旨範本: 「{患者姓氏} 您的植入物保固已完成登錄」
信件內容: 動態帶入：患者姓氏、診所、醫師、手術日、產品型號、序號

### 後台（公司使用）
登入權限: 至少分「可編輯」與「唯讀」
資料列表: 搜尋 / 篩選 (姓名、身分證、診所、SN…)
詳細頁: 顯示所有欄位與建立 / 修改時間
資料編輯: 手動更正欄位，需留異動紀錄
重寄確認信: 一鍵重新發送
匯出報表: CSV / Excel


### 系統必備特性
1. 身分證字號等敏感欄位加密儲存
2. 每日自動備份（保留 30 天）
3. 操作異動記錄（Audit Log）

## 技術架構

### 後端技術棧
- **語言**: Go 1.21+
- **Web 框架**: Echo v4 (高效能、豐富中間件)
- **資料庫**: PostgreSQL 15+ (支援加密、JSONB、全文搜尋)
- **ORM**: Bob (github.com/stephenafamo/bob/dialect/psql) - 型別安全的查詢建構器
- **認證**: JWT Token + bcrypt 密碼雜湊
- **信件服務**: Mailgun API
- **檔案處理**: excelize (Excel 匯入/匯出)
- **加密**: AES-256-GCM (敏感資料加密)
- **日誌**: logrus + 資料庫 audit log
- **設定管理**: Viper
- **API 文件**: Echo Swagger 中間件

### 系統架構設計

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端 (Web)    │    │   患者 QR 掃描   │    │   後台管理      │
│   React/Vue     │    │   手機瀏覽器     │    │   Admin Panel   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   API Gateway   │
                    │  (Echo Router)  │
                    └─────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Service  │    │ Warranty Service│    │  Admin Service  │
│   JWT/Session   │    │ QR/Registration │    │  CRUD/Reports   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │   主資料庫      │
                    │   6 個主要表    │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   File System   │
                    │   Excel Files   │
                    │   Backups       │
                    └─────────────────┘
```

### 資料安全設計
1. **敏感資料加密**: 身分證字號、手機號碼使用 AES-256-GCM 加密
2. **密碼安全**: bcrypt + salt 雜湊
3. **API 安全**: JWT Token + CORS + Rate Limiting
4. **資料庫安全**: 連線加密 + 參數化查詢防 SQL Injection
5. **Audit Log**: 所有操作記錄到資料庫，包含時間戳、用戶、操作類型、變更內容

### 目錄結構
```
├── cmd/
│   └── server/          # 主程式入口
├── internal/
│   ├── config/          # 設定管理
│   ├── models/          # 資料模型
│   ├── handlers/        # HTTP 處理器
│   ├── services/        # 業務邏輯
│   ├── repositories/    # 資料存取層
│   ├── middleware/      # 中間件
│   └── utils/           # 工具函數
├── migrations/          # 資料庫遷移
├── docs/               # API 文件
├── uploads/            # Excel 檔案上傳
├── exports/            # Excel 匯出檔案
├── backups/            # 備份檔案
└── configs/            # 設定檔案
```

## 🚀 快速開始

### 前置需求
- Go 1.21+
- Docker & Docker Compose (推薦) 或 PostgreSQL 13+
- Git

### 方法一：使用 Docker (推薦)

```bash
# 克隆專案
git clone <repository-url>
cd breast-implant-warranty-system/backend

# 啟動 Docker 環境
docker-compose up -d

# 測試設定和遷移
go run test_config.go
go run test_migration.go

# 啟動應用
go run cmd/server/main.go
```

### 方法二：手動設置

```bash
# 安裝依賴
go mod download

# 設定環境變數
cp .env.example .env

# 生成安全密鑰
./scripts/generate-keys.sh

# 編輯設定文件（設置資料庫密碼等）
nano .env
```

### 資料庫設定 (手動設置時需要)

```sql
-- 建立資料庫和使用者
CREATE DATABASE breast_implant_warranty;
CREATE USER warranty_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE breast_implant_warranty TO warranty_user;
```

### 啟動開發伺服器

```bash
# 使用快速啟動腳本（推薦）
./scripts/dev-start.sh

# 或手動啟動
go run cmd/server/main.go
```

伺服器將在 `http://localhost:8080` 啟動。

### 🐳 Docker 環境

詳細的 Docker 使用說明請參考 [DOCKER.md](DOCKER.md)。

**服務地址：**
- 應用伺服器: http://localhost:8080
- PostgreSQL: localhost:5432

### 🔧 開發工具

```bash
# 測試設定載入
go run test_config.go

# 生成新的安全密鑰
./scripts/generate-keys.sh

# 檢查編譯
./scripts/test.sh
```

### 🌍 環境變數設定

系統支援 `.env` 文件進行本地開發設定。主要設定項目：

| 變數名稱 | 說明 | 必填 |
|---------|------|------|
| `DB_HOST` | 資料庫主機 | ✅ |
| `DB_PASSWORD` | 資料庫密碼 | ✅ |
| `JWT_SECRET` | JWT簽名密鑰 | ✅ |
| `ENCRYPTION_KEY` | 資料加密密鑰 | ✅ |
| `SERVER_PORT` | 伺服器阜號 | ❌ |
| `DEBUG` | 除錯模式 | ❌ |

詳細設定說明請參考 [SETUP.md](SETUP.md)。

### 🔒 安全注意事項

1. **絕對不要**將 `.env` 文件提交到版本控制
2. 使用 `./scripts/generate-keys.sh` 生成強密鑰
3. 定期更換 JWT 密鑰和加密密鑰
4. 確保資料庫連接使用 SSL（生產環境）

## 開發環境設定

### 前置需求
- Go 1.21+
- PostgreSQL 15+
- Git

### 快速開始
```bash
# 1. 完整專案設定（一鍵設定）
make setup

# 2. 手動設定步驟
make setup-env          # 建立環境檔案
make deps               # 安裝依賴
make db-setup           # 設定資料庫
```

### 環境變數設定
編輯 `.env` 檔案設定必要的環境變數：
```bash
# 資料庫設定
DB_PASSWORD=your_password

# JWT 密鑰（請更改為安全的密鑰）
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# 加密密鑰（32字元）
ENCRYPTION_KEY=your-32-character-encryption-key-here

# Mailgun 設定
MAILGUN_DOMAIN=your-domain.mailgun.org
MAILGUN_API_KEY=your-mailgun-api-key
```

## 執行

### 開發模式
```bash
# 一般執行
make run

# 熱重載開發模式（推薦）
make dev
```

### 常用開發命令
```bash
make help              # 查看所有可用命令
make test              # 執行測試
make test-cover        # 執行測試並生成覆蓋率報告
make lint              # 程式碼檢查
make fmt               # 格式化程式碼
make db-reset          # 重置資料庫（危險操作）
```

## 編譯

```bash
# 本地編譯
go build -o bin/server cmd/server/main.go

# 交叉編譯 (Linux)
GOOS=linux GOARCH=amd64 go build -o bin/server-linux cmd/server/main.go
```

## 測試

```bash
# 執行所有測試
go test ./...

# 執行測試並顯示覆蓋率
go test -cover ./...

# 生成覆蓋率報告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
