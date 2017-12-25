# iota-mon

Usage of iota-mon:

option | type | description | default value | additional description
------ | ---- | ----------- | ------------- | ----------------------
-i | string | IOTA node's command endpoint | `http://localhost:14265`
-m | string | file with neighbor mappings || `{ "<neighbor1_address>": { "name": "<neighbor1_name>", "slack": "<neighbor1_slack_username>" }, ... }`
-o | string | StatsD daemon address (daemon must have dog tags support) | `localhost:8125`
-r | int | request interval in seconds | `5`
