#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

export CASSANDRA_HOSTS="${CASSANDRA_HOSTS:-127.0.0.1}"
export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
export KAFKA_BROKERS="${KAFKA_BROKERS:-127.0.0.1:9092}"
export SERVER_PORT="${SERVER_PORT:-8080}"

if command -v docker >/dev/null 2>&1; then
  echo "Starting infrastructure with Docker Compose..."
  docker compose up -d
  echo "Waiting for Cassandra..."
  for _ in $(seq 1 30); do
    if docker compose exec -T cassandra cqlsh -e "describe keyspaces" >/dev/null 2>&1; then
      break
    fi
    sleep 2
  done
  docker compose exec -T cassandra cqlsh -f /dev/stdin < migrations/001_init_schema_local.cql
else
  echo "Starting infrastructure with Homebrew services..."
  brew services start redis
  brew services start cassandra
  brew services start kafka

  echo "Waiting for Cassandra..."
  for _ in $(seq 1 60); do
    if cqlsh -e "describe keyspaces" >/dev/null 2>&1; then
      break
    fi
    sleep 2
  done

  cqlsh -f migrations/001_init_schema_local.cql
fi

echo "Building and starting API server on :${SERVER_PORT}..."
go build -o server ./cmd/server
./server
