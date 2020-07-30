# omnibus-cache

## Building

To build the docker image, simply run from the project root:

    docker build -t cachecash/go-cachecash .

# Running

A configuration file must be bind-mounted at `/etc/cache.config.json`.  This is temporary: once caches do not require
pregenerated configuration, this will be reworked.

```
# This assumes that you have started ElasticSearch with `docker-compose up -d` and that
# you have generated configuration files with the `generate-config` utility.

docker run -it --rm --init \
  --name test-cache \
  -v "$PWD"/cfg/cache-0.config.json:/etc/cache.config.json \
  --network elasticsearch-kibana_default \
  -e 'ELASTICSEARCH_URL=http://elasticsearch:9200/' \
  -e 'OMNIBUS=1' \
  cachecash/omnibus-cache
```
