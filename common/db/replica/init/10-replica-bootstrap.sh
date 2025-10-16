#!/bin/sh
set -euo pipefail

echo "=== Replica Bootstrap Script (idempotent) ==="

: "${PGDATA:=/var/lib/postgresql/data}"
: "${REPL_HOST:?REPL_HOST required}"
: "${REPL_PORT:=5432}"
: "${REPL_USER:?REPL_USER required}"
: "${REPL_PASSWORD:?REPL_PASSWORD required}"
: "${REPL_SLOT:=phys_slot_1}"

# 容器里所有 pg 命令都要显式用户/密码，避免默认 root
export PGPASSWORD="$REPL_PASSWORD"

# 如果已经是备库就直接跳过（存在 standby.signal）
if [ -f "$PGDATA/standby.signal" ]; then
  echo "standby.signal exists; already a standby. Skipping basebackup."
  exit 0
fi

# 如果数据目录非空但不是备库，报错并让你删卷重来（防止脏数据）
if [ -s "$PGDATA/PG_VERSION" ] && [ ! -f "$PGDATA/standby.signal" ]; then
  echo "!! $PGDATA exists but not a standby (no standby.signal). Remove the replica volume and retry."
  exit 1
fi

rm -rf "$PGDATA"/*
mkdir -p "$PGDATA"
chown -R postgres:postgres "$PGDATA"

# 用 pg_basebackup 克隆主库，并自动写入 standby.signal / postgresql.auto.conf（-R）
# -X stream 复制WAL，-C -S 创建/绑定物理槽（与主库 max_replication_slots 配合）
su - postgres -c "PGPASSWORD=$REPL_PASSWORD pg_basebackup -h $REPL_HOST -p $REPL_PORT -U $REPL_USER \
  -D $PGDATA -R -X stream -S $REPL_SLOT --progress --verbose"

# 可选：为了只读查询更顺滑
echo "hot_standby = on" >> "$PGDATA/postgresql.conf"

echo "=== Bootstrap finished; ready to start postgres as standby ==="