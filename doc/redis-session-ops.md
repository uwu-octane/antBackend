# Redis Session 操作速查

当前后端会在 Redis 中管理登录会话、刷新令牌以及限流指标。默认前缀取自 `auth/etc/auth.yaml` 中的 `AuthRedis.Key`（默认为 `auth:`），若采用不同前缀，请自行替换以下示例。

## 主要 Key 结构
- `auth:refresh:<jti>`：Refresh Token 与用户 ID 的映射，TTL 等于刷新令牌有效期。
- `auth:reuse:<jti>`：被登出或复用检测到的 Refresh Token 标记，存在表示该令牌已失效。
- `auth:user:<uid>:sids`：某用户持有的所有 Session ID 集合（Set）。
- `auth:sid:<sid>`：某 Session 绑定的 Refresh Token JTI 集合（Set）。
- `auth:jti_sid:<jti>`：Refresh Token JTI 到 Session ID 的索引，便于追溯。
- `ratelimit:*`：Gateway 登录限流使用的令牌桶数据（Redis Key 来自 `gateway/etc/gateway-api.yaml` 中的 `RateLimitRedis.Key`）。

## 查询示例
```bash
# 查看全部 refresh token
redis-cli --scan --pattern 'auth:refresh:*'

# 检查某个 session 下绑定的 refresh jti
redis-cli SMEMBERS auth:sid:<sid>

# 反查 jti 对应的 session
redis-cli GET auth:jti_sid:<jti>

# 列出用户的全部 session
redis-cli SMEMBERS auth:user:<uid>:sids

# 查看登录限流计数
redis-cli --scan --pattern 'ratelimit:*'
```

## 清除与回收
```bash
# 清空单个 session（登出效果）
redis-cli DEL auth:sid:<sid>
redis-cli SREM auth:user:<uid>:sids <sid>

# 删除 refresh token 及其索引
redis-cli DEL auth:refresh:<jti> auth:jti_sid:<jti> auth:reuse:<jti>

# 批量清除所有 refresh token（谨慎）
redis-cli --scan --pattern 'auth:refresh:*' | xargs -r redis-cli DEL

# 重置登录限流窗口
redis-cli --scan --pattern 'ratelimit:*' | xargs -r redis-cli DEL
```

## 建议
- 操作前确认目标环境，并备份关键 Key（例如通过 `--scan | xargs redis-cli DUMP`）。
- 清理命令可能影响在线用户，请先在测试环境验证。
- 如果启用了命名空间或密码，记得附加 `-a $REDIS_PASSWORD` 与 `-n <db>` 参数。
