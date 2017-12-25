package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strconv"

	"strings"

	"encoding/json"
	"io/ioutil"

	"github.com/kriptor/giota"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alexcesaro/statsd.v2"
)

type NeighborMappedData struct {
	Name  string `json:"name"`
	Slack string `json:"slack"`
}

func (a *NeighborMappedData) cleanedUpName() string {
	n := tagValueCleanupReplacer.Replace(strings.Trim(a.Name, " \t\n\r"))
	if len(n) > 0 {
		return n
	}
	return "n/a"
}

func (a *NeighborMappedData) cleanedUpSlack() string {
	s := tagValueCleanupReplacer.Replace(strings.Trim(a.Slack, " \t\n\r@"))
	if len(s) > 0 {
		return s
	}
	return "n/a"
}

type NeighborMappedDataMap map[string]NeighborMappedData

const TIME_FORMAT = "2006-01-02 15:04:05.000000"

var tagValueCleanupReplacer = strings.NewReplacer(":", "_", "|", "[pipe]", "@", "[at]")

var iotaNodeCmdEndpoint, statsDAddress, mappingsFile string
var requestInterval int64

var neighborMappedDataMap NeighborMappedDataMap

func init() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = TIME_FORMAT
	log.SetFormatter(customFormatter)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the debug severity or above.
	log.SetLevel(log.DebugLevel)

	flag.StringVar(&iotaNodeCmdEndpoint, "i", "http://localhost:14265", "IOTA node's command endpoint")
	flag.StringVar(&statsDAddress, "o", "localhost:8125", "StatsD daemon address")
	flag.Int64Var(&requestInterval, "r", 5, "request interval in seconds")
	flag.StringVar(&mappingsFile, "m", "", "file with neighbor mappings: { \"<neighbor1_address>\": { \"name\": \"<neighbor1_name>\", \"slack\": \"<neighbor1_slack_username>\" }, ... }")
	flag.Parse()

	mappingsFile = strings.Trim(mappingsFile, " \n\t\r")
	if len(mappingsFile) > 0 {
		if data, err := ioutil.ReadFile(mappingsFile); err != nil {
			log.WithError(err).Fatal("cannot read file: ", mappingsFile)
		} else {
			if err = json.Unmarshal(data, &neighborMappedDataMap); err != nil {
				log.WithError(err).Fatal("cannot unmarshal data from file: ", mappingsFile)
			}
		}
	}

	log.WithField("iota_endpoint", iotaNodeCmdEndpoint).
		WithField("statsd_address", statsDAddress).
		WithField("request_interval", requestInterval).
		WithField("neighbor_mappings", len(neighborMappedDataMap)).Info("program arguments used")
}

func main() {
	var gracefulStop = make(chan os.Signal)
	defer close(gracefulStop)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		log.WithFields(log.Fields{"sig": sig, "exitInSec": 2, "exitCode": 0}).Info("signal caught")
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	options := []statsd.Option{statsd.Address(statsDAddress),
		statsd.Prefix("node."),
		statsd.Network("udp"),
		statsd.ErrorHandler(func(err error) {
			log.WithError(err).Error("StatsD error")
		}),
		statsd.TagsFormat(statsd.Datadog),
		statsd.Tags("app_ep", tagValueCleanupReplacer.Replace(iotaNodeCmdEndpoint),
			"statsd_addr", tagValueCleanupReplacer.Replace(statsDAddress),
			"request_interval", strconv.FormatInt(requestInterval, 10))}

	statsDClientFailSleep := 0
	statsDClient, err := statsd.New(options...)
	for err != nil {
		if statsDClientFailSleep < 60 {
			statsDClientFailSleep += 5
		}
		log.WithError(err).Error("failed to create StatsD client with options: ", options, "- sleeping for ", statsDClientFailSleep, "seconds ...")
		time.Sleep(time.Second * time.Duration(statsDClientFailSleep))
	}
	defer statsDClient.Close()

	api := giota.NewAPI(iotaNodeCmdEndpoint, nil)

	repeatTicker := time.NewTicker(time.Second * time.Duration(requestInterval))
	defer repeatTicker.Stop()
	go checkIotaNode(repeatTicker, api, statsDClient)

	//sleep forever
	select {}
}

