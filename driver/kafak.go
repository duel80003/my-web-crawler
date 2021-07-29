package driver

import (
	"context"
	"github.com/segmentio/kafka-go"
	"time"
)

var (
	topic = "my-topic"
	partition = 0
)

func GetKafkaConnection() (*kafka.Conn, error) {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		return nil, err
	}
	err = conn.SetWriteDeadline(time.Now().Add(10*time.Second))
	return conn, err
}