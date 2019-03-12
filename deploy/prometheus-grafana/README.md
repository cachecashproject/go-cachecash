# docker-compose: prometheus-grafana

Based on [vegasbrianc/prometheus](https://github.com/vegasbrianc/prometheus/).

## Overview

`docker-compose` configuration for Prometheus and Grafana instances, for development and testing purposes.  Includes
`node-exporter` for monitoring the Docker containers as well.

## Volume creation

Docker will create new bind-mounted directories as root, but since the containers do not run as root, you will encounter
permissions errors.  You can resolve this by chowning the data directories, or by commenting out the bind-mount
configuration lines in `docker-compose.yml`.

```
mkdir -p data/grafana && sudo chown -R 472:472 data/grafana
mkdir -p data/prometheus && sudo chown -R 65534:65534 data/prometheus
```

## Quick reference

```
# Start stack in daemon mode
docker-compose up -d
# Check status of docker-compose cluster
docker-compose ps
```

Prometheus is available at http://localhost:9090/.  Grafana is available at http://localhost:3000/.
