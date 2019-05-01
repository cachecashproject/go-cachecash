-- +migrate Up

-- TODO: some of those fields should be nullable
CREATE TABLE caches (
    public_key BYTEA NOT NULL,

    -- software version, including both semantic/product versions and build number
    version VARCHAR NOT NULL,

    -- free and total available resources (memory, disk, etc.)
    free_memory BIGINT NOT NULL,
    total_memory BIGINT NOT NULL,

    free_disk BIGINT NOT NULL,
    total_disk BIGINT NOT NULL,

    -- their uptime
    startup_time DATETIME NOT NULL,

    --network interface details, including e.g. externally-visible IP addresses
    -- TODO: we need explicit endpoints
    -- interfaces VARCHAR,
    external_ip VARCHAR NOT NULL,
    port INTEGER NOT NULL,

    -- a URI where interested publishers can contact them
    contact_url VARCHAR NOT NULL,

    -- we use this to remove stale announcements
    last_ping DATETIME NOT NULL,

    CONSTRAINT pk_cache_fp PRIMARY KEY(public_key)
);

-- +migrate Down

DROP TABLE caches;
