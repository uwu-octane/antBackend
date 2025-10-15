#!/usr/bin/env bash

# PostgreSQL Master-Replica Docker Setup
set -euo pipefail

# Export environment variables
export DB_USER=postgres
export DB_PASSWORD=postgres_password
export DB_NAME_MASTER=antdb_master
export DB_NAME_REPLICA=antdb_replica

echo "Starting PostgreSQL Master-Replica Cluster..."
echo "================================"
echo "Master: localhost:5433"
echo "Replica: localhost:5434"
echo "================================"

# Clean up any existing containers and volumes
echo "🧹 Cleaning up existing containers..."
docker-compose down -v 2>/dev/null || true

# Start the cluster
echo "🔨 Building and starting containers..."
docker-compose up -d

echo ""
echo "Waiting for services to be healthy..."
sleep 5

# Check master status
echo ""
echo "Master Status:"
docker exec pg-master psql -U ${DB_USER} -d ${DB_NAME_MASTER} -c "SELECT version();" || true

# Wait a bit more for replica to bootstrap
echo ""
echo "Waiting for replica to bootstrap (this may take 30-60 seconds)..."
sleep 30

# Check replica status
echo ""
echo "Replica Status:"
docker exec pg-replica psql -U ${DB_USER} -d ${DB_NAME_MASTER} -c "SELECT version();" || true

echo ""
echo "Setup complete! Run './verify.sh' to verify replication."

