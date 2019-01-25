-- +migrate Up
-- CREATE TABLE escrow (id int);

-- N.B.: The field sizes here need to match the `common.*Size` constants.
CREATE TABLE logical_cache_mapping (
    escrow_id          bytea NOT NULL,
    slot_idx           bigint NOT NULL, -- XXX: int64, not uint64
    block_escrow_id    bytea NOT NULL,
    block_id           bytea NOT NULL,
    CONSTRAINT pk_escrow_slot PRIMARY KEY (escrow_id, slot_idx)
);

-- CREATE TABLE request (id int);

-- CREATE TABLE ticket (id int);

-- CREATE TABLE object_metadata (id int);


-- +migrate Down
-- DROP TABLE escrow;

DROP TABLE logical_cache_mapping;

-- DROP TABLE request;

-- DROP TABLE ticket;

-- DROP TABLE object_metadata;
