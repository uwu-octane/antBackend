#!/usr/bin/env bash

set -euo pipefail

PGDATA="${PGDATA:-/var/lib/postgresql/data}"
REPL_HOST="${REPL_HOST:-pg-master}"
REPL_PORT="${REPL_PORT:-5432}"
REPL_USER="${REPL_USER:-replicator}"
REPL_PASSWORD="${REPL_PASSWORD:-repl_password}"
REPL_SLOT="${REPL_SLOT:-phys_slot_1}"

echo "=== Replica Bootstrap Script ==="

# Check if PGDATA is empty or not initialized as replica
if [ -z "$(ls -A "$PGDATA" 2>/dev/null)" ] || [ ! -f "$PGDATA/standby.signal" ]; then
    echo ">> Initializing replica from ${REPL_HOST}:${REPL_PORT}"
    
    # Clean data directory if it exists but is not a replica
    if [ -n "$(ls -A "$PGDATA" 2>/dev/null)" ]; then
        echo ">> Cleaning existing non-replica data..."
        rm -rf "${PGDATA:?}"/*
    fi

    export PGPASSWORD="$REPL_PASSWORD"

    echo ">> Waiting for master to be ready..."
    until pg_isready -h "$REPL_HOST" -p "$REPL_PORT" -U "$REPL_USER"; do
        echo "   ... still waiting"
        sleep 2
    done
    echo ">> Master is ready!"

    echo ">> Starting pg_basebackup (this may take a while)..."
    pg_basebackup \
        -h "$REPL_HOST" \
        -p "$REPL_PORT" \
        -U "$REPL_USER" \
        -D "$PGDATA" \
        -R \
        -X stream \
        -S "$REPL_SLOT" \
        -P \
        --wal-method=stream \
        -v

    echo ">> Setting permissions..."
    chown -R postgres:postgres "$PGDATA"
    
    echo ">> Replica bootstrap completed successfully!"
else
    echo ">> Replica already initialized, skipping bootstrap"
fi