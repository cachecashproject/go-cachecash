# omnibus-cache

## Building

This image must be built from the root of the repository by running

```
docker build -f deploy/omnibus-cache/Dockerfile -t cachecash/omnibus-cache .
```

It depends on the `cachecash/go-cachecash` image for binaries, so you may wish to rebuild that image first.

# Running

A configuration file must be bind-mounted at `/etc/cache.config.json`.  This is temporary: once caches do not require
pregenerated configuration, this will be reworked.

```
# This assumes that you have started ElasticSearch with `docker-compose up -d` and that
# you have generated configuration files with the `generate-config` utility.

docker run -it --rm \
  --name test-cache \
  -v "$PWD"/cfg/cache-0.config.json:/etc/cache.config.json \
  --network elasticsearch-kibana_default \
  -e 'ELASTICSEARCH_URL=http://elasticsearch:9200/' \
  cachecash/omnibus-cache
```
