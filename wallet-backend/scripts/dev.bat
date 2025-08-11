@echo off
chcp 65001 >nul

echo 🚀 启动钱包后端开发环境...

REM 检查Go是否安装
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go未安装，请先安装Go 1.21+
    pause
    exit /b 1
)

echo ✅ Go版本检查通过

REM 下载依赖
echo 📦 下载Go依赖...
go mod tidy

REM 检查配置文件
if not exist "config.yaml" (
    echo ⚠️  配置文件不存在，请手动创建config.yaml
    echo 💡 提示: 可以复制config.yaml.example并修改配置
)

REM 运行测试
echo 🧪 运行测试...
go test ./... -v

REM 构建应用
echo 🔨 构建应用...
go build -o main.exe ./cmd/main.go

REM 启动应用
echo 🚀 启动应用...
echo 📱 应用将在 http://localhost:8080 启动
echo 📊 健康检查: http://localhost:8080/health
echo 📖 API文档: http://localhost:8080/api/v1
echo.
echo 按 Ctrl+C 停止应用

main.exe 