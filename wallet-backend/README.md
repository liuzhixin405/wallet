# 区块链钱包后端

这是一个基于 Golang 和 Gin 框架开发的区块链钱包后端系统。

## 功能特性

- 用户认证和授权（JWT）
- 多链地址管理（比特币、以太坊等）
- 余额查询和管理
- 充值和提现记录
- 交易记录管理
- 多币种支持
- 区块链节点集成

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin Gonic
- **数据库**: MySQL
- **ORM**: GORM
- **认证**: JWT
- **配置管理**: Viper
- **区块链库**: 
  - go-ethereum (以太坊)
  - btcd (比特币)

## 项目结构

```
wallet-backend/
├── cmd/
│   └── main.go                 # 主程序入口
├── config.yaml                 # 配置文件
├── go.mod                      # Go模块文件
├── internal/
│   ├── config/                 # 配置管理
│   ├── database/               # 数据库连接和迁移
│   ├── handlers/               # HTTP处理器
│   ├── middleware/             # 中间件
│   ├── models/                 # 数据模型
│   ├── services/               # 业务逻辑服务
│   └── utils/                  # 工具函数
├── pkg/
│   ├── wallet/                 # 钱包相关功能
│   └── blockchain/             # 区块链集成
└── README.md                   # 项目说明
```

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis (可选，用于缓存)

### 安装依赖

```bash
go mod tidy
```

### 配置数据库

1. 创建MySQL数据库
2. 修改 `config.yaml` 中的数据库连接信息

### 运行应用

```bash
go run cmd/main.go
```

应用将在 `http://localhost:8080` 启动

## API接口

### 认证接口

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录

### 地址管理

- `GET /api/addresses` - 获取用户地址列表
- `POST /api/addresses/generate` - 生成新地址
- `POST /api/addresses/bind` - 绑定外部地址

### 余额管理

- `GET /api/balances` - 获取用户余额列表
- `GET /api/balances/:currency` - 获取指定币种余额

### 提现管理

- `GET /api/withdraws` - 获取提现记录
- `POST /api/withdraws` - 创建提现申请
- `GET /api/withdraws/:id` - 获取提现详情

### 充值记录

- `GET /api/deposits` - 获取充值记录
- `GET /api/deposits/:id` - 获取充值详情
- `GET /api/deposits/address/:address` - 按地址查询充值

### 交易记录

- `GET /api/transactions` - 获取交易记录
- `GET /api/transactions/:id` - 获取交易详情
- `GET /api/transactions/address/:address` - 按地址查询交易
- `api/transactions/currency/:currency` - 按币种查询交易

### 货币配置

- `GET /api/currencies` - 获取支持的货币列表
- `GET /api/currencies/:symbol` - 获取指定货币配置
- `GET /api/currencies/chains/supported` - 获取支持的链类型

## 数据库表结构

系统包含以下主要数据表：

- `users` - 用户信息
- `address_library` - 地址库
- `balance` - 用户余额
- `withdraw_record` - 提现记录
- `deposit_record` - 充值记录
- `chain_bill` - 链上交易记录
- `currency_chain_config` - 货币链配置

## 配置说明

主要配置项在 `config.yaml` 文件中：

```yaml
server:
  port: 8080

database:
  host: localhost
  port: 3306
  username: root
  password: password
  name: wallet_db

jwt:
  secret: your-jwt-secret
  expire_hours: 24

wallet:
  master_key: your-master-key
  hd_path: "m/44'/60'/0'/0"
```

## 开发说明

### 添加新的区块链支持

1. 在 `pkg/blockchain/` 下创建新的客户端文件
2. 实现相应的接口方法
3. 在 `internal/services/` 中添加对应的服务逻辑
4. 更新配置和路由

### 添加新的API接口

1. 在 `internal/handlers/` 下创建处理器
2. 在 `cmd/main.go` 中添加路由
3. 更新相应的服务层逻辑

## 部署

### Docker部署

```bash
docker build -t wallet-backend .
docker run -p 8080:8080 wallet-backend
```

### 生产环境

1. 使用环境变量覆盖敏感配置
2. 配置HTTPS
3. 设置适当的日志级别
4. 配置监控和告警

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！