package clickhouse

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// NewTableID builds a new TableID
func NewTableID(db string, name string) TableID {
	return &tableIDType{
		db:   db,
		name: name,
	}
}

// NewClickhouse builds a new Clickhouse
func NewClickhouse(host string, port uint, user string, pwd string) Clickhouse {
	return &clickhouseType{
		host: host,
		port: port,
		user: user,
		pwd:  pwd,
		main: true,
	}
}

// -----------------------------------------------------------------------------------

func getPendingTables(server Clickhouse, tableID TableID) (res []TableID) {
	query := fmt.Sprintf("SHOW TABLES IN %s LIKE '%s%%' FORMAT TSV", tableID.Db(), tableID.getLoadingPrefix())
	pendingLoadTables := server.Result(query)

	if pendingLoadTables != "" {
		res = make([]TableID, 0, 5)
		scanner := bufio.NewScanner(strings.NewReader(pendingLoadTables))
		for scanner.Scan() {
			res = append(res, &tableIDType{
				db:   tableID.Db(),
				name: scanner.Text(),
			})
		}
	}

	return
}

func dropLoadingTables(server Clickhouse, tables []TableID) {
	dropLoadingTablesDist(server, tables, "")
}

func dropLoadingTablesDist(server Clickhouse, tables []TableID, cluster string) {
	for _, table := range tables {
		dropTableOnCluster(server, table, cluster)
	}
}

func dropTable(server Clickhouse, tableID TableID) {
	dropTableOnCluster(server, tableID, "")
}

func dropTableOnCluster(server Clickhouse, tableID TableID, cluster string) {
	if cluster != "" {
		server.Exec(fmt.Sprintf("DROP TABLE %s.%s ON CLUSTER %s", tableID.Db(), tableID.Name(), cluster))
	} else {
		server.Exec(fmt.Sprintf("DROP TABLE %s.%s", tableID.Db(), tableID.Name()))
	}
}

func processPendingTablesDist(server Clickhouse, tableID TableID, pendingTables []TableID, cluster string, cInfo clusterInfo) {
	for _, pendingTable := range pendingTables {
		movePartitionsToFinalTableDist(server, pendingTable, tableID, cluster, cInfo)
		dropTableOnCluster(server, pendingTable, cluster)
	}
}

func movePartitionsToFinalTableDist(server Clickhouse, pendingID TableID, sourceID TableID, cluster string, cInfo clusterInfo) {
	partitions := getPartitionsOnCluster(server, pendingID, cluster)

	for _, node := range cInfo {
		for _, partID := range partitions {
			node.ch.Exec(fmt.Sprintf("ALTER TABLE %s.%s ATTACH PARTITION ID '%s' FROM %s.%s", sourceID.Db(), sourceID.Name(), partID, pendingID.Db(), pendingID.Name()))
			node.ch.Exec(fmt.Sprintf("ALTER TABLE %s.%s DROP PARTITION ID '%s'", pendingID.Db(), pendingID.Name(), partID))
		}
	}
}

func processPendingTables(server Clickhouse, tableID TableID, pendingTables []TableID) {
	for _, pendingTable := range pendingTables {
		movePartitionsToFinalTable(server, tableID, pendingTable)
		dropTable(server, pendingTable)
	}
}

func getPartitionsOnCluster(server Clickhouse, tableID TableID, cluster string) (res []string) {
	query := ""
	if cluster == "" {
		query = fmt.Sprintf("SELECT distinct(partition_id) AS partition FROM system.parts WHERE database='%s' AND table='%s' AND active=1 FORMAT TSV", tableID.Db(), tableID.Name())
	} else {
		query = fmt.Sprintf("SELECT distinct(partition_id) AS partition FROM cluster('%s',system.parts) WHERE database='%s' AND table='%s' AND active=1 FORMAT TSV", cluster, tableID.Db(), tableID.Name())
	}

	partitions := server.Result(query)

	if partitions != "" {
		res = make([]string, 0, 5)
		scanner := bufio.NewScanner(strings.NewReader(partitions))
		for scanner.Scan() {
			res = append(res, scanner.Text())
		}
	}

	return
}

func getPartitions(server Clickhouse, tableID TableID) (res []string) {
	return getPartitionsOnCluster(server, tableID, "")
}

func movePartitionsToFinalTable(server Clickhouse, tableID TableID, pending TableID) {
	partitions := getPartitions(server, pending)

	for _, partID := range partitions {
		attach := fmt.Sprintf("ALTER TABLE %s.%s ATTACH PARTITION ID '%s' FROM %s.%s", tableID.Db(), tableID.Name(), partID, pending.Db(), pending.Name())
		server.Exec(attach)
		drop := fmt.Sprintf("ALTER TABLE %s.%s DROP PARTITION ID '%s'", pending.Db(), pending.Name(), partID)
		server.Exec(drop)
	}
}

func getEngineLoading(engineShard string) (engineLoading string) {
	engineLoading = engineShard

	if strings.HasPrefix(engineShard, "Replicated") {
		engineLoading = engineLoading[10:]
		start := strings.Index(engineLoading, "(")
		secondComma := strings.IndexFunc(engineLoading, func() func(r rune) bool {
			c := 0
			return func(r rune) bool {
				if r == ',' {
					c++
				}
				if c == 2 {
					return true
				}

				return false
			}
		}())
		firstClose := strings.Index(engineLoading, ")")

		end := firstClose
		if secondComma != -1 && secondComma < firstClose {
			end = secondComma + 1
		}
		engineLoading = engineLoading[:start+1] + engineLoading[end:]
	}

	return strings.TrimSpace(engineLoading)
}

func stop() {
	log.Println("-- Refuse to process because there are some pending load tables")
	os.Exit(0)
}
