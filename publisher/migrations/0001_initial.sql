-- +migrate Up

CREATE TYPE escrow_state AS ENUM ('ok', 'aborted');

CREATE TABLE escrow (
    id SERIAL PRIMARY KEY,
    txid BYTEA NOT NULL,
    start_block SERIAL NOT NULL,
    end_block SERIAL NOT NULL,
    state escrow_state NOT NULL,
    public_key BYTEA NOT NULL,
    private_key BYTEA NOT NULL,
    raw BYTEA NOT NULL
);

CREATE TABLE bundle (
    id SERIAL PRIMARY KEY,
    escrow_id SERIAL NOT NULL REFERENCES escrow(id),
    block_id SERIAL NOT NULL,
    raw BYTEA NOT NULL,
    request_sequence_no SERIAL NOT NULL,
    client_public_key VARCHAR NOT NULL,
    objectid VARCHAR NOT NULL
);

CREATE TABLE cache (
    id SERIAL PRIMARY KEY,
    public_key BYTEA NOT NULL,
    inetaddr BYTEA NOT NULL,
    inet6addr BYTEA NOT NULL,
    port INT NOT NULL
);
CREATE UNIQUE INDEX cache_public_key ON cache (public_key);

CREATE TABLE escrow_caches (
    id SERIAL PRIMARY KEY,
    escrow_id SERIAL NOT NULL REFERENCES escrow(id),
    cache_id SERIAL NOT NULL REFERENCES cache(id),
    inner_master_key BYTEA NOT NULL
);


-- +migrate Down
DROP TABLE escrow_caches;
DROP TABLE cache;
DROP TABLE bundle;
DROP TABLE ticket;
DROP TABLE escrow;
