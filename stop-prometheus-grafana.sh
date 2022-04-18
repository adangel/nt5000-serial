#!/usr/bin/env bash

docker stop grafana && docker rm grafana
docker stop prometheus && docker rm prometheus