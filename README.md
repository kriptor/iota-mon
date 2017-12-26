# iota-mon

## Building sample

GOOS=linux GOARCH=amd64 go build -o iota-mon -a -v main.go

## Usage

program option | type | description | default value | additional description
-------------- | ---- | ----------- | ------------- | ----------------------
-i | string | IOTA node's command endpoint | `http://localhost:14265`
-m | string | file with neighbor mappings || `{ "<neighbor1_address>": { "name": "<neighbor1_name>", "slack": "<neighbor1_slack_username>" }, ... }`
-o | string | StatsD daemon address (daemon must have dog tags support) | `localhost:8125`
-r | int | request interval in seconds | `5`
-h | `n/a` | prints all options

## Deploying with InfluxDB, Grafana and Telegraf using Docker Compose

Prerequisite is that your IOTA IRI command interface (usually http://127.0.0.1:14265) is bound to some IP other than loopback (127.0.0.1), then:  
- install the latest Docker engine (1.13.0+) 
- put `Dockerfile` and `docker-compose.yaml` (from `Docker` directory) into some directory
- replace `IOTA_IRI_IP_LOOPBACK_WILL_NOT_WORK` (inside `docker-compose.yaml`) with the IP where your IOTA IRI command interface is listening
- create directory paths (mkdir -p) `/iota-mon/config/grafana`, `/iota-mon/data/influxdb` and `/iota-mon/data/grafana`
- put `influxdb.conf`, `telegraf.conf` and `iota-mon_neighbors_map.json` (from `sample-config` directory) into `/iota-mon/config` you just created
- edit `iota-mon_neighbors_map.json` to map your neighbors
- cd to the directory you put `Dockerfile` and `docker-compose.yaml` in
- to start execute `docker-compose -f docker-compose.yaml up --build --no-start && docker-compose -f docker-compose.yaml start`
- to stop execute `docker-compose -f docker-compose.yaml stop`

## Set-up Grafana
- Grafana listens on exposed port 3000
- default username is `admin` and default password is `changeme`
- create InfluxDB datasource with url `http://influxdb:8086` and database `IOTA`
- create a dashboard, configure alerting, ...
