#!/usr/bin/env bash
set -eufo pipefail

docker network create 9eyes || true

./stop-server.sh
mkdir -p data/grafana
sudo chown -R 472:472 data/grafana
docker run -d \
    --name=grafana \
    --net=9eyes \
	-p 0.0.0.0:3000:3000 \
	-v $(pwd)/data/grafana:/var/lib/grafana \
	grafana/grafana:7.1.5

mkdir -p data/timescaledb/
sudo chown -R 0:0 data/timescaledb
docker run -d \
    --name timescaledb \
    --net=9eyes \
    -p 0.0.0.0:5432:5432 \
    -e POSTGRES_PASSWORD=password \
    -e PGDATA=/var/lib/postgresql/data/pgdata \
    -v $(pwd)/data/timescaledb:/var/lib/postgresql/data \
    timescale/timescaledb:latest-pg12
