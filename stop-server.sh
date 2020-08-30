#!/usr/bin/env bash
set -eufo pipefail

docker rm -f grafana || true
docker rm -f timescaledb || true
