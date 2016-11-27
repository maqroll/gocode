package main

import (
	"bytes"
	"encoding/json"
	"github.com/atotto/clipboard"
)

const ()

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func main() {
	text, _ := clipboard.ReadAll()
	if result, err := prettyprint([]byte(text)); err != nil {
		clipboard.WriteAll(err.Error())
	} else {
		clipboard.WriteAll(string(result))
	}
}
