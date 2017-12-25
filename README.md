# iota-mon

Usage of iota-mon:
  -i string
    	IOTA node's command endpoint (default "http://localhost:14265")
  -m string
    	file with neighbor mappings: { "<neighbor1_address>": { "name": "<neighbor1_name>", "slack": "<neighbor1_slack_username>" }, ... }
  -o string
    	StatsD daemon address (default "localhost:8125")
  -r int
    	request interval in seconds (default 5)
