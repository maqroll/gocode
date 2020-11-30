package clickhouse

import (
	"fmt"
	"log"
	"os"
	"time"
)

type localStrategyType struct {
	server       *clickhouseType
	engine       string
	loadingTable TableID
	TableID
}

func (strategy localStrategyType) Load(format string, pending string) {
	nsec := time.Now().UnixNano()
	strategy.loadingTable = &tableIDType{
		db:   strategy.Db(),
		name: fmt.Sprintf("%s%d", strategy.getLoadingPrefix(), nsec),
	}

	pendingTables := strategy.server.getPendingTables(strategy.TableID)
	if len(pendingTables) > 0 {
		switch pending {
		case StopOption:
			stop()
		case DeleteOption:
			strategy.server.dropTables(pendingTables)
		case ProcessOption:
			strategy.processPendingTables(pendingTables)
		}
	}

	checkStdin()

	engine := getEngineLoading(strategy.engine)
	strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s AS %s.%s ENGINE=%s", strategy.loadingTable.Db(), strategy.loadingTable.Name(), strategy.Db(), strategy.Name(), engine))

	strategy.server.Pipe(fmt.Sprintf("INSERT INTO %s.%s FORMAT %s", strategy.loadingTable.Db(), strategy.loadingTable.Name(), format))

	strategy.movePartitionsToFinalTable(strategy.loadingTable)

	strategy.server.dropTable(strategy.loadingTable)
}

func (strategy localStrategyType) processPendingTables(pendingTables []TableID) {
	for _, pendingTable := range pendingTables {
		strategy.movePartitionsToFinalTable(pendingTable)
		strategy.server.dropTable(pendingTable)
	}
}

func (strategy localStrategyType) movePartitionsToFinalTable(from TableID) {
	partitions := getPartitions(strategy.server, strategy.loadingTable)

	workers := &workersType{}
	workers.start(20)

	for _, partID := range partitions {
		attach := fmt.Sprintf("ALTER TABLE %s.%s ATTACH PARTITION ID '%s' FROM %s.%s", strategy.TableID.Db(), strategy.TableID.Name(), partID, from.Db(), from.Name())
		drop := fmt.Sprintf("ALTER TABLE %s.%s DROP PARTITION ID '%s'", from.Db(), from.Name(), partID)

		workers.sendCommand(&commandType{
			node: &clusterNode{
				ch: strategy.server,
			},
			query: []string{attach, drop},
		})
	}

	close(workers.input)
	failedCommands := workers.getFailedCommands()

	if len(failedCommands) > 0 {
		for _, response := range failedCommands {
			log.Println("--")
			log.Printf("%s", response.query())
			log.Printf("%s", response.err())
		}
		os.Exit(-1)
	} else {
		log.Printf("-- Moved partitions to final table without errors")
	}
}
