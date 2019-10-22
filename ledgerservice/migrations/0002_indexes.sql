-- +migrate Up

-- Only a single block with a given id is permitted in the ledger
CREATE UNIQUE INDEX block_block_id_idx ON block (block_id);

-- Pagination support for block explorer until we refactor block explorer to have a replica cache of its own.
CREATE UNIQUE INDEX block_height_block_id_idx ON block (height, block_id);

-- +migrate Down

DROP INDEX block_block_id_idx;
DROP INDEX block_height_block_id_idx;
