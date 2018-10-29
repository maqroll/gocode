package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/segmentio/kafka-go"
	"time"
)

var topic = flag.String("topic", "topic", "topic")
var partition = flag.Int("partition", 0, "partition")
var broker = flag.String("broker", "localhost:3000", "broker location")

func main() {

	flag.Parse()

	conn, _ := kafka.DialLeader(context.Background(), "tcp", *broker, *topic, *partition)

	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	defer batch.Close()

	for {
		b := make([]byte, 10e3) // 10KB max per message
		_, err := batch.Read(b)
		if err != nil {
			break
		}
		fmt.Println(string(b))
		fmt.Println(batch.Offset())
	}
}
