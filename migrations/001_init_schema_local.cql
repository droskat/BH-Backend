-- Local dev schema (single-node Cassandra)
-- Run with: cqlsh -f migrations/001_init_schema_local.cql

CREATE KEYSPACE IF NOT EXISTS image_platform
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE image_platform;

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
