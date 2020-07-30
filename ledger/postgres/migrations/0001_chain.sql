-- +migrate Up
CREATE TABLE raw_block (
    blockid BYTEA PRIMARY KEY,
    height  BIGINT NOT NULL,
    bytes   BYTEA NOT NULL
);

CREATE TABLE raw_tx (
    txid    BYTEA PRIMARY KEY,
    bytes   BYTEA NOT NULL
);

-- +migrate Down
DROP TABLE raw_tx;
DROP TABLE raw_block;
