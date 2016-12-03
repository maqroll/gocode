package main

import (
	"bytes"
	"encoding/xml"
	"github.com/atotto/clipboard"
	"io"
	"strings"
)

const ()

func prettyprint(b []byte) ([]byte, error) {
	var err error = nil
	buf := new(bytes.Buffer)
	d := xml.NewDecoder(strings.NewReader(string(b)))
	e := xml.NewEncoder(buf)
	e.Indent("", " ")

tokenize:
	for {
		tok, err := d.Token()
		switch {
		case err == io.EOF:
			e.Flush()
			break tokenize
		case err != nil:
			return nil, err
		}
		e.EncodeToken(tok)
	}

	return buf.Bytes(), err
}

func main() {
	text, _ := clipboard.ReadAll()
	if result, err := prettyprint([]byte(text)); err != nil {
		clipboard.WriteAll(err.Error())
	} else {
		clipboard.WriteAll(string(result))
	}
}
