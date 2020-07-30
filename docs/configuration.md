# Configuration

All daemons use [viper] for configuration, which supports either loading a
config file or overwriting individual settings with environment variables.

[viper]: https://github.com/spf13/viper

The location for the config can be set with `-config`, the defaults are:

| Daemon     | Default config location    |
| ---------- | -------------------------- |
| cached     | `./cache.config.toml`      |
| publisherd | `./publisher.config.toml`  |
| bootstrapd | `./bootstrapd.config.toml` |

Environment variables are prefixed:

| Daemon     | Prefix       |
| ---------- | ------------ |
| cached     | `CACHE_`     |
| publisherd | `PUBLISHER_` |
| bootstrapd | `BOOTSTRAP_` |

So if you want to overwrite status_addr for cached you would set `CACHE_STATUS_ADDR=`.

# cached example config
```
grpc_addr = ":9000"
http_addr = ":9443"
status_addr = ":9100"
bootstrap_addr = "bootstrapd:7777"

badger_directory = "./chunks/"
database = "cache.db"
contact_url = ""
```

# publisher example config
```
grpc_addr = ":7070"
status_addr = ":8100"
bootstrap_addr = "bootstrapd:7777"
default_cache_duration = 300

upstream = "http://localhost/"
database = "host=publisher-db port=5432 user=postgres dbname=publisher sslmode=disable"
```

# bootstrapd example config
```
grpc_addr = ":7777"
database = "./bootstrapd.db"
status_addr = ":8100"
```
