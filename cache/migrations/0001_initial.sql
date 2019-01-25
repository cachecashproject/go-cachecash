-- +migrate Up
-- CREATE TABLE escrow (id int);

-- N.B.: The field sizes here need to match the `common.*Size` constants.
CREATE TABLE logical_cache_mapping (
    escrow_id   varbinary(16) NOT NULL,
    slot_idx    unsigned int(4) NOT NULL,
    datum_id    varbinary(16) NOT NULL,
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
