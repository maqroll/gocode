package main

import (
	"flag"
	"log"

	ch "./clickhouse"
)

var (
	host    = flag.String("host", "localhost", "clickhouse server host")
	port    = flag.Int("port", 9000, "clickhouse server port")
	table   = flag.String("tbl", "", "table to ingest")
	db      = flag.String("db", "default", "database")
	pending = flag.String("pending", ch.StopOption, "action for pending tables: "+ch.StopOption+","+ch.ProcessOption+","+ch.DeleteOption)
	format  = flag.String("format", "CSV", "input format, any accepted by clickhouse")
	user    = flag.String("user", "default", "clickhouse user")
	pwd     = flag.String("pwd", "", "clickhouse pwd")
)

func init() {
	log.SetFlags(0)
}

func main() {

	flag.Parse()

	checkParams()

	server := ch.NewClickhouse(*host, uint(*port), *user, *pwd)
	tableID := ch.NewTableID(*db, *table)

	loadStrategy := server.LoaderFor(tableID)

	loadStrategy.Load(*format, *pending)
}

func checkParams() {
	if *table == "" {
		log.Fatalln("tbl flag is compulsory")
	}

	if *pending != ch.StopOption && *pending != ch.ProcessOption && *pending != ch.DeleteOption {
		log.Fatalf("pending flag value is invalid. Should be (%s|%s|%s)", ch.ProcessOption, ch.DeleteOption, ch.ProcessOption)
	}
}
