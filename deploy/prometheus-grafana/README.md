# docker-compose: prometheus-grafana

Based on [alexellis/quickstart-prometheus](https://github.com/alexellis/quickstart-prometheus).

A more complete example setup that we might want to crib things from is available at https://github.com/vegasbrianc/prometheus.

## Overview

`docker-compose` configuration for Prometheus and Grafana instances, for development and testing purposes.  Includes
`node-exporter` for monitoring the Docker containers as well.

## Quick reference

```
# Start stack in daemon mode
docker-compose up -d
# Check status of docker-compose cluster
docker-compose ps -a
```

Prometheus is available at http://localhost:9090/.  Grafana is available at http://localhost:3000/.
