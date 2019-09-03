-- +migrate Up
CREATE TABLE blocks (
    id      INTEGER PRIMARY KEY NOT NULL,
    height  INTEGER NOT NULL,
    bytes   BYTEA NOT NULL
);

CREATE TABLE utxo (
    id              INTEGER PRIMARY KEY NOT NULL,
    txid            BYTEA NOT NULL,
    idx             INTEGER NOT NULL,
    amount          INTEGER NOT NULL,
    script_pubkey   BYTEA NOT NULL
);

-- +migrate Down
DROP TABLE utxo;
DROP TABLE blocks;
