-- +migrate Up

CREATE TABLE block (
    rowid SERIAL PRIMARY KEY,
    height INT NOT NULL, -- 0 for the genesis block.
    block_id BYTEA NOT NULL,
    parent_id BYTEA,  -- Null only for the genesis block.
    raw BYTEA NOT NULL
);

-- Well-formed and properly signed transactions that have not yet been included in a block.
CREATE TABLE mempool_transaction (
    rowid SERIAL PRIMARY KEY,
    raw BYTEA NOT NULL
);

-- @KK: Eventually, we probably want a table that indexes mined transactions by ID.  We'll still need to store the raw
--      (serialized) blocks, though.  Also, this gets a little bit tricky, because shallow forks may cause the same
--      transaction to be part of multiple blocks at similar heights.

-- +migrate Down
DROP TABLE mempool_transaction;
DROP TABLE block;
