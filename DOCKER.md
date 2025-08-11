# Docker 開發環境

本專案提供簡潔的 Docker 開發環境，包含 PostgreSQL 資料庫。

## 🚀 快速開始

### 1. 啟動 Docker 環境

```bash
docker-compose up -d
```

### 2. 檢查服務狀態

```bash
docker-compose ps
```

### 3. 設置環境變數

```bash
# 複製環境變數範本
cp .env.docker .env

# 編輯環境變數 (可選)
nano .env
```

### 4. 執行資料庫遷移

```bash
# 方法一：使用 Go 遷移 (推薦)
go run test_migration.go

# 方法二：使用 Shell 腳本
cd migrations
./run_migrations.sh
cd ..

# 測試設定
go run test_config.go
```

### 5. 啟動應用

```bash
go run cmd/server/main.go
```

## 📋 服務資訊

| 服務 | 地址 | 用戶名 | 密碼 |
|------|------|--------|------|
| PostgreSQL | localhost:5432 | warranty_user | warranty_password_2024 |

## 🔧 常用命令

### Docker Compose 命令

```bash
# 啟動服務
docker-compose up -d

# 停止服務
docker-compose down

# 查看日誌
docker-compose logs -f

# 查看特定服務日誌
docker-compose logs -f postgres

# 重啟服務
docker-compose restart

# 重建並啟動
docker-compose up -d --build
```

### 資料庫操作

```bash
# 進入 PostgreSQL 容器
docker-compose exec postgres bash

# 直接連接資料庫
docker-compose exec postgres psql -U warranty_user -d breast_implant_warranty

# 備份資料庫
docker-compose exec postgres pg_dump -U warranty_user breast_implant_warranty > backup.sql

# 還原資料庫
docker-compose exec -T postgres psql -U warranty_user breast_implant_warranty < backup.sql
```



## 📁 檔案結構

```
.
├── docker-compose.yml          # Docker Compose 設定
├── .env.docker                 # Docker 環境變數範本
├── .env                        # 實際環境變數
└── migrations/                 # 資料庫遷移文件
    ├── 001_initial_schema.sql
    ├── 002_indexes_and_constraints.sql
    └── 003_initial_data.sql
```

## ⚙️ 設定說明

### 環境變數

主要設定在 `.env` 文件中：

```env
# 資料庫設定
DB_HOST=localhost
DB_PORT=5432
DB_NAME=breast_implant_warranty
DB_USER=warranty_user
DB_PASSWORD=warranty_password_2024

# 應用設定
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DEBUG=true
```

### 資料持久化

Docker 使用命名卷來持久化資料：

- `postgres_data`: PostgreSQL 資料

## 🛠️ 故障排除

### 常見問題

1. **阜號衝突**
   ```bash
   # 檢查阜號使用情況
   lsof -i :5432
   ```

2. **資料庫連接失敗**
   ```bash
   # 檢查容器狀態
   docker-compose ps
   
   # 查看資料庫日誌
   docker-compose logs postgres
   ```

3. **權限問題**
   ```bash
   # 重設資料卷權限
   docker-compose down -v
   docker-compose up -d
   ```

### 清理環境

```bash
# 停止並移除容器
docker-compose down

# 移除所有資料 (⚠️ 會刪除所有資料)
docker-compose down -v

# 清理未使用的 Docker 資源
docker system prune -f
```

## 🔒 安全注意事項

1. **生產環境**: 請修改所有預設密碼
2. **資料備份**: 定期備份重要資料
3. **SSL/TLS**: 生產環境中啟用 SSL 連接

## 📚 相關文檔

- [PostgreSQL 官方文檔](https://www.postgresql.org/docs/)
- [Docker Compose 文檔](https://docs.docker.com/compose/)
