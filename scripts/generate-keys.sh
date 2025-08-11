#!/bin/bash

echo "🔐 乳房植入物保固系統 - 安全密鑰生成器"
echo "========================================"

# 檢查 openssl 是否可用
if ! command -v openssl &> /dev/null; then
    echo "❌ OpenSSL 未安裝，請先安裝 OpenSSL"
    exit 1
fi

echo ""
echo "📝 生成的密鑰請複製到您的 .env 文件中："
echo ""

# 生成 JWT Secret
echo "🔑 JWT_SECRET (用於JWT令牌簽名):"
JWT_SECRET=$(openssl rand -base64 32)
echo "JWT_SECRET=$JWT_SECRET"
echo ""

# 生成加密密鑰
echo "🔒 ENCRYPTION_KEY (用於敏感資料加密):"
ENCRYPTION_KEY=$(openssl rand -hex 32)
echo "ENCRYPTION_KEY=$ENCRYPTION_KEY"
echo ""

# 生成隨機資料庫密碼建議
echo "🗄️ 建議的資料庫密碼:"
DB_PASSWORD=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-16)
echo "DB_PASSWORD=$DB_PASSWORD"
echo ""

echo "⚠️  重要提醒："
echo "1. 請將這些密鑰保存在安全的地方"
echo "2. 不要將 .env 文件提交到版本控制"
echo "3. 在生產環境中，請使用更強的密碼"
echo "4. 定期更換這些密鑰以提高安全性"
echo ""

# 詢問是否要自動更新 .env 文件
read -p "是否要自動更新 .env 文件？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -f ".env" ]; then
        # 備份現有的 .env 文件
        cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
        echo "✅ 已備份現有的 .env 文件"
        
        # 更新密鑰
        sed -i.tmp "s/^JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env
        sed -i.tmp "s/^ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" .env
        rm .env.tmp 2>/dev/null
        
        echo "✅ 已更新 .env 文件中的密鑰"
    else
        echo "❌ .env 文件不存在，請先複製 .env.example 為 .env"
    fi
fi

echo ""
echo "🎉 密鑰生成完成！"
