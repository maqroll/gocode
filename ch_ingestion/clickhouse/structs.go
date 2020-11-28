package clickhouse

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
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

type clusterInfo []*clusterNode

//-----------------------------------------------------------

type responseType struct {
	nodeF  string
	queryF string
	errF   string
}

func (r responseType) node() string {
	return r.nodeF
}

func (r responseType) query() string {
	return r.queryF
}

func (r responseType) err() string {
	return r.errF
}

//-----------------------------------------------------------

type commandType struct {
	node  *clusterNode
	query []string
}

func (command commandType) exec() (res response) {
	for _, query := range command.query {
		stderr := command.node.ch.captureErr(query)
		if stderr != "" {
			return &responseType{
				nodeF:  fmt.Sprintf("%s:%d", command.node.hostName, command.node.port),
				queryF: query,
				errF:   stderr,
			}
		}
	}
	return nil
}

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

//----------------------------------------------------------------------------------

// si queremos evitar acceso a los campos de workers desde fuera de workers
// ¿tengo que crear un nuevo paquete? ¿definir una interfaz?
type workersType struct {
	input   chan command
	output  []chan response
	failed  []response
	waiting chan struct{}
}

func (w *workersType) start(n uint) {
	w.input = make(chan command, n)
	w.output = make([]chan response, 0, n)
	w.waiting = make(chan struct{})

	for i := 0; i < int(n); i++ {
		input := w.input
		outputChannel := make(chan response, 10)
		w.output = append(w.output, outputChannel)

		go func() {
			for {
				select {
				case com, ok := <-input:
					if !ok {
						close(outputChannel)
						return
					}
					resp := com.exec()

					if resp != nil {
						outputChannel <- resp
					}
				}
			}
		}()

	}

	responses := w.output

	go func() {
		defer w.stop()
		closed := 0
		cases := make([]reflect.SelectCase, len(responses))
		for i, ch := range responses {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		}

		for {
			chosen, r, ok := reflect.Select(cases)
			if !ok {
				// handle closed channels
				cases = append(cases[:chosen], cases[chosen+1:]...)
				closed++
				if closed == len(responses) {
					return
				}
				continue
			}
			response, _ := r.Interface().(response)
			if response.err() != "" {
				w.failed = append(w.failed, response)
			}
		}
	}()
}

func (w *workersType) stop() {
	close(w.waiting)
}

func (w *workersType) sendCommand(c command) {
	w.input <- c
}

func (w *workersType) getFailedCommands() []response {
	<-w.waiting // block until all commands finished
	return w.failed
}

//------------------------------------------------------------------------------------

type clickhouseType struct {
	host string
	port uint
	user string
	pwd  string
	main bool
	cli  []string
}

func (ch clickhouseType) printQuery(query string) {
	if !ch.main {
		log.Printf("-- @%s:%d", ch.host, ch.port)
	}

	log.Println(query)
}

func (ch clickhouseType) cmd(query string) (cmd *exec.Cmd) {
	cmd = ch.cmdWithStderr(query, true)
	return
}

func (ch clickhouseType) cmdWithStderr(query string, setStderr bool) (cmd *exec.Cmd) {
	cli := append(ch.cli, "-q", query)
	cmd = exec.Command(cli[0], cli[1:]...)
	if setStderr {
		cmd.Stderr = os.Stderr
	}
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

func (ch clickhouseType) captureErr(query string) (res string) {
	cmd := ch.cmdWithStderr(query, false)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		return "Failed to start process"
	}

	b, _ := ioutil.ReadAll(stderr)
	res = string(b)

	err = cmd.Wait()
	if err != nil && res == "" {
		res = "Unspecified problem running command"
	}

	return
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

func (ch clickhouseType) run(query string, withStdin bool) {
	cmd := ch.cmd(query)

	if withStdin {
		cmd.Stdin = os.Stdin
	}

	ch.printQuery(query)
	if err := cmd.Run(); err != nil {
		os.Exit(-1)
	}
}

func (ch clickhouseType) Pipe(query string) {
	ch.run(query, true)
}

func (ch clickhouseType) Exec(query string) {
	ch.run(query, false)
}
