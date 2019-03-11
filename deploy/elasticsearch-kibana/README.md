# docker-compose-elasticsearch-kibana

Based on [maxyermayank/docker-compose-elasticsearch-kibana](https://github.com/maxyermayank/docker-compose-elasticsearch-kibana/).

## Overview

`docker-compose` configuration for a 3-node Elasticsearch cluster and Kibana behind an Nginx instance, for development
and testing purposes.

We use the official Docker images for [ElasticSearch and Kibana](https://www.docker.elastic.co/).  Documentation is
available for the images for [ElasticSearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/docker.html) and for [Kibana](https://www.elastic.co/guide/en/kibana/current/docker.html).

## Quick reference

```
# Start stack in daemon mode
docker-compose up -d
# Check status of docker-compose cluster
docker-compose ps -a
# Cluster Node Info
curl http://localhost:9200/_nodes?pretty=true
# Access Kibana
http://localhost:5601
# Accessing Kibana through Nginx
http://localhost:8080
# Access Elasticsearch
http://localhost:9200
```
