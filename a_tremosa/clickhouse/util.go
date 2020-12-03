package clickhouse

import (
	"log"
	"os"
	"regexp"
	"unicode"
)

var distRegexp = regexp.MustCompile(`Distributed[[:space:]]*\((.*)\)`)

func trim(r rune) bool {
	return unicode.IsSpace(r) || r == '\'' || r == '\\'
}

func checkStdin() {
	if fi, err := os.Stdin.Stat(); err != nil {
		panic(err)
	} else {
		// Seen here https://stackoverflow.com/questions/22563616/determine-if-stdin-has-data-with-go
		if fi.Mode()&os.ModeNamedPipe == 0 {
			// nothing to load
			log.Println("-- Nothing to load on stdin")
			os.Exit(0)
		}
	}
}
