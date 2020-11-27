package clickhouse

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type clusterNode struct {
	hostName    string
	hostAddress string
	port        uint
	shardNum    uint
	shardWeight uint
	ch          *clickhouseType
}

type clusterInfo []clusterNode

//-----------------------------------------------------------

type tableIDType struct {
	db   string
	name string
}

func (table tableIDType) Db() string {
	return table.db
}

func (table tableIDType) Name() string {
	return table.name
}

func (table tableIDType) getLoadingPrefix() string {
	return fmt.Sprintf("%s_loading_", table.Name())
}

//------------------------------------------------------------------------------------

type clickhouseType struct {
	host string
	port uint
	user string
	pwd  string
	main bool
}

// TODO el receptor debería ser un puntero??

func (ch clickhouseType) printQuery(query string) {
	if !ch.main {
		log.Printf("-- @%s:%d", ch.host, ch.port)
	}

	log.Println(query)
}

func (ch clickhouseType) cmd(query string) (cmd *exec.Cmd) {
	cmd = exec.Command("docker", "exec", "-i", "clickhouse-cluster_clickhouse-ch3_1", "clickhouse-client", "-h", ch.host, "--port", strconv.Itoa(int(ch.port)), "-q", query)
	cmd.Stderr = os.Stderr
	return
}

func (ch clickhouseType) getEngine(tbl TableID) string {
	return ch.Result(fmt.Sprintf("SELECT engine_full FROM system.tables WHERE database='%s' AND name='%s' FORMAT TabSeparatedRaw", tbl.Db(), tbl.Name()))
}

func (ch clickhouseType) LoaderFor(tbl TableID) (res Loader) {
	engineFull := ch.getEngine(tbl)

	if engineFull == "" {
		log.Fatalf("-- Couldn't find table %s.%s", tbl.Db(), tbl.Name())
	}

	if strings.HasPrefix(engineFull, "Distributed(") {
		return &distStrategyType{
			server:  &ch,
			TableID: tbl,
			engine:  engineFull,
		}
	}

	return &localStrategyType{
		server:  &ch,
		TableID: tbl,
		engine:  engineFull,
	}
}

// TODO common factor between Exec() y Pipe()
func (ch clickhouseType) Exec(query string) {
	cmd := ch.cmd(query)

	ch.printQuery(query)
	if err := cmd.Run(); err != nil {
		os.Exit(-1)
	}
}

func (ch clickhouseType) Result(query string) (res string) {
	cmd := ch.cmd(query)

	ch.printQuery(query)
	resAsBytes, err := cmd.Output()
	if err != nil {
		os.Exit(-1)
	}
	res = string(resAsBytes)
	return
}

func (ch clickhouseType) Pipe(query string) {
	cmd := ch.cmd(query)

	cmd.Stdin = os.Stdin

	ch.printQuery(query)
	if err := cmd.Run(); err != nil {
		os.Exit(-1)
	}
}
