# omnibus-cache

## Building

This image must be built from the root of the repository by running

```
docker build -f deploy/omnibus-cache/Dockerfile .
```

# Running

A configuration file must be bind-mounted at `/etc/cache.config.json`.  This is temporary: once caches do not require
pregenerated configuration, this will be reworked.
