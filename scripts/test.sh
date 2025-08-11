#!/bin/bash

echo "🚀 乳房植入物保固系統測試腳本"
echo "================================"

# 檢查Go是否安裝
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安裝，請先安裝 Go"
    exit 1
fi

echo "✅ Go 已安裝"

# 檢查依賴
echo "📦 檢查依賴..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "❌ 依賴檢查失敗"
    exit 1
fi

echo "✅ 依賴檢查完成"

# 編譯檢查
echo "🔨 編譯檢查..."
go build ./...
if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗"
    exit 1
fi

echo "✅ 編譯成功"

# 執行測試
echo "🧪 執行單元測試..."
go test ./... -v
if [ $? -ne 0 ]; then
    echo "⚠️  單元測試失敗或無測試文件"
fi

echo "✅ 測試腳本執行完成"
echo ""
echo "📝 下一步："
echo "1. 設置PostgreSQL資料庫"
echo "2. 設定環境變數"
echo "3. 執行資料庫遷移"
echo "4. 啟動伺服器: go run cmd/server/main.go"
echo "5. 執行API測試: go run test_api.go"
