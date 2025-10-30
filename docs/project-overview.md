# 项目框架与功能概述

## 服务拓扑
- 整体基于 go-zero 微服务框架，`go.work` 将 `gateway`、`auth`、`user`、`common`、`cmd` 等模块组合为一个多模块工作区，方便共享依赖与统一构建。
- HTTP 接入层在 `gateway/`，负责 REST API、JWT 鉴权中间件、登录限流以及将请求路由到后端 RPC 或上游服务。
- 鉴权 RPC 服务位于 `auth/`，提供登录、刷新令牌、登出等能力，依赖 PostgreSQL 与 Redis 管理账户与会话。
- 用户信息 RPC 服务位于 `user/`，暴露用户资料读取接口，并通过读写分离策略访问数据库。
- 共享库收敛在 `common/`，提供环境变量加载与数据库读副本选择工具；`api/` 保存 goctl 生成的 gRPC/Protobuf 定义。
- `cmd/boot` 提供一键引导入口，按配置顺序启动 Gateway、Auth、User 三个服务。
- `ai/nuxt-ai` 基于 Nuxt 3 构建 AI 交互前端，经 Consul 注册后由 Gateway 按 `/nuxtapi/` 前缀代理统一对外。

## Gateway 模块
- 配置位于 `gateway/etc/gateway-api.yaml`，定义监听地址、Consul 发现、JWT 验证、登录限流与上游转发配置。
- `app/` 的 `BuildGatewayServer` 负责读取配置、初始化服务上下文并注册 REST 路由，同时挂载请求 ID、路径归一化与 JWT 校验等中间件。
- `internal/handler` 下的 `auth` 分组实现 `/api/v1/login`、`/refresh`、`/logout`、`/logout-all`、`/me` 等接口，`user` 分组负责 `/api/v1/user/info`。
- 登录接口调用 `internal/logic/auth.LoginLogic` 远程的 Auth RPC，并从 gRPC 响应头抽取刷新令牌写入 Cookie；刷新、登出逻辑也通过 gRPC 与 Auth 服务交互。
- `middleware/jwt.go` 解析并验证 Bearer Token，将用户标识、JTI、签发时间放入请求上下文供后续逻辑使用；路由白名单允许未登录访问登录与健康检查接口。
- `internal/svc/servicecontext.go` 构建 Auth/User RPC 客户端、登录 `PeriodLimit` 限流器，以及 Consul 上游代理管理器，支持按配置动态转发到 `Upstreams` 指定的 HTTP 服务。

## Auth 模块
- 配置 `auth/etc/auth.yaml` 定义 gRPC 监听、Consul 注册信息、JWT 过期时间、Redis Key 前缀以及 PostgreSQL 主从 DSN。
- `app/BuildAuthRpcServer` 与 `auth.go` 提供构建与独立启动两种入口，均会加载环境变量、创建 `svc.ServiceContext` 并注册 gRPC 服务实现。
- `internal/svc.ServiceContext` 统一持有 Redis 连接、PostgreSQL 主从连接、`singleflight.Group` 以及会话所需的 `TokenHelper` 和 DAO。
- `internal/logic` 中：
  - `LoginLogic` 校验密码（当前支持 bcrypt），生成访问令牌与刷新令牌，将用户会话与 JTI 关系写入 Redis（`auth:user:<uid>:sids`、`auth:sid:<sid>`、`auth:refresh:<jti>` 等）。
  - `RefreshLogic` 从 gRPC 元数据读取刷新令牌，依赖 Redis Lua 脚本 `dao/refresh_rotate.lua` 做 JTI 轮换、防重放，并通过 `TokenHelper` 下发新的令牌对。
  - `LogoutLogic` 与 `LogoutAllLogic` 根据 Session ID 清理 Redis 中的刷新令牌、标记复用并移除用户与会话索引。
  - `PingLogic` 作为存活检测。
- 数据访问层 `internal/model` 使用 `common/commonutil.Selector` 实现主从读写切换；当前提供按用户名/邮箱查询账号能力。

## User 模块
- 配置 `user/etc/user.yaml` 同样描述 Consul、PostgreSQL 主从与 Redis（预留），通过 `svc.ServiceContext` 建立数据库连接。
- `internal/model.UserModel` 使用与 Auth 相同的读写选择器，提供按用户 ID 读取资料的接口。
- `internal/logic.GetUserInfoLogic` 读取 JWT 中的用户 ID，调用模型查库并封装 gRPC 响应；`PingLogic` 提供基础健康检查。
- gRPC 入口由 goctl 生成的 `internal/server` 自动注册，Client 代码位于 `userservice/` 并供 Gateway 调用。

