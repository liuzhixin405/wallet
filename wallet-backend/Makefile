# Makefile for Wallet Backend

# 变量定义
APP_NAME=wallet-backend
VERSION=1.0.0
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Go相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=main
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker相关变量
DOCKER_IMAGE=$(APP_NAME):$(VERSION)
DOCKER_LATEST=$(APP_NAME):latest

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Wallet Backend Makefile"
	@echo "======================="
	@echo ""
	@echo "可用的命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发相关命令
.PHONY: build
build: ## 构建应用
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/main.go

.PHONY: run
run: ## 运行应用
	$(GOCMD) run ./cmd/main.go

.PHONY: test
test: ## 运行测试
	$(GOTEST) -v ./...

.PHONY: clean
clean: ## 清理构建文件
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

.PHONY: deps
deps: ## 下载依赖
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: fmt
fmt: ## 格式化代码
	$(GOCMD) fmt ./...

# Docker相关命令
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	docker build -t $(DOCKER_IMAGE) .
	docker tag $(DOCKER_IMAGE) $(DOCKER_LATEST)

.PHONY: docker-run
docker-run: ## 运行Docker容器
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# Docker Compose相关命令
.PHONY: up
up: ## 启动所有服务
	docker-compose up -d

.PHONY: down
down: ## 停止所有服务
	docker-compose down

.PHONY: logs
logs: ## 查看服务日志
	docker-compose logs -f

.PHONY: restart
restart: ## 重启服务
	docker-compose restart

.PHONY: ps
ps: ## 查看服务状态
	docker-compose ps

# 健康检查
.PHONY: health
health: ## 检查服务健康状态
	@echo "检查服务健康状态..."
	@curl -f http://localhost:8080/health || echo "服务不可用"

# 一键构建和部署
.PHONY: all
all: clean deps build ## 完整构建流程

.PHONY: deploy
deploy: docker-build up ## 构建并部署 