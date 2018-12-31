package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/segmentio/kafka-go"
)

var topic = flag.String("topic", "topic", "topic")

//var partition = flag.Int("partition", 0, "partition")
var cg = flag.String("cg", "consumer-group-id", "consumer group id")
var broker = flag.String("broker", "localhost:3000", "broker location")

func main() {

	flag.Parse()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{*broker},
		GroupID:  *cg,
		Topic:    *topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
	}

	r.Close()

}
