
# Build

    go build

# Cross compile

    GOOS=windows GOARCH=amd64 go build

# Docu
* https://svn.fhem.de/trac/browser/trunk/fhem/contrib/70_NT5000.pm
* https://wiki.fhem.de/wiki/NT5000
* https://medium.com/aeturnuminc/configure-prometheus-and-grafana-in-dockers-ff2a2b51aa1d
* https://prometheus.io/docs/guides/go-application/
* https://biancatamayo.me/blog/docker-add-host-ip/

# Features
* commands: getdata, settime
* prometheus interface

# Protocol
## Read online data

Send: "\x00\x01\x02\x01\x04". Last byte is checksum, 5 bytes in total
Receive: 13 bytes in buffer

1. UDC (voltage DC): buffer[0]*2.8+100, unit: V
2. IDC (current DC): buffer[1]*0.08, unit: A
3. UAC (voltage AC): buffer[2]+100.0, unit: V
4. IAC (current AC): buffer[3]*0.120, unit: A
5. Temperature: buffer[4]-40.0, unit: °C
6. PDC (Power DC): ($udc*$idc)/1000, unit: kW
7. PAC (Power AC): ($uac*$iac)/1000, unit: kW
8. Energy Today: (buffer[6] * 256 + buffer[7])/1000, unit: kWh
9. Energy Total: buffer[8] * 256 + buffer[9], unit: kWh
10. Heat flux: buffer[5]*6.0, unit: W/m^2

## Read time

Send: "\x00\x01\x06\x01\x08". Last byte is checksum, 5 bytes in total
Receive: 13 bytes

1. year
2. month
3. day
4. hour
5. minute

Remaining 7 bytes are zero, 13th (last) byte is checksum.

## Set time

Multiple commands:
1. Set year: "\x00\x01\x50"
2. Set month: "\x00\x01\x51"
3. Set day: "\x00\x01\x52"
4. Set hour: "\x00\x01\x53"
5. Set minute: "\x00\x01\x54"

4th byte is the actual value, 5th byte is checksum

No response.

## Read Serial Number

Send: "\x00\x01\x08\x01\x0A"

Response: 12 bytes + checksum

## Read Protocol and Firmware Version

Send: "\x00\x01\x09\x01\x0B"

Response: 6 bytes + 6 null bytes + checksum

# Prometheus and Grafana

```
HOST_IP=`ip -4 addr show scope global dev docker0|grep inet |awk '{print $2}'|cut -d / -f 1`
sudo ufw allow from 172.17.0.0/16 to 172.17.0.1 port 8080
sudo ufw allow from 172.17.0.0/16 to 172.17.0.1 port 9090

docker run --add-host outside:$HOST_IP -d --name prometheus -p 9090:9090 -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus --config.file=/etc/prometheus/prometheus.yml

docker run -d --name grafana -p 3000:3000 grafana/grafana

```

Configure grafana: http://localhost:3000
* default username "admin" and password "admin"
* add data source, prometheus
* url: http://172.17.0.1:9090 (depending on HOST_IP)
* import dashboard `NT5000-grafana-dashboard.json`

Shutdown:

```
docker stop grafana && docker rm grafana
docker stop prometheus && docker rm prometheus
```

