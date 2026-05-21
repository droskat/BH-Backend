-- Image Analysis Platform Lite - Cassandra Schema Migration
-- Run with: cqlsh -f migrations/001_init_schema.cql

CREATE KEYSPACE IF NOT EXISTS image_platform
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 3};

USE image_platform;

-- Access Pattern 1: Fetch metadata for a specific asset by its unique ID
CREATE TABLE IF NOT EXISTS images_by_id (
    image_id uuid,
    user_id uuid,
    original_filename text,
    upload_date timestamp,
    width int,
    height int,
    file_size bigint,
    file_type text,
    status text,
    PRIMARY KEY (image_id)
);

-- Access Pattern 2: List all assets owned by a user, sorted newest first
CREATE TABLE IF NOT EXISTS images_by_user (
    user_id uuid,
    upload_date timestamp,
    image_id uuid,
    original_filename text,
    file_size bigint,
    file_type text,
    status text,
    PRIMARY KEY (user_id, upload_date)
) WITH CLUSTERING ORDER BY (upload_date DESC);
