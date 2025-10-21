# JWT Session 化方案实施步骤

## 背景概览
- 现有认证流由 `gateway` HTTP 层调用 `auth` RPC 完成登陆、刷新与登出，前端需持久化 `refresh_token` 并在 `/refresh`、`/logout` 中回传。
- `auth` 服务已在 Redis 中维护 `sid` 体系：`auth:user:<uid>:sids`、`auth:sid:<sid>`、`auth:jti_sid:<jti>` 以及 `auth:refresh:<jti>`，但 `sid` 当前未暴露给前端。
- 登录与刷新逻辑使用 `TokenHelper` 生成 JWT，并通过 `dao.RefreshRotate` + Lua 脚本实现刷新时的 JTI 轮换与重放保护。

## 目标
以 `session_id` HTTP-Only Cookie 替代前端的 `refresh_token` 存储，通过服务端 Redis 管理 `session_id -> refresh_jti` 与用户会话，前端仅持久化 `access_token`（或直接依赖 Cookie），刷新与登出均基于 Cookie 自动携带的 `session_id`。

## 更新步骤
1. **协议与契约调整**
   - 在 `api/v1/auth/auth.proto` 的 `LoginResp` 增加 `session_id` 字段，`RefreshReq`/`LogoutReq` 改为携带 `session_id`（或新增消息体）。生成新的 gRPC 代码并同步 `gateway/gateway.api`、`gateway/internal/types`。
   - HTTP 层 `/api/v1/login` 响应结构去掉 `refresh_token`，改为仅返回访问令牌信息。

2. **Auth 服务改造**
   - `LoginLogic`: 沿用现有 `sid := uuid.NewString()` 作为会话 ID，新增 `redis.Setex(sessionKey(sid), refreshJti, TTL)`。登录响应返回 `session_id`，`refresh_token` 可继续返回给网关内部使用但不透出。
   - `RefreshLogic`: 将入参改为 `session_id`，通过 `sessionKey(sid)` 拿到当前 refresh JTI，再走现有 `ValidateRefreshToken`/`RefreshRotate` 流程；轮换后写回新的 refresh JTI，并保持 `auth:jti_sid`、`auth:sid:<sid>` 的同步。
   - `LogoutLogic` 与 `LogoutAllLogic`: 通过 `session_id` 找到关联 JTI，删除 `auth:refresh:<jti>`、`auth:jti_sid:<jti>`、`sessionKey(sid)`，并从用户/会话集合中移除。保留重放标记逻辑。
   - 补充 `sessionKey` 工具方法（如 `auth/internal/util`) 统一命名，TTL 与 `RefreshExpireSeconds` 对齐。

3. **Gateway HTTP 层**
   - `LoginHandler`: 调用 RPC 后，将 `session_id` 写入 `Set-Cookie`（`HttpOnly`、`Secure` 按环境配置、`SameSite=Lax/Strict`），响应体不再包含 refresh token。
   - `RefreshHandler` 与 `LogoutHandler`: 从 `http.Request` Cookie 读取 `session_id`，若缺失则返回未认证；调用新的 RPC 输入结构；刷新成功后更新 Cookie（如变更 session ID 时写回新值）。
   - 补充配置项（`gateway/etc/gateway-api.yaml`）用于设置 Cookie Domain、Secure 开关、SameSite 策略。

4. **Redis Schema & 迁移**
   - 新增 `auth:session:<sid>` 或复用现有 `sid` 集合存储最新 refresh JTI；确保 `dao.RefreshRotate`/Lua 与 session key 的写入/删除保持原子性或在应用层串联。
   - 考虑灰度：允许在配置中开启“会话模式”，期间同时接受旧的 `refresh_token` 请求（通过判空兼容），上线后关闭旧接口。

5. **前端/客户端对接**
   - 前端删除 `useRefreshToken` 全局状态，改为依赖 Cookie 自动携带；刷新接口使用空体或最小参数并设置 `credentials: 'include'`。
   - 调整登出流程同样传递 `credentials: 'include'`，清理任何本地遗留的 refresh token 缓存。

6. **测试与验证**
   - 更新单测：`auth/internal/logic/*` 新增 session 场景，覆盖 Redis session 缺失、JTI 轮换、Cookie 复用等路径。
   - 增补集成脚本：扩展 `test_auth.sh` 或新增 `test_session.sh`，验证登录→刷新→登出→重放失败的全链路。
   - 在测试环境开启 HTTPS/自签证书，确认浏览器端 Cookie 写入与跨域 CORS 设置正确；验证旧客户端的兼容策略。

7. **部署注意事项**
   - 部署前清理历史 Redis reuse 数据，或提供一次性脚本迁移现存 `auth:refresh:*` 至 `auth:session:*`。
   - 配置下发时确保 `JWT_SECRET` 在 gateway 与 auth 一致，另外提供 Cookie Domain、Secure、HTTPOnly 等环境变量。
   - 灰度阶段监控 Redis 会话键数量与刷新错误日志，确认 session 轮换正常后再移除旧的 refresh token API 支持。
