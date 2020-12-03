package clickhouse

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// NewTableID builds a new TableID
func NewTableID(db string, name string) TableID {
	return &tableIDType{
		db:   db,
		name: name,
	}
}

func newClickhouseType(host string, port uint, user string, pwd string, cli string) *clickhouseType {
	return &clickhouseType{
		host: host,
		port: port,
		user: user,
		pwd:  pwd,
		main: true,
		cli:  cli,
	}
}

// NewClickhouse builds a new Clickhouse
func NewClickhouse(host string, port uint, user string, pwd string, cli string) Clickhouse {
	return newClickhouseType(host, port, user, pwd, cli)
}

// -----------------------------------------------------------------------------------

func expandCluster(cluster clusterInfo) string {
	clusterExpanded := &strings.Builder{}
	for i, node := range cluster {
		if i != 0 {
			clusterExpanded.WriteString(",")
		}
		clusterExpanded.WriteString(node.hostName)
		clusterExpanded.WriteString(":")
		clusterExpanded.WriteString(strconv.Itoa(int(node.port)))
	}

	return clusterExpanded.String()
}

func getPartitionsOnCluster(server *clickhouseType, tableID TableID, cluster clusterInfo) (res []string) {
	query := ""
	if cluster == nil {
		query = fmt.Sprintf("SELECT distinct(partition_id) AS partition FROM system.parts WHERE database='%s' AND table='%s' AND active=1 FORMAT TSV", tableID.Db(), tableID.Name())
	} else {
		clusterExpanded := expandCluster(cluster)
		query = fmt.Sprintf("SELECT distinct(partition_id) AS partition FROM remote('%s',system.parts,'%s','%s') WHERE database='%s' AND table='%s' AND active=1 FORMAT TSV", clusterExpanded, server.user, server.pwd, tableID.Db(), tableID.Name())
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

func getPartitions(server *clickhouseType, tableID TableID) (res []string) {
	return getPartitionsOnCluster(server, tableID, nil)
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
