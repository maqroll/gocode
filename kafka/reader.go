package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/segmentio/kafka-go"
)

var topic = flag.String("topic", "topic", "topic")
var partition = flag.Int("partition", 0, "partition")
var offset = flag.Int64("offset", 0, "offset")
var broker = flag.String("broker", "localhost:3000", "broker location")

func main() {

	flag.Parse()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{*broker},
		Topic:     *topic,
		Partition: *partition,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	r.SetOffset(*offset)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	r.Close()
}
