# Proto 代码生成指南

## 📋 目录结构

```
antBackend/
├── api/
│   ├── api/                    # proto 文件统一存放目录
│   │   ├── auth/v1/auth.proto
│   │   └── user/v1/user.proto
│   └── gen/go/                 # 生成的 pb.go 文件
│       ├── auth/
│       │   ├── auth.pb.go
│       │   └── auth_grpc.pb.go
│       └── user/
│           ├── user.pb.go
│           └── user_grpc.pb.go
├── auth/                       # auth 微服务
├── user/                       # user 微服务
├── gateway/                    # 网关服务
├── .goctl.yaml                 # goctl 配置文件
└── Makefile                    # 构建脚本
```

## 🚀 快速开始

### 1. 查看所有可用命令

```bash
make help
# 或者
make
```

### 2. 生成 pb.go 文件（推荐）

#### 生成单个服务

```bash
make proto-auth    # 只生成 auth 的 pb.go
make proto-user    # 只生成 user 的 pb.go
```

#### 生成所有服务

```bash
make proto-all
```

#### 特点
- ✅ **只生成** `*.pb.go` 和 `*_grpc.pb.go` 文件
- ✅ 文件输出到 `api/gen/go/` 目录
- ✅ **不会**生成完整的服务框架代码
- ✅ **不需要**目标服务目录存在 go.mod
- ✅ 适合大多数场景

### 3. 生成完整的 RPC 服务框架（高级）

如果需要生成完整的微服务框架（包括 logic、handler、config 等）：

```bash
make gen-service-auth    # 生成 auth 完整服务
make gen-service-user    # 生成 user 完整服务
```

#### 特点
- 🏗️ 生成完整的 RPC 服务代码结构
- 🏗️ 自动创建 `internal/logic`、`internal/server` 等目录
- 🏗️ 自动初始化 go.mod（如果不存在）
- 🏗️ 适合新建微服务

## 📝 添加新的 Proto 服务

### 步骤 1: 创建 proto 文件

在 `api/api/` 下创建新的 proto 文件：

```bash
mkdir -p api/api/order/v1
```

创建 `api/api/order/v1/order.proto`：

```protobuf
syntax = "proto3";

package order.v1;
option go_package = "api/api/order/v1";

message CreateOrderRequest {
  string user_id = 1;
  repeated string product_ids = 2;
}

message CreateOrderResponse {
  string order_id = 1;
  string status = 2;
}

service Order {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
}
```

### 步骤 2: 在 Makefile 中添加目标

在 Makefile 中添加：

```makefile
# 生成 order 服务的 pb.go 文件
proto-order:
	@echo "📦 生成 order 的 pb.go 文件..."
	@protoc --go_out=./ --go_opt=paths=source_relative \
		--go-grpc_out=./ --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/order/v1/order.proto
	@echo "✅ order pb.go 生成完成: $(GEN_DIR)/order/"

# 生成 order 完整服务框架
gen-service-order:
	@echo "🏗️  生成 order 完整服务框架..."
	@if [ ! -d "order" ]; then mkdir -p order; fi
	@cd order && if [ ! -f "go.mod" ]; then \
		go mod init order; \
	fi
	goctl rpc protoc $(PROTO_DIR)/order/v1/order.proto \
		--go_out=./ \
		--go-grpc_out=./ \
		--zrpc_out=./order
	@echo "✅ order 服务框架生成完成: ./order/"
```

更新 `proto-all` 目标：

```makefile
proto-all:
	@make proto-auth
	@make proto-user
	@make proto-order  # 添加这一行
```

### 步骤 3: 生成代码

```bash
make proto-order
# 或
make proto-all
```

## 🔧 配置说明

### .goctl.yaml

```yaml
# Proto 生成的 pb.go 文件输出目录
rpc:
  pbOutput: "./api/gen/go"  # pb.go 文件统一放在这里
```

### Makefile 变量

```makefile
PROTO_DIR = api/api          # proto 文件目录
GEN_DIR = api/gen/go         # pb.go 输出目录
```

## 🧹 清理命令

### 清理生成的 pb.go 文件

```bash
make clean-proto
```

### 格式化 proto 文件

```bash
make fmt-proto
```

## ⚠️ 常见问题

### 1. 错误：protoc: command not found

**解决方法：**

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt install protobuf-compiler

# 验证安装
protoc --version
```

### 2. 错误：protoc-gen-go: program not found or is not executable

**解决方法：**

```bash
# 安装 protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 确保 $GOPATH/bin 在 PATH 中
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 3. 错误：go: open .../go.mod: no such file or directory

这个错误只会在使用 `gen-service-*` 命令时出现。

**解决方法：**
- 使用 `make proto-*` 命令替代（推荐）
- 或者手动创建目标目录和 go.mod

### 4. pb.go 文件没有生成到正确的位置

**检查：**
1. proto 文件中的 `option go_package` 设置是否正确
2. `.goctl.yaml` 中的 `pbOutput` 配置
3. 确保使用 `--go_opt=paths=source_relative` 参数

## 📚 参考文档

- [go-zero 官方文档](https://go-zero.dev/en/docs/tutorials/cli/api)
- [Protocol Buffers 文档](https://protobuf.dev/)
- [gRPC Go 快速开始](https://grpc.io/docs/languages/go/quickstart/)

## 🎯 最佳实践

### 1. Proto 文件组织

```
api/api/
├── auth/v1/auth.proto        # 认证服务
├── user/v1/user.proto        # 用户服务
├── order/v1/order.proto      # 订单服务
└── common/v1/common.proto    # 公共类型
```

### 2. go_package 配置

```protobuf
// ✅ 推荐：使用统一的路径
option go_package = "api/api/{service}/v1";

// ❌ 不推荐：使用绝对路径
option go_package = "github.com/your/project/api/api/{service}/v1";
```

### 3. 版本管理

- 使用 `v1`、`v2` 等版本号管理 API 变更
- 向后兼容时在 v1 中修改
- 破坏性变更时创建 v2

### 4. 命名规范

- Service 名称：PascalCase（如 `UserService`）
- RPC 方法：PascalCase（如 `GetUser`）
- Message 名称：PascalCase（如 `GetUserRequest`）
- 字段名称：snake_case（如 `user_id`）

## 🔄 工作流程

```bash
# 1. 编写/修改 proto 文件
vim api/api/auth/v1/auth.proto

# 2. 生成代码
make proto-auth

# 3. 使用生成的代码
# 在你的服务中 import "api/gen/go/auth"

# 4. 提交代码
git add api/
git commit -m "feat: update auth proto"
```

## 💡 提示

- 大多数情况下使用 `make proto-*` 命令即可
- 只在创建新服务时使用 `make gen-service-*`
- 定期运行 `make proto-all` 确保所有 pb.go 文件是最新的
- 将生成的 pb.go 文件提交到 git（方便团队协作）

