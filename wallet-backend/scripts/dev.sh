#!/bin/bash

# 开发环境启动脚本

set -e

echo "🚀 启动钱包后端开发环境..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go 1.21+"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go版本过低，需要1.21+，当前版本: $GO_VERSION"
    exit 1
fi

echo "✅ Go版本检查通过: $GO_VERSION"

# 检查MySQL是否运行
if ! command -v mysql &> /dev/null; then
    echo "⚠️  MySQL客户端未安装，跳过连接检查"
else
    if ! mysql -u root -p -e "SELECT 1;" &> /dev/null; then
        echo "⚠️  无法连接到MySQL，请确保MySQL服务正在运行"
        echo "💡 提示: 可以使用Docker启动MySQL"
        echo "   docker run -d --name mysql -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=wallet_db -p 3306:3306 mysql:8.0"
    else
        echo "✅ MySQL连接正常"
    fi
fi

# 下载依赖
echo "📦 下载Go依赖..."
go mod tidy

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "⚠️  配置文件不存在，创建默认配置..."
    cp config.yaml.example config.yaml 2>/dev/null || {
        echo "❌ 无法创建配置文件，请手动创建config.yaml"
        exit 1
    }
    echo "✅ 配置文件已创建，请根据需要修改配置"
fi

# 运行测试
echo "🧪 运行测试..."
go test ./... -v

# 构建应用
echo "🔨 构建应用..."
go build -o main ./cmd/main.go

# 启动应用
echo "🚀 启动应用..."
echo "📱 应用将在 http://localhost:8080 启动"
echo "📊 健康检查: http://localhost:8080/health"
echo "📖 API文档: http://localhost:8080/api/v1"
echo ""
echo "按 Ctrl+C 停止应用"

./main 