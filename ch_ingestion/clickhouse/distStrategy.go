package clickhouse

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type distStrategyType struct {
	server            *clickhouseType
	engine            string
	cluster           string
	shard             TableID
	shardingKey       string
	clusterNodes      clusterInfo
	distLoadingTable  TableID
	shardLoadingTable TableID
	TableID
}

func (strategy *distStrategyType) Load(format string, pending string) {
	strategy.parseDistributedParams()
	strategy.populateClusterInfo()

	nsec := time.Now().UnixNano()

	strategy.distLoadingTable = &tableIDType{
		db:   strategy.Db(),
		name: fmt.Sprintf("%s%d", strategy.getLoadingPrefix(), nsec),
	}

	strategy.shardLoadingTable = &tableIDType{
		db:   strategy.shard.Db(),
		name: fmt.Sprintf("%s%d", strategy.shard.getLoadingPrefix(), nsec),
	}

	pendingLoadShardTables := strategy.server.getPendingTables(strategy.shard)
	pendingLoadDistTables := strategy.server.getPendingTables(strategy)

	if len(pendingLoadDistTables) > 0 || len(pendingLoadShardTables) > 0 {
		switch pending {
		case StopOption:
			stop()
		case DeleteOption:
			strategy.server.dropTablesDist(pendingLoadShardTables, strategy.cluster) // first remove shards
			strategy.server.dropTablesDist(pendingLoadDistTables, strategy.cluster)  // later remove distributed
		case ProcessOption:
			strategy.processPendingTablesDist(pendingLoadShardTables)
			strategy.server.dropTablesDist(pendingLoadDistTables, strategy.cluster) // distributed don't require processing, just remove
		}
	}

	checkStdin()

	engineShard := strategy.server.getEngine(strategy.shard)
	engineShard = getEngineLoading(string(engineShard))
	strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s ON CLUSTER %s AS %s.%s ENGINE=%s", strategy.shardLoadingTable.Db(), strategy.shardLoadingTable.Name(), strategy.cluster, strategy.shard.Db(), strategy.shard.Name(), engineShard))

	if strategy.shardingKey != "" {
		strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s ON CLUSTER %s AS %s.%s ENGINE=Distributed(%s,%s,%s,%s)",
			strategy.distLoadingTable.Db(),
			strategy.distLoadingTable.Name(),
			strategy.cluster,
			strategy.Db(),
			strategy.Name(),
			strategy.cluster,
			strategy.shardLoadingTable.Db(),
			strategy.shardLoadingTable.Name(),
			strategy.shardingKey))
	} else {
		strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s ON CLUSTER %s AS %s.%s ENGINE=Distributed(%s,%s,%s)",
			strategy.distLoadingTable.Db(),
			strategy.distLoadingTable.Name(),
			strategy.cluster,
			strategy.Db(),
			strategy.Name(),
			strategy.cluster,
			strategy.shardLoadingTable.Db(),
			strategy.shardLoadingTable.Name()))
	}

	strategy.server.Pipe(fmt.Sprintf("INSERT INTO %s.%s FORMAT %s", strategy.distLoadingTable.Db(), strategy.distLoadingTable.Name(), format))
	strategy.server.Exec(fmt.Sprintf("SYSTEM FLUSH DISTRIBUTED %s.%s", strategy.distLoadingTable.Db(), strategy.distLoadingTable.Name()))

	strategy.movePartitionsToFinalTableDist(strategy.shardLoadingTable)

	strategy.server.dropTablesDist([]TableID{strategy.distLoadingTable}, strategy.cluster)
	strategy.server.dropTablesDist([]TableID{strategy.shardLoadingTable}, strategy.cluster)
}

func (strategy *distStrategyType) processPendingTablesDist(pendingTables []TableID) {
	for _, pendingTable := range pendingTables {
		strategy.movePartitionsToFinalTableDist(pendingTable)
		strategy.server.dropTableOnCluster(pendingTable, strategy.cluster)
	}
}

func (strategy *distStrategyType) movePartitionsToFinalTableDist(pendingID TableID) {
	partitions := getPartitionsOnCluster(strategy.server, pendingID, strategy.clusterNodes)

	workers := &workersType{}
	workers.start(30)

	for _, node := range strategy.clusterNodes {
		for _, partID := range partitions {
			attachQuery := fmt.Sprintf("ALTER TABLE %s.%s ATTACH PARTITION ID '%s' FROM %s.%s", strategy.shard.Db(), strategy.shard.Name(), partID, pendingID.Db(), pendingID.Name())
			dropQuery := fmt.Sprintf("ALTER TABLE %s.%s DROP PARTITION ID '%s'", pendingID.Db(), pendingID.Name(), partID)

			workers.sendCommand(&commandType{
				node:  node,
				query: []string{attachQuery, dropQuery},
			})
		}
	}

	close(workers.input)
	failedCommands := workers.getFailedCommands()

	if len(failedCommands) > 0 {
		for _, response := range failedCommands {
			log.Printf("--@ %s", response.node())
			log.Printf("%s", response.query())
			log.Printf("%s", response.err())
		}
		os.Exit(-1)
	} else {
		log.Printf("-- Moved partitions to final table on cluster without errors")
	}
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

	strategy.clusterNodes = make([]*clusterNode, 0, 5)

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

		node := &clusterNode{
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
				cli:  strategy.server.cli,
			},
		}
		strategy.clusterNodes = append(strategy.clusterNodes, node)
	}
}
