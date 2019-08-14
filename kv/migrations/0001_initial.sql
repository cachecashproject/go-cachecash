-- +migrate Up

CREATE TABLE if not exists kvstore (
  member varchar not null,
  key varchar not null,
  value bytea not null,

  PRIMARY KEY (member, key)
);

-- +migrate Down

DROP TABLE kvstore;
