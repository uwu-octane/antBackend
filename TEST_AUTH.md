# Auth & User 服务测试说明

## 更新内容

### 1. 数据库迁移更新
已更新 `common/db/migrations/auth/0001_init_auth_users.sql`：
- ✅ 添加了 `username` 字段（varchar(64) unique）
- ✅ 添加了测试用户数据：
  - 用户名：`admin`
  - 密码：`admin123`
  - Email：`admin@example.com`

### 2. 测试脚本
创建了 `test_auth.sh` 测试脚本，测试以下功能：
- ✅ Ping 端点
- ✅ 用户登录
- ✅ 获取当前令牌信息 (/me)
- ✅ 获取用户信息 (/user/info)
- ✅ 刷新令牌
- ✅ 登出

## 使用步骤

### 1. 运行数据库迁移

首先需要运行数据库迁移来同步 auth_users 表结构：

```bash
# 使用 goose 运行迁移（假设已安装 goose）
cd common/db/migrations/auth
goose postgres "你的数据库连接字符串" up

# 或者重置并重新运行
goose postgres "你的数据库连接字符串" reset
goose postgres "你的数据库连接字符串" up
```

数据库连接字符串格式示例：
```
host=localhost port=5432 user=postgres password=your_password dbname=your_db sslmode=disable
```

### 2. 启动所有服务

确保以下服务正在运行：

```bash
# 启动 etcd（如果使用 etcd 做服务发现）
# ...

# 启动 Redis
# ...

# 启动 Auth RPC 服务
cd auth
go run auth.go -f etc/auth.yaml

# 启动 User RPC 服务
cd user
go run user.go -f etc/user.yaml

# 启动 Gateway API 服务
cd gateway
go run gateway.go -f etc/gateway-api.yaml
```

### 3. 运行测试脚本

```bash
# 在项目根目录下运行
./test_auth.sh
```

## 测试用户凭据

- **用户名**: `admin`
- **密码**: `admin123`
- **Email**: `admin@example.com`

## API 端点

Gateway API 运行在 `http://localhost:28256`

### 公开端点（无需认证）
- `GET /api/v1/ping` - 健康检查
- `POST /api/v1/login` - 用户登录
- `POST /api/v1/refresh` - 刷新令牌
- `POST /api/v1/logout` - 登出
- `POST /api/v1/logout-all` - 登出所有会话

### 需要认证的端点
- `GET /api/v1/me` - 获取当前令牌信息
- `GET /api/v1/user/info` - 获取用户信息

## 预期响应

### 登录成功响应
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "token_type": "bearer"
}
```

### 获取用户信息响应
```json
{
  "user_id": 123,
  "username": "admin",
  "email": "admin@example.com"
}
```

### /me 端点响应
```json
{
  "uid": "用户ID",
  "jti": "JWT ID",
  "iat": 1234567890
}
```

## 故障排查

### 1. 登录失败 - "invalid credentials"
- 检查数据库迁移是否成功执行
- 检查测试用户是否已插入到 auth_users 表
- 检查密码哈希是否正确

### 2. 获取用户信息失败
- 检查 users 表中是否存在对应的用户（email: admin@example.com）
- 检查 User RPC 服务是否正常运行
- 检查数据库连接配置

### 3. 服务连接失败
- 检查 etcd 是否运行
- 检查 Redis 是否运行
- 检查各个 RPC 服务的端口是否被占用
- 检查环境变量是否正确设置

### 4. 令牌验证失败
- 检查 JWT_SECRET 环境变量在所有服务中是否一致
- 检查时钟同步问题（LeewaySeconds 设置为 2 秒）

## 环境变量

确保设置以下环境变量：

```bash
# PostgreSQL
export PG_MASTER_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
export PG_REPLICA_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"

# Redis
export REDIS_HOST="localhost:6379"
export REDIS_PASSWORD=""

# etcd
export ETCD_HOST="localhost:2379"

# JWT
export JWT_SECRET="your-secret-key"

# CORS (可选)
export VITE_HOST="http://localhost:5173"
```

