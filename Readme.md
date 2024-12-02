![Kebler](https://github.com/Sharkboy-j/jbd_bms_grafana/raw/main/grafana.png)

docker-compose.yml

```
version: "2.1"
services:
  bms:
    image: bms
    container_name: bms
    environment:
      - TIMEOUT=4
      - INFLUX_DBURL=http://localhost:8086
      - INFLUX_TOKEN=
      - INFLUX_ORG=jbd
      - INFLUX_BUCKET=jbd
    #if darwin
    # - BMS_UUID=59d9d8cf-7dc9-2f43-ab65-dc2907a5fc4d
    #else
      - BMS_MAC=A5:C2:37:06:1B:C9
    read_only: false
    stop_grace_period: 30m
    network_mode: "host"
    tmpfs:
      - /tmp
    tty: true
    restart: unless-stopped
    volumes:
      - /var/run/dbus/:/var/run/dbus/```
