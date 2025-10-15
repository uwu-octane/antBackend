# 数据库迁移工具

本目录包含使用 [goose](https://github.com/pressly/goose) 进行数据库迁移的工具和脚本。

## 目录结构

```
migrations/
├── Makefile          # Make 命令方式
├── goose.sh          # Shell 脚本方式
├── README.md         # 本文档
├── auth/             # Auth 服务迁移文件
│   └── 0001_init_auth_users.sql
└── user/             # User 服务迁移文件
    └── 0001_init_users.sql
```

## 前置条件

1. **PostgreSQL 主从集群已启动**
   ```bash
   cd ../
   ./start.sh
   ```

2. **环境变量配置**
   
   确保项目根目录的 `.env` 文件包含以下配置：
   ```env
   DB_USER=postgres
   DB_PASSWORD=postgres_password
   DB_HOST=localhost
   DB_PORT=5433
   DB_NAME=antdb_master
   ```

3. **安装 Goose**
   ```bash
   # 使用 Makefile
   make goose.install
   
   # 或使用脚本
   ./goose.sh install
   ```

## 使用方法

### 方式一：使用 Makefile

```bash
cd /Users/taoting/Documents/antD/antBackend/common/db/migrations

# 查看帮助
make help

# 安装 goose
make goose.install

# 查看 goose 版本
make goose.version

# Auth 迁移
make goose.auth.status    # 查看状态
make goose.auth.up        # 执行迁移
make goose.auth.down      # 回滚一个版本
make goose.auth.reset     # 重置所有迁移

# User 迁移
make goose.user.status    # 查看状态
make goose.user.up        # 执行迁移
make goose.user.down      # 回滚一个版本
make goose.user.reset     # 重置所有迁移

# 批量操作
make goose.up.all         # 执行所有迁移
make goose.status.all     # 查看所有状态
```

### 方式二：使用 Shell 脚本

```bash
cd /Users/taoting/Documents/antD/antBackend/common/db/migrations

# 查看帮助
./goose.sh help

# 安装 goose
./goose.sh install

# 查看版本
./goose.sh version

# Auth 迁移
./goose.sh auth:status    # 查看状态
./goose.sh auth:up        # 执行迁移
./goose.sh auth:down      # 回滚一个版本
./goose.sh auth:reset     # 重置所有迁移

# User 迁移
./goose.sh user:status    # 查看状态
./goose.sh user:up        # 执行迁移
./goose.sh user:down      # 回滚一个版本
./goose.sh user:reset     # 重置所有迁移

# 批量操作
./goose.sh all:up         # 执行所有迁移
./goose.sh all:status     # 查看所有状态
```

## 快速开始

### 1. 首次部署

```bash
# 启动数据库
cd /Users/taoting/Documents/antD/antBackend/common/db
./start.sh

# 等待数据库启动完成后，执行迁移
cd migrations
make goose.up.all
```

### 2. 查看迁移状态

```bash
make goose.status.all
```

输出示例：
```
==========================================
📊 All Migration Status
==========================================

🔐 AUTH Migrations:
    Applied At                  Migration
    =======================================
    Pending                  -- 0001_init_auth_users.sql

👤 USER Migrations:
    Applied At                  Migration
    =======================================
    Pending                  -- 0001_init_users.sql
==========================================
```

### 3. 执行迁移

```bash
# 执行 auth 迁移
make goose.auth.up

# 执行 user 迁移
make goose.user.up
```

### 4. 验证迁移结果

```bash
# 连接主库查看
docker exec -it pg-master psql -U postgres -d antdb_master

# 查看 auth 表
\dt auth.*

# 查看 user 表
\dt public.*
```

## 注意事项

⚠️ **重要提示**

1. **只对主库执行迁移**：所有迁移命令只在主库 (pg-master) 上执行，从库会自动通过复制同步。

2. **生产环境谨慎操作**：
   - 执行 `down` 和 `reset` 命令会删除数据
   - 建议在执行前备份数据库

3. **环境变量**：
   - 工具会自动加载项目根目录的 `.env` 文件
   - 不会修改现有的 `.env` 文件

4. **连接信息**：
   - 主库地址: `localhost:5433`
   - 数据库名: `antdb_master`
   - 从库会自动同步主库的所有更改

## 创建新的迁移

使用 goose 创建新的迁移文件：

```bash
# 创建 auth 迁移
goose -dir auth create add_new_table sql

# 创建 user 迁移
goose -dir user create add_new_column sql
```

这会在对应目录下创建一个新的 SQL 文件，格式如：`YYYYMMDDHHMMSS_add_new_table.sql`

## 迁移文件格式

Goose 迁移文件示例：

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

## 故障排查

### Goose 未安装

```bash
Error: goose: command not found
```

**解决方案**：
```bash
make goose.install
```

### 数据库连接失败

```bash
Error: failed to connect to database
```

**解决方案**：
1. 检查数据库是否启动：`docker ps | grep pg-master`
2. 检查 `.env` 配置是否正确
3. 确认端口 5433 未被占用

### 权限问题

```bash
Error: permission denied
```

**解决方案**：
```bash
chmod +x goose.sh
```

## 相关链接

- [Goose 官方文档](https://github.com/pressly/goose)
- [PostgreSQL 迁移最佳实践](https://www.postgresql.org/docs/current/backup-dump.html)
- [数据库主从复制说明](../README.md)

## 维护者

如有问题，请联系项目维护者或提交 Issue。

