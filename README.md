# iota-mon

## Building sample

GOOS=linux GOARCH=amd64 go build -o iota-mon -a -v main.go

## Usage

option | type | description | default value | additional description
------ | ---- | ----------- | ------------- | ----------------------
-i | string | IOTA node's command endpoint | `http://localhost:14265`
-m | string | file with neighbor mappings || `{ "<neighbor1_address>": { "name": "<neighbor1_name>", "slack": "<neighbor1_slack_username>" }, ... }`
-o | string | StatsD daemon address (daemon must have dog tags support) | `localhost:8125`
-r | int | request interval in seconds | `5`
-h | `n/a` | prints all options
