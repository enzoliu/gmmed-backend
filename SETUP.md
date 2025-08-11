# 乳房植入物保固系統 - 開發環境設置指南

## 📋 前置需求

- Go 1.21 或更高版本
- PostgreSQL 13 或更高版本
- Git

## 🚀 快速開始

### 1. 克隆專案

```bash
git clone git@github.com:enzoliu/jymedical-backend.git
cd jymedical-backend
```

### 2. 安裝依賴

```bash
go mod download
```

### 3. 設置環境變數

複製環境變數範本：

```bash
cp .env.example .env
```

編輯 `.env` 文件並填入實際的設定值：

```bash
nano .env  # 或使用您喜歡的編輯器
```

### 4. 重要設定說明

#### 🔐 安全密鑰生成

**JWT Secret (建議至少32字符):**
```bash
openssl rand -base64 32
```

**加密密鑰 (必須是32字節):**
```bash
openssl rand -hex 32
```

#### 🗄️ 資料庫設置

1. 建立PostgreSQL資料庫：
```sql
CREATE DATABASE breast_implant_warranty;
CREATE USER warranty_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE breast_implant_warranty TO warranty_user;
```

2. 更新 `.env` 中的資料庫設定：
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=warranty_user
DB_PASSWORD=your_password
DB_NAME=breast_implant_warranty
DB_SSLMODE=disable
```

### 5. 執行資料庫遷移

```bash
# 如果您有遷移文件
psql -d breast_implant_warranty -f migrations/001_initial_schema.sql
```

### 6. 啟動開發伺服器

```bash
go run cmd/server/main.go
```

伺服器將在 `http://localhost:8080` 啟動。

## 🧪 測試

### 編譯檢查
```bash
go build ./...
```

### 執行測試
```bash
go test ./...
```

### API測試
```bash
go run test_api.go
```

## 📁 重要目錄說明

```
backend/
├── cmd/server/          # 主程式入口
├── internal/
│   ├── config/         # 設定管理
│   ├── database/       # 資料庫連接
│   ├── handlers/       # HTTP處理器
│   ├── middleware/     # 中間件
│   ├── models/         # 資料模型
│   ├── repositories/   # 資料庫操作層
│   ├── services/       # 業務邏輯層
│   └── utils/          # 工具函數
├── migrations/         # 資料庫遷移
├── scripts/           # 腳本文件
├── .env               # 環境變數 (不提交到Git)
├── .env.example       # 環境變數範本
└── go.mod             # Go模組定義
```

## 🔧 開發工具推薦

### VS Code 擴展
- Go (官方Go擴展)
- PostgreSQL (PostgreSQL語法支援)
- REST Client (API測試)

### 資料庫管理工具
- pgAdmin
- DBeaver
- TablePlus

## 🌍 環境變數說明

| 變數名稱 | 說明 | 預設值 | 必填 |
|---------|------|--------|------|
| `DB_HOST` | 資料庫主機 | localhost | ✅ |
| `DB_PORT` | 資料庫阜號 | 5432 | ✅ |
| `DB_USER` | 資料庫使用者 | postgres | ✅ |
| `DB_PASSWORD` | 資料庫密碼 | - | ✅ |
| `DB_NAME` | 資料庫名稱 | breast_implant_warranty | ✅ |
| `JWT_SECRET` | JWT簽名密鑰 | - | ✅ |
| `ENCRYPTION_KEY` | 資料加密密鑰 | - | ✅ |
| `SERVER_PORT` | 伺服器阜號 | 8080 | ❌ |
| `DEBUG` | 除錯模式 | false | ❌ |

## 🚨 安全注意事項

1. **絕對不要**將 `.env` 文件提交到版本控制
2. 在生產環境中使用強密碼和隨機密鑰
3. 定期更換JWT密鑰和加密密鑰
4. 確保資料庫連接使用SSL（生產環境）

## 📞 支援

如果遇到問題，請檢查：

1. Go版本是否正確
2. PostgreSQL是否正在執行
3. 環境變數是否正確設置
4. 資料庫連接是否正常

## 🔄 常見問題

**Q: 啟動時出現資料庫連接錯誤？**
A: 檢查PostgreSQL是否執行，以及 `.env` 中的資料庫設定是否正確。

**Q: JWT令牌驗證失敗？**
A: 確保 `JWT_SECRET` 已正確設置且足夠複雜。

**Q: 加密/解密失敗？**
A: 確保 `ENCRYPTION_KEY` 是32字節的十六進制字符串。