func checkIotaNode(ticker *time.Ticker, api *giota.API, defaultStatsDClient *statsd.Client) {
	var prevNeighborsMap = make(map[string]giota.Neighbor)
	var statsDClientsMap = make(map[string]*statsd.Client)
	iterationsToSkip := 1

	for range ticker.C {
		log.Info("calling GetNodeInfo ...")
		if resp, err := api.GetNodeInfo(); err != nil {
			log.WithError(err).Error("failed to receive node status from iota node from:", iotaNodeCmdEndpoint)
			c := statsDClientsMap["nodeinfo"]
			if c == nil {
				c = defaultStatsDClient
			}
			c.Gauge("heartbeat", 0)
			continue
		} else {
			log.WithField("appName", resp.AppName).
				WithField("appVer", resp.AppVersion).
				WithField("jreCpu", resp.JREAvailableProcessors).
				WithField("jreFreeMem", resp.JREFreeMemory).
				WithField("jreVer", resp.JREVersion).
				WithField("jreMaxMem", resp.JREMaxMemory).
				WithField("jreTotalMem", resp.JRETotalMemory).
				WithField("latestMil", resp.LatestMilestone).
				WithField("latestMilIdx", resp.LatestMilestoneIndex).
				WithField("latestMilSolid", resp.LatestSolidSubtangleMilestone).
				WithField("latestMilSolidIdx", resp.LatestSolidSubtangleMilestoneIndex).
				WithField("neighbors", resp.Neighbors).
				WithField("packetQSize", resp.PacketQueueSize).
				WithField("time", resp.Time).
				WithField("tips", resp.Tips).
				WithField("txsToReq", resp.TransactionsToRequest).
				WithField("duration", resp.Duration).
				Info("node data")

			c := statsDClientsMap["nodeinfo"]
			if c == nil {
				c = defaultStatsDClient.Clone(statsd.Tags("app_name", resp.AppName,
					"app_ver", resp.AppVersion,
					"jre_cpu_cnt", strconv.FormatInt(resp.JREAvailableProcessors, 10),
					"jre_ver", resp.JREVersion))
				statsDClientsMap["nodeinfo"] = c
			}

			// statsd
			c.Gauge("heartbeat", 1)
			switch resp.LatestSolidSubtangleMilestoneIndex == resp.LatestMilestoneIndex {
			case true:
				c.Gauge("synced", 1)
			default:
				c.Gauge("synced", 0)
			}

			c.Gauge("latest_milestone_idx", resp.LatestMilestoneIndex)
			c.Gauge("latest_solid_subtangle_milestone_idx", resp.LatestSolidSubtangleMilestoneIndex)
			c.Gauge("tips", resp.Tips)
			c.Gauge("txs_to_request", resp.TransactionsToRequest)
			c.Gauge("packet_queue_size", resp.PacketQueueSize)
			c.Gauge("neighbors", resp.Neighbors)
			c.Gauge("jre_free_mem", resp.JREFreeMemory)
			c.Gauge("jre_max_mem", resp.JREMaxMemory)
			c.Gauge("jre_total_mem", resp.JRETotalMemory)
		}

		log.Info("calling GetNeighbors ...")
		if resp, err := api.GetNeighbors(); err != nil {
			log.WithError(err).Error("failed to receive neighbors stats from iota node:", iotaNodeCmdEndpoint)
			c := statsDClientsMap["nodeinfo"]
			if c == nil {
				c = defaultStatsDClient
			}
			c.Gauge("heartbeat", 0)
		} else {
			for _, n := range resp.Neighbors {
				log.WithField("address", n.Address).
					WithField("connType", n.ConnectionType).
					WithField("txsAll", n.NumberOfAllTransactions).
					WithField("txsNew", n.NumberOfNewTransactions).
					WithField("txsInvalid", n.NumberOfInvalidTransactions).
					WithField("txsSent", n.NumberOfSentTransactions).
					WithField("txRndReqs", n.NumberOfRandomTransactionRequests).
					Info("neighbour data")

				neighborAddr := string(n.Address)
				neighborHostAndPort := strings.Split(neighborAddr, ":")
				neighborHost := neighborHostAndPort[0]
				neighborPort := neighborHostAndPort[1]
				neighborKey := n.ConnectionType + "://" + neighborAddr
				prevNeighbor := prevNeighborsMap[neighborKey]

				c := statsDClientsMap[neighborKey]
				if c == nil {
					nmd := neighborMappedDataMap[neighborKey]
					c = defaultStatsDClient.Clone(statsd.Tags("neighbor_address", tagValueCleanupReplacer.Replace(neighborKey),
						"neighbor_host", neighborHost,
						"neighbor_port", neighborPort,
						"neighbor_conn_type", n.ConnectionType,
						"neighbor_name", nmd.cleanedUpName(),
						"neighbor_slack", nmd.cleanedUpSlack()))
					statsDClientsMap[neighborKey] = c
				}

				// statsd
				if iterationsToSkip <= 0 {
					c.Count("neighbor_txs_all", n.NumberOfAllTransactions-prevNeighbor.NumberOfAllTransactions)
					c.Count("neighbor_txs_new", n.NumberOfNewTransactions-prevNeighbor.NumberOfNewTransactions)
					c.Count("neighbor_txs_invalid", n.NumberOfInvalidTransactions-prevNeighbor.NumberOfInvalidTransactions)
					c.Count("neighbor_txs_sent", n.NumberOfSentTransactions-prevNeighbor.NumberOfSentTransactions)
					c.Count("neighbor_tx_rnd_reqs", n.NumberOfRandomTransactionRequests-prevNeighbor.NumberOfRandomTransactionRequests)
				}

				prevNeighborsMap[neighborKey] = n
			}
			iterationsToSkip--
		}
	}
}
