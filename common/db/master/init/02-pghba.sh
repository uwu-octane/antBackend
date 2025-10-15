#!/usr/bin/env bash

set -euo pipefail
PGDATA="${PGDATA:-/var/lib/postgresql/data}"

# Allow replication connection from replica container
echo "host replication replicator 0.0.0.0/0 scram-sha-256" >> "${PGDATA}/pg_hba.conf"
