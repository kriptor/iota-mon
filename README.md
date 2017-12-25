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
- put Dockerfile and docker-compose.yaml into some directory
- replace `IOTA_IRI_IP_LOOPBACK_WILL_NOT_WORK` (inside docker-compose.yaml) with the IP where your IOTA IRI command interface is listening
- create directory paths (mkdir -p) `/iota-mon/config/grafana`, `/iota-mon/data/influxdb` and `/iota-mon/data/grafana`
- copy 
