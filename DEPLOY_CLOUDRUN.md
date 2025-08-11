# GCP Cloud Run 部署指南

## 🚀 **快速部署**

### **前置需求**
1. 安裝 [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
2. 安裝 [Docker](https://docs.docker.com/get-docker/)
3. 擁有 GCP 專案並啟用計費

### **步驟 1：設定 GCP 專案**
```bash
# 登入 GCP
gcloud auth login

# 設定專案 ID (替換為您的專案 ID)
export PROJECT_ID="your-gcp-project-id"
gcloud config set project $PROJECT_ID

# 啟用必要的 API
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
```

### **步驟 2：修改部署腳本**
編輯 `deploy-cloudrun.sh` 中的配置：
```bash
PROJECT_ID="your-gcp-project-id"
SERVICE_NAME="breast-implant-warranty"
REGION="asia-east1"
```

### **步驟 3：執行部署**
```bash
# 讓腳本可執行
chmod +x deploy-cloudrun.sh

# 執行部署
./deploy-cloudrun.sh
```

## ⚙️ **環境變數設定**

部署後，在 GCP Console 中設定以下環境變數：

### **必要環境變數**
```bash
# 資料庫設定
DB_HOST=your-cloud-sql-ip
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=breast_implant_warranty
DB_SSLMODE=require

# JWT 設定
JWT_SECRET=your-super-secret-jwt-key

# 郵件設定
MAILGUN_DOMAIN=your-mailgun-domain
MAILGUN_API_KEY=your-mailgun-api-key
MAILGUN_FROM_EMAIL=noreply@your-domain.com

# 應用程式設定
APP_URL=https://your-cloudrun-service-url
PORT=8080
GIN_MODE=release
```

## 🗄️ **資料庫設定**

### **選項 1：Cloud SQL (推薦)**
```bash
# 創建 Cloud SQL 實例
gcloud sql instances create breast-implant-db \
    --database-version=POSTGRES_15 \
    --tier=db-f1-micro \
    --region=asia-east1

# 創建資料庫
gcloud sql databases create breast_implant_warranty \
    --instance=breast-implant-db

# 創建用戶
gcloud sql users create warranty_user \
    --instance=breast-implant-db \
    --password=your-secure-password
```

### **選項 2：外部資料庫**
確保資料庫允許 Cloud Run 的 IP 範圍連接。

## 📊 **映像優化特點**

✅ **超輕量**: 僅 15.7MB  
✅ **多階段構建**: 分離構建和執行環境  
✅ **靜態鏈接**: 無外部依賴  
✅ **Scratch 基礎**: 最小攻擊面  
✅ **包含 SSL 證書**: 支援 HTTPS 連接  
✅ **時區支援**: 正確的時間處理  

## 🔧 **常用命令**

### **查看服務狀態**
```bash
gcloud run services describe breast-implant-warranty \
    --region=asia-east1
```

### **查看日誌**
```bash
gcloud run services logs read breast-implant-warranty \
    --region=asia-east1 --limit=50
```

### **更新服務**
```bash
# 重新構建並部署
./deploy-cloudrun.sh

# 或僅更新環境變數
gcloud run services update breast-implant-warranty \
    --region=asia-east1 \
    --set-env-vars="NEW_VAR=value"
```

### **設定自定義域名**
```bash
gcloud run domain-mappings create \
    --service=breast-implant-warranty \
    --domain=your-domain.com \
    --region=asia-east1
```

## 💰 **成本優化**

- **最小實例數**: 0 (冷啟動)
- **最大實例數**: 10
- **記憶體**: 512Mi
- **CPU**: 1
- **並發**: 80 個請求/實例

## 🔒 **安全性建議**

1. **使用 Cloud SQL** 而非公開資料庫
2. **設定 IAM 權限** 限制服務存取
3. **啟用 Cloud Armor** 防護 DDoS
4. **使用 Secret Manager** 管理敏感資料
5. **定期更新映像** 修補安全漏洞

## 🚨 **故障排除**

### **常見問題**

**1. 服務無法啟動**
```bash
# 檢查日誌
gcloud run services logs read breast-implant-warranty --region=asia-east1

# 檢查環境變數
gcloud run services describe breast-implant-warranty --region=asia-east1
```

**2. 資料庫連接失敗**
- 檢查 Cloud SQL 實例是否執行
- 確認網路設定允許 Cloud Run 連接
- 驗證資料庫憑證

**3. 記憶體不足**
```bash
# 增加記憶體限制
gcloud run services update breast-implant-warranty \
    --region=asia-east1 \
    --memory=1Gi
```

## 📈 **監控和警報**

在 GCP Console 中設定：
- **CPU 使用率警報**
- **記憶體使用率警報**
- **錯誤率警報**
- **回應時間警報**

## 🔄 **CI/CD 設定**

使用 `cloudbuild.yaml` 設定自動部署：
```bash
# 連接 GitHub 倉庫
gcloud builds triggers create github \
    --repo-name=your-repo \
    --repo-owner=your-username \
    --branch-pattern="^main$" \
    --build-config=cloudbuild.yaml
```