## Common 与 API 定义
- `common/envloader` 通过 `godotenv` 先后加载仓库根目录与当前模块下的 `.env` 文件，确保本地开发配置生效。
- `common/commonutil.Selector` 封装数据库读写优先级与回退逻辑，被 Auth/User 模块用于读副本优先、失败回退主库。
- `api/v1` 下保存 Auth 与 User 的 proto 文件及 goctl 生成的 gRPC Stub，保证服务与客户端使用同一套类型定义。

## AI/Nuxt Upstream 服务
- 代码位于 `ai/nuxt-ai`，采用 Nuxt 3 + Bun（亦支持 npm/pnpm/yarn）构建前端与 AI 功能界面，入口组件在 `app/app.vue`，服务端逻辑保存在 `server/`。
- `gateway/etc/gateway-api.yaml` 将 `nuxt-ai` 定义为上游服务：`PathPrefix` 为 `/nuxtapi/`，转发时保留原始路径（`StripPrefix` 留空），并透传 `Authorization`、`X-Request-Id` 头方便身份校验与链路追踪。
- `internal/handler.InitUpstreamProxies` 基于 Consul 发现实时维护目标实例，`NotFoundHandler` 中的 `UpstreamEntry` 兜底匹配 `/nuxtapi/` 前缀，将请求代理到 `nuxt-ai`，从而与 REST API 共用 28256 端口和域名。
- 本地开发时在 `ai/nuxt-ai` 目录执行 `bun install && bun run dev`（或对应包管理器命令）启动服务；部署后通过 Consul 注册 `nuxt-ai` 名称即可被 Gateway 自动拾取。

## 模块间通信流程
1. **登录**：Gateway 接收 `/api/v1/login` 请求，命中限流后转发给 Auth RPC。Auth 校验凭证、生成访问与刷新令牌，写入 Redis，并通过 gRPC Header 回传刷新令牌。Gateway 将访问令牌写入响应体、刷新令牌与 Session ID 写入 HttpOnly Cookie。
2. **刷新令牌**：Gateway 从 Cookie 读取 Session ID 与刷新令牌，调用 Auth 的 Refresh RPC。Auth 使用 Redis Lua 脚本轮换旧 JTI、生成新令牌并刷新会话 TTL，成功后 Gateway 更新 Cookie 并返回新的访问令牌。
3. **获取用户信息**：JWT 中间件验证访问令牌，将 UID 放入上下文。`/api/v1/user/info` 路由调用 User RPC，User 服务访问 PostgreSQL Replica 拉取资料并回传给 Gateway。
4. **登出/全体登出**：Gateway 将 Session ID 传递给 Auth RPC，Auth 清理 Redis 中对应的刷新令牌、重用标记与用户会话集合；`logout-all` 会额外遍历用户持有的全部 Session。
5. **上游转发**：Gateway 根据 `Upstreams` 配置通过 Consul 动态发现 HTTP 服务，`handler.UpstreamEntry` 在 404 场景兜底转发至注册的外部服务；当前主路由为 `nuxt-ai`，承载 `/nuxtapi/` 前缀的 AI 相关接口。

## 运维与运行入口
- `cmd/boot/main.go` 集成加载 `.env`、校验配置路径，并按 Gateway → Auth → User 的顺序构建服务，统一纳入 go-zero 的 `ServiceGroup` 管理启动与停止。
- 单服务可分别通过 `gateway/gateway.go`、`auth/auth.go`、`user/user.go` 启动；集成运行可执行 `go run ./cmd/boot -gateway gateway/etc/gateway-api.yaml -auth auth/etc/auth.yaml -user user/etc/user.yaml`。
- Consul 用于服务注册与发现；Redis 存储登录态、刷新令牌、登录限流指标；PostgreSQL 提供用户与账号数据的主从存储。

## 测试与辅助脚本
- 单元测试示例位于 `auth/internal/logic`（如 `logoutlogic_test.go`、`takecareofside_test.go`），验证会话与 Redis 交互逻辑。
- `go test ./...` 可在仓库根目录触发全部模块测试，需提前执行 `go work sync` 保持依赖一致。
- `./test_ratelimit.sh` 手工校验 Gateway 登录限流行为，默认依赖服务运行在 `http://127.0.0.1:28256`。
