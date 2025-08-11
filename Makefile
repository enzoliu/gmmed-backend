# 豐胸隆乳矽膠保固查詢系統 Makefile

.PHONY: help build run test clean deps migrate db-setup dev

# 預設目標
help: ## 顯示幫助資訊
	@echo "可用的命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# 開發相關
deps: ## 安裝依賴套件
	go mod download
	go mod tidy

build: ## 編譯應用程式
	go build -o bin/server cmd/server/main.go

run: ## 執行應用程式
	go run cmd/server/main.go

dev: ## 開發模式（使用 air 熱重載）
	@if ! command -v air > /dev/null; then \
		echo "安裝 air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	air

test: ## 執行完整測試套件
	@echo "🧪 執行完整測試套件..."
	@chmod +x tests/run_tests.sh
	@./tests/run_tests.sh

test-unit: ## 執行單元測試
	@echo "🧪 執行單元測試..."
	@go test -v ./tests/... -short

test-integration: ## 執行整合測試
	@echo "🧪 執行整合測試..."
	@go test -v ./tests/... -run Integration

test-cover: ## 執行測試並生成覆蓋率報告
	@echo "📊 生成測試覆蓋率報告..."
	@go test -v ./tests/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 覆蓋率報告已生成：coverage.html"

test-clean: ## 清理測試資料
	@echo "🧹 清理測試資料..."
	@rm -f coverage.out coverage.html
	@docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS breast_implant_warranty_test;" 2>/dev/null || true

clean: ## 清理編譯檔案
	rm -rf bin/
	rm -f coverage.out coverage.html

# 資料庫相關
db-setup: ## 設定資料庫（建立資料庫並執行遷移）
	@echo "設定資料庫..."
	./migrations/run_migrations.sh

migrate: ## 執行資料庫遷移
	@echo "執行資料庫遷移..."
	./migrations/run_migrations.sh

db-reset: ## 重置資料庫（危險操作）
	@echo "警告：這將刪除所有資料！"
	@read -p "確定要繼續嗎？(y/N): " confirm && [ "$$confirm" = "y" ]
	dropdb breast_implant_warranty || true
	createdb breast_implant_warranty
	./migrations/run_migrations.sh

# 程式碼品質
lint: ## 執行程式碼檢查
	@if ! command -v golangci-lint > /dev/null; then \
		echo "安裝 golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run

fmt: ## 格式化程式碼
	go fmt ./...
	goimports -w .

# 部署相關
docker-build: ## 建立 Docker 映像
	docker build -t breast-implant-warranty-system .

docker-run: ## 執行 Docker 容器
	docker-compose up -d

docker-stop: ## 停止 Docker 容器
	docker-compose down

# 工具
generate-qr: ## 生成測試用 QR Code
	@echo "生成測試 QR Code..."
	go run scripts/generate_test_qr.go

backup-db: ## 備份資料庫
	@echo "備份資料庫..."
	pg_dump breast_implant_warranty > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql

restore-db: ## 還原資料庫（需要指定備份檔案）
	@if [ -z "$(FILE)" ]; then \
		echo "請指定備份檔案：make restore-db FILE=backup_file.sql"; \
		exit 1; \
	fi
	@echo "還原資料庫從 $(FILE)..."
	dropdb breast_implant_warranty || true
	createdb breast_implant_warranty
	psql breast_implant_warranty < $(FILE)

# 環境設定
setup-env: ## 設定環境檔案
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "已建立 .env 檔案，請編輯設定值"; \
	else \
		echo ".env 檔案已存在"; \
	fi

# 完整設定
setup: setup-env deps db-setup ## 完整專案設定
	@echo "專案設定完成！"
	@echo "請編輯 .env 檔案設定必要的環境變數"
	@echo "然後執行 'make run' 啟動應用程式"

# 開發工具安裝
install-tools: ## 安裝開發工具
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# 檢查環境
check-env: ## 檢查開發環境
	@echo "檢查 Go 版本..."
	@go version
	@echo "檢查 PostgreSQL..."
	@psql --version || echo "PostgreSQL 未安裝"
	@echo "檢查必要工具..."
	@command -v air > /dev/null && echo "✓ air 已安裝" || echo "✗ air 未安裝"
	@command -v golangci-lint > /dev/null && echo "✓ golangci-lint 已安裝" || echo "✗ golangci-lint 未安裝"
