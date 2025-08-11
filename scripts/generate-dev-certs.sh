#!/bin/bash

# 生成開發用的自簽名 SSL 證書

set -e

CERT_DIR="./certs"
DOMAIN="localhost"

echo "🔐 生成開發用 SSL 證書..."

# 建立證書目錄
mkdir -p "$CERT_DIR"

# 生成私鑰
echo "📝 生成私鑰..."
openssl genrsa -out "$CERT_DIR/server.key" 2048

# 生成證書簽名請求 (CSR)
echo "📝 生成證書簽名請求..."
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" -subj "/C=TW/ST=Taiwan/L=Taipei/O=Development/OU=IT/CN=$DOMAIN"

# 生成自簽名證書
echo "📝 生成自簽名證書..."
openssl x509 -req -days 365 -in "$CERT_DIR/server.csr" -signkey "$CERT_DIR/server.key" -out "$CERT_DIR/server.crt" -extensions v3_req -extfile <(
cat <<EOF
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = 127.0.0.1
IP.1 = 127.0.0.1
IP.2 = ::1
EOF
)

# 清理臨時文件
rm "$CERT_DIR/server.csr"

echo "✅ SSL 證書生成完成！"
echo ""
echo "📁 證書文件位置:"
echo "  私鑰: $CERT_DIR/server.key"
echo "  證書: $CERT_DIR/server.crt"
echo ""
echo "🔧 使用方法:"
echo "  1. 在 .env 文件中設置:"
echo "     HTTPS_ENABLED=true"
echo "     CERT_FILE=./certs/server.crt"
echo "     KEY_FILE=./certs/server.key"
echo ""
echo "  2. 重啟伺服器"
echo ""
echo "⚠️  注意:"
echo "  - 這是自簽名證書，瀏覽器會顯示安全警告"
echo "  - 僅用於開發環境，不要在生產環境使用"
echo "  - 可以在瀏覽器中添加例外來忽略警告"
