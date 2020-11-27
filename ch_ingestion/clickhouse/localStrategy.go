package clickhouse

import (
	"fmt"
	"time"
)

type localStrategyType struct {
	server *clickhouseType
	engine string
	TableID
}

func (strategy localStrategyType) Load(format string, pending string) {
	nsec := time.Now().UnixNano()
	loadingTable := &tableIDType{
		db:   strategy.Db(),
		name: fmt.Sprintf("%s%d", strategy.getLoadingPrefix(), nsec),
	}

	pendingTables := getPendingTables(strategy.server, strategy.TableID)
	if len(pendingTables) > 0 {
		switch pending {
		case StopOption:
			stop()
		case DeleteOption:
			dropLoadingTables(strategy.server, pendingTables)
		case ProcessOption:
			processPendingTables(strategy.server, strategy.TableID, pendingTables)
		}
	}

	checkStdin()

	engine := getEngineLoading(strategy.engine)
	strategy.server.Exec(fmt.Sprintf("CREATE TABLE %s.%s AS %s.%s ENGINE=%s", loadingTable.Db(), loadingTable.Name(), strategy.Db(), strategy.Name(), engine))

	strategy.server.Pipe(fmt.Sprintf("INSERT INTO %s.%s FORMAT %s", loadingTable.Db(), loadingTable.Name(), format))

	movePartitionsToFinalTable(strategy.server, strategy.TableID, loadingTable)

	dropTable(strategy.server, loadingTable)
}
