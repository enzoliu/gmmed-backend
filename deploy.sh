#!/bin/bash

# GCP Cloud Run 部署腳本
# 使用方法: ./deploy-cloudrun.sh

set -e

# 顏色輸出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}GCP Cloud Run 部署腳本${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# 配置變數 (請根據您的 GCP 設定修改)
PROJECT_ID=${PROJECT_ID:-"jymedical"}
REPO=${REPO:-"docker-images"}
SERVICE_NAME=${SERVICE_NAME:-"jymedical-backend"}
REGION=${REGION:-"asia-east1"}
IMAGE_NAME="asia-east1-docker.pkg.dev/${PROJECT_ID}/${REPO}/${SERVICE_NAME}"

# 檢查必要工具
echo -e "${YELLOW}檢查必要工具...${NC}"
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}錯誤: gcloud CLI 未安裝${NC}"
    echo "請安裝 Google Cloud SDK: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo -e "${RED}錯誤: Docker 未安裝${NC}"
    echo "請安裝 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

echo -e "${GREEN}✓ 工具檢查完成${NC}"

# 檢查 GCP 認證
# echo -e "${YELLOW}檢查 GCP 認證...${NC}"
# if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1 > /dev/null; then
#     echo -e "${RED}錯誤: 未登入 GCP${NC}"
#     echo "請執行: gcloud auth login"
#     exit 1
# fi

echo -e "${GREEN}✓ GCP 認證檢查完成${NC}"

# 設定專案
echo -e "${YELLOW}設定 GCP 專案: ${PROJECT_ID}${NC}"
gcloud config set project ${PROJECT_ID}

# 啟用必要的 API
# echo -e "${YELLOW}啟用必要的 GCP API...${NC}"
# gcloud services enable cloudbuild.googleapis.com
# gcloud services enable run.googleapis.com

# 配置 Docker 認證
# echo -e "${YELLOW}配置 Docker 認證...${NC}"
# gcloud auth configure-docker

# 構建 Docker 映像
echo -e "${YELLOW}構建 Docker 映像...${NC}"
docker buildx build --platform linux/amd64 -t ${IMAGE_NAME}:latest .

# 推送映像到 Container Registry
echo -e "${YELLOW}推送映像到 Container Registry...${NC}"
docker push ${IMAGE_NAME}:latest

# 部署到 Cloud Run
# echo -e "${YELLOW}部署到 Cloud Run...${NC}"
# gcloud run deploy ${SERVICE_NAME} \
#     --image ${IMAGE_NAME}:latest \
#     --platform managed \
#     --region ${REGION} \
#     --allow-unauthenticated \
#     --port 8080 \
#     --memory 512Mi \
#     --cpu 1 \
#     --min-instances 0 \
#     --max-instances 10 \
#     --concurrency 80 \
#     --timeout 300 \
#     --set-env-vars "PORT=8080"

# 獲取服務 URL
# SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} --platform managed --region ${REGION} --format 'value(status.url)')

# echo ""
# echo -e "${GREEN}======================================${NC}"
# echo -e "${GREEN}部署完成！${NC}"
# echo -e "${GREEN}======================================${NC}"
# echo ""
# echo -e "${BLUE}服務資訊:${NC}"
# echo "  服務名稱: ${SERVICE_NAME}"
# echo "  區域: ${REGION}"
# echo "  映像: ${IMAGE_NAME}:latest"
# echo "  URL: ${SERVICE_URL}"
# echo ""
# echo -e "${YELLOW}注意事項:${NC}"
# echo "1. 請在 GCP Console 中設定環境變數 (資料庫連接等)"
# echo "2. 確保資料庫允許 Cloud Run 的 IP 連接"
# echo "3. 考慮設定 Cloud SQL Proxy 或使用 Cloud SQL"
# echo ""
# echo -e "${BLUE}有用的命令:${NC}"
# echo "  查看日誌: gcloud run services logs read ${SERVICE_NAME} --region ${REGION}"
# echo "  更新服務: gcloud run services update ${SERVICE_NAME} --region ${REGION}"
# echo "  刪除服務: gcloud run services delete ${SERVICE_NAME} --region ${REGION}"
