package clickhouse

// TableID identifies a table in clickhouse server
type TableID interface {
	Db() string
	Name() string
	getLoadingPrefix() string
}

// Loader allows to load tables on clickhouse servers
type Loader interface {
	Load(format string, pending string)
}

// Clickhouse allows to execute operations on clickhouse servers
type Clickhouse interface {
	LoaderFor(tbl TableID) Loader
	Exec(cmd string)
	Result(cmd string) string
	Pipe(cmd string)
}

type command interface {
	exec() response
}

type response interface {
	node() string
	query() string
	err() string
}

type workers interface {
	sendCommand(c command)
	getFailedCommands() []response
}
