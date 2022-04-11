#!/usr/bin/env bash
set -e

HOST_IP=`ip -4 addr show scope global dev docker0|grep inet |awk '{print $2}'|cut -d / -f 1`

echo "Docker HOST_IP: $HOST_IP"

# allow access from containers to specific ports on the host
#sudo ufw allow from 172.17.0.0/16 to $HOST_IP port 8080
#sudo ufw allow from 172.17.0.0/16 to $HOST_IP port 9090

# start prometheus
docker run --add-host outside:$HOST_IP \
    -d \
    --name prometheus \
    -p 9090:9090 \
    -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
    \
    prom/prometheus \
    --config.file=/etc/prometheus/prometheus.yml
echo "Started prometheus"

# start grafana
docker run --add-host outside:$HOST_IP \
    -d \
    --name grafana \
    -p 3000:3000 \
    \
    grafana/grafana
echo "Started grafana"

sleep 5
xdg-open http://localhost:3000
