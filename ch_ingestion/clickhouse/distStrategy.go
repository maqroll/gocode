package clickhouse

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type distStrategyType struct {
	server       *clickhouseType
	engine       string
	cluster      string
	shard        TableID
	shardingKey  string
	clusterNodes clusterInfo
	TableID
}

func (strategy *distStrategyType) Load(format string, pending string) {
	strategy.parseDistributedParams()
	strategy.populateClusterInfo()

	nsec := time.Now().UnixNano()

	distLoadingTable := &tableIDType{
		db:   strategy.Db(),
		name: fmt.Sprintf("%s%d", strategy.getLoadingPrefix(), nsec),
	}

	shardLoadingTable := &tableIDType{
		db:   strategy.shard.Db(),
		name: fmt.Sprintf("%s%d", strategy.shard.getLoadingPrefix(), nsec),
	}

	pendingLoadShardTables := getPendingTables(strategy.server, strategy.shard)
	pendingLoadDistTables := getPendingTables(strategy.server, strategy)

	if len(pendingLoadDistTables) > 0 || len(pendingLoadShardTables) > 0 {
		switch pending {
		case StopOption:
			stop()
		case DeleteOption:
			dropLoadingTablesDist(strategy.server, pendingLoadShardTables, strategy.cluster) // first remove shards
			dropLoadingTablesDist(strategy.server, pendingLoadDistTables, strategy.cluster)  // later remove distributed
		case ProcessOption:
			processPendingTablesDist(strategy.server, strategy.shard, pendingLoadShardTables, strategy.cluster, strategy.clusterNodes)
			dropLoadingTablesDist(strategy.server, pendingLoadDistTables, strategy.cluster) // distributed don't require processing, just remove
		}
	}

	checkStdin()

	engineShard := strategy.server.getEngine(strategy.shard)
	engineShard = getEngineLoading(string(engineShard))
	strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s ON CLUSTER %s AS %s.%s ENGINE=%s", shardLoadingTable.Db(), shardLoadingTable.Name(), strategy.cluster, strategy.shard.Db(), strategy.shard.Name(), engineShard))

	if strategy.shardingKey != "" {
		strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s ON CLUSTER %s AS %s.%s ENGINE=Distributed(%s,%s,%s,%s)",
			distLoadingTable.Db(),
			distLoadingTable.Name(),
			strategy.cluster,
			strategy.Db(),
			strategy.Name(),
			strategy.cluster,
			shardLoadingTable.Db(),
			shardLoadingTable.Name(),
			strategy.shardingKey))
	} else {
		strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s ON CLUSTER %s AS %s.%s ENGINE=Distributed(%s,%s,%s)",
			distLoadingTable.Db(),
			distLoadingTable.Name(),
			strategy.cluster,
			strategy.Db(),
			strategy.Name(),
			strategy.cluster,
			shardLoadingTable.Db(),
			shardLoadingTable.Name()))
	}

	strategy.server.Pipe(fmt.Sprintf("INSERT INTO %s.%s FORMAT %s", distLoadingTable.Db(), distLoadingTable.Name(), format))

	// TODO execute this in parallel?? harder to visualize in console
	movePartitionsToFinalTableDist(strategy.server, shardLoadingTable, strategy.shard, strategy.cluster, strategy.clusterNodes)

	dropLoadingTablesDist(strategy.server, []TableID{distLoadingTable}, strategy.cluster)
	dropLoadingTablesDist(strategy.server, []TableID{shardLoadingTable}, strategy.cluster)
}

func (strategy *distStrategyType) parseDistributedParams() {
	matches := distRegexp.FindStringSubmatch(strategy.engine)

	if matches == nil {
		log.Fatalf("-- engine exp %q didn't look like a distributed table", strategy.engine)
	}

	if len(matches) != 2 {
		log.Fatalf("-- distributed table regexp groups didn't match for %q", strategy.engine)
	}

	// Assume no storage policy and sharding without ,
	params := strings.Split(matches[1], ",")

	strategy.cluster = strings.TrimFunc(params[0], trim)

	strategy.shard = &tableIDType{
		db:   strings.TrimFunc(params[1], trim),
		name: strings.TrimFunc(params[2], trim),
	}

	if len(params) > 3 {
		strategy.shardingKey = strings.TrimFunc(params[3], trim)
	}
}

func (strategy *distStrategyType) populateClusterInfo() {
	query := fmt.Sprintf("SELECT cluster, shard_num, shard_weight, host_name, host_address, port FROM system.clusters WHERE cluster = '%s' FORMAT CSV", strategy.cluster)
	clusterInfo := strategy.server.Result(query)

	strategy.clusterNodes = make([]clusterNode, 0, 5)

	r := csv.NewReader(strings.NewReader(string(clusterInfo)))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}

		port, _ := strconv.Atoi(record[5])
		shardNum, _ := strconv.Atoi(record[1])
		shardWeight, _ := strconv.Atoi(record[2])

		node := clusterNode{
			hostName:    record[3],
			hostAddress: record[4],
			port:        uint(port),
			shardNum:    uint(shardNum),
			shardWeight: uint(shardWeight),
			ch: &clickhouseType{
				host: record[3],
				port: uint(port),
				user: strategy.server.user,
				pwd:  strategy.server.pwd,
				main: false,
			},
		}
		strategy.clusterNodes = append(strategy.clusterNodes, node)
	}
}
