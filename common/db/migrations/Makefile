# Goose Migration Makefile
# 
# 用法:
#   make goose.install          - 安装 goose 工具
#   make goose.auth.status      - 查看 auth 迁移状态
#   make goose.auth.up          - 执行 auth 迁移
#   make goose.auth.down        - 回滚 auth 迁移
#   make goose.user.status      - 查看 user 迁移状态
#   make goose.user.up          - 执行 user 迁移
#   make goose.user.down        - 回滚 user 迁移

# 加载项目根目录的 .env 文件
ifneq (,$(wildcard ../../../.env))
    include ../../../.env
    export
endif

# 主库端口：5433 (docker-compose.yml: 5433:5432)
# 从库端口：5434 (docker-compose.yml: 5434:5432)
DB_USER ?= ${DB_USER}
DB_PASSWORD ?= ${DB_PASSWORD}
DB_HOST ?= ${DB_HOST}
DB_PORT ?= 5433
DB_NAME ?= ${DB_NAME}

# 主库 DSN 连接字符串
PG_MASTER_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# 迁移目录（相对于当前 Makefile 的位置）
MIGRATIONS_DIR = .

# ==== goose 安装 ====
.PHONY: goose.install
goose.install:
	@echo "📦 Installing goose..."
	go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "✅ Goose installed successfully"

# ==== goose 版本检查 ====
.PHONY: goose.version
goose.version:
	@goose -version || echo "⚠️  Goose not installed. Run 'make goose.install' first."

# ==== AUTH 迁移（只对主库） ====
.PHONY: goose.auth.status
goose.auth.status:
	@echo "📊 Checking auth migration status..."
	@goose -dir $(MIGRATIONS_DIR)/auth postgres "$(PG_MASTER_DSN)" status

.PHONY: goose.auth.up
goose.auth.up:
	@echo "⬆️  Running auth migrations up..."
	@goose -dir $(MIGRATIONS_DIR)/auth postgres "$(PG_MASTER_DSN)" up
	@echo "✅ Auth migrations completed"

.PHONY: goose.auth.down
goose.auth.down:
	@echo "⬇️  Rolling back auth migration..."
	@goose -dir $(MIGRATIONS_DIR)/auth postgres "$(PG_MASTER_DSN)" down
	@echo "✅ Auth migration rolled back"

.PHONY: goose.auth.reset
goose.auth.reset:
	@echo "🔄 Resetting auth migrations..."
	@goose -dir $(MIGRATIONS_DIR)/auth postgres "$(PG_MASTER_DSN)" reset
	@echo "✅ Auth migrations reset"

# ==== USER 迁移（只对主库） ====
.PHONY: goose.user.status
goose.user.status:
	@echo "📊 Checking user migration status..."
	@goose -dir $(MIGRATIONS_DIR)/user postgres "$(PG_MASTER_DSN)" status

.PHONY: goose.user.up
goose.user.up:
	@echo "⬆️  Running user migrations up..."
	@goose -dir $(MIGRATIONS_DIR)/user postgres "$(PG_MASTER_DSN)" up
	@echo "✅ User migrations completed"

.PHONY: goose.user.down
goose.user.down:
	@echo "⬇️  Rolling back user migration..."
	@goose -dir $(MIGRATIONS_DIR)/user postgres "$(PG_MASTER_DSN)" down
	@echo "✅ User migration rolled back"

.PHONY: goose.user.reset
goose.user.reset:
	@echo "🔄 Resetting user migrations..."
	@goose -dir $(MIGRATIONS_DIR)/user postgres "$(PG_MASTER_DSN)" reset
	@echo "✅ User migrations reset"

# ==== 全部迁移 ====
.PHONY: goose.up.all
goose.up.all: goose.auth.up goose.user.up
	@echo "✅ All migrations completed"

.PHONY: goose.status.all
goose.status.all:
	@echo "=========================================="
	@echo "📊 All Migration Status"
	@echo "=========================================="
	@echo ""
	@echo "🔐 AUTH Migrations:"
	@make --no-print-directory goose.auth.status
	@echo ""
	@echo "👤 USER Migrations:"
	@make --no-print-directory goose.user.status
	@echo ""
	@echo "=========================================="

# ==== 帮助信息 ====
.PHONY: help
help:
	@echo "🗄️  Goose Migration Commands"
	@echo ""
	@echo "Installation:"
	@echo "  make goose.install          - Install goose tool"
	@echo "  make goose.version          - Check goose version"
	@echo ""
	@echo "Auth Migrations:"
	@echo "  make goose.auth.status      - Check auth migration status"
	@echo "  make goose.auth.up          - Run auth migrations"
	@echo "  make goose.auth.down        - Rollback auth migration"
	@echo "  make goose.auth.reset       - Reset all auth migrations"
	@echo ""
	@echo "User Migrations:"
	@echo "  make goose.user.status      - Check user migration status"
	@echo "  make goose.user.up          - Run user migrations"
	@echo "  make goose.user.down        - Rollback user migration"
	@echo "  make goose.user.reset       - Reset all user migrations"
	@echo ""
	@echo "Batch Operations:"
	@echo "  make goose.up.all           - Run all migrations"
	@echo "  make goose.status.all       - Check all migration status"
	@echo ""
	@echo "Configuration:"
	@echo "  PG_MASTER_DSN: $(PG_MASTER_DSN)"
	@echo ""

.DEFAULT_GOAL := help

