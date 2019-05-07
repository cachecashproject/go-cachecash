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

CREATE TABLE escrow (
    id                     int PRIMARY KEY NOT NULL,
    escrow_id              bytea NOT NULL,
    inner_master_key       varchar NOT NULL,
    outer_master_key       varchar NOT NULL,
    slots                  int NOT NULL,
    publisher_cache_addr   varchar NOT NULL
);

-- -- QQ: Do we want ticket numbers to be per-escrow or per-escrow-per-block?
-- CREATE TABLE ticket_l1 (id int) (
--     escrow_id          bytea NOT NULL,
--     block_no           bytea NOT NULL,
--     ticket_no          bigint NOT NULL,
--     raw                blob NOT NULL,
--     CONSTRAINT pk_escrow_ticket PRIMARY KEY (escrow_id, ticket_no)
--     -- Probably also want an index by block number, because we'll be selecting tickets that way.
--     -- Probably also need some state fields: has this ticket won/lost yet?  If it won, have we redeemed it? etc.
-- );

-- CREATE TABLE request (id int);

-- CREATE TABLE object_metadata (id int);


-- +migrate Down
-- DROP TABLE escrow;

DROP TABLE logical_cache_mapping;

DROP TABLE escrow;

-- DROP TABLE request;

-- DROP TABLE ticket;

-- DROP TABLE object_metadata;
