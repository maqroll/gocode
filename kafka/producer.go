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

	conn, err := kafka.DialLeader(context.Background(), "tcp", *broker, *topic, *partition)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	conn.WriteMessages(
		kafka.Message{Value: []byte(time.Now().String())},
	)

}
