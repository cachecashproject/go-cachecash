-- +migrate Up
CREATE TABLE raw_block (
    blockid BLOB PRIMARY KEY NOT NULL,
    height  INTEGER NOT NULL,
    bytes   BLOB NOT NULL
);

CREATE TABLE raw_tx (
    txid    BLOB PRIMARY KEY NOT NULL,
    bytes   BLOB NOT NULL
);

-- +migrate Down
DROP TABLE raw_tx;
DROP TABLE raw_block;
