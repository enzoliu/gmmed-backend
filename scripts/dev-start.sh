#!/bin/bash

echo "🚀 乳房植入物保固系統 - 開發環境啟動"
echo "====================================="

# 檢查 .env 文件是否存在
if [ ! -f ".env" ]; then
    echo "❌ .env 文件不存在"
    echo "請先複製 .env.example 為 .env 並設定相關設定"
    echo ""
    echo "快速設置："
    echo "  cp .env.example .env"
    echo "  ./scripts/generate-keys.sh"
    echo "  nano .env  # 編輯資料庫設定"
    exit 1
fi

echo "✅ 找到 .env 文件"

# 測試設定載入
echo "🔧 測試設定載入..."
go run test_config.go
if [ $? -ne 0 ]; then
    echo "❌ 設定載入失敗，請檢查 .env 文件"
    exit 1
fi

echo ""
echo "🔨 編譯檢查..."
go build ./...
if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗"
    exit 1
fi

echo "✅ 編譯成功"

# 檢查PostgreSQL連接（可選）
echo ""
echo "🗄️ 檢查資料庫連接..."

# 從 .env 讀取資料庫設定
source .env

# 嘗試連接資料庫
if command -v psql &> /dev/null; then
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" &> /dev/null
    if [ $? -eq 0 ]; then
        echo "✅ 資料庫連接成功"
    else
        echo "⚠️  資料庫連接失敗，請確保："
        echo "   1. PostgreSQL 正在執行"
        echo "   2. 資料庫 '$DB_NAME' 已建立"
        echo "   3. 使用者 '$DB_USER' 有權限訪問"
        echo ""
        echo "建立資料庫的SQL命令："
        echo "   CREATE DATABASE $DB_NAME;"
        echo "   CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
        echo "   GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
        echo ""
        read -p "是否繼續啟動伺服器？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
else
    echo "⚠️  psql 未安裝，跳過資料庫連接檢查"
fi

echo ""
echo "🌟 啟動開發伺服器..."
echo "伺服器將在 http://localhost:$SERVER_PORT 啟動"
echo "按 Ctrl+C 停止伺服器"
echo ""

# 啟動伺服器
go run cmd/server/main.go
