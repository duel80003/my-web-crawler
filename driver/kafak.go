package driver

import (
	"context"
	"github.com/segmentio/kafka-go"
	"my-web-cralwer/utils"
	"time"
)

var (
	partition         = 0
	logger            = utils.LoggerInstance()
	kafkaHost         = utils.GetEnv("KAKFA_HOST")
	simplePlayerTopic = utils.GetEnv("SIMPLE_PLAYER_INFO_TOPIC")
	standingInfoTopic = utils.GetEnv("STANDING_INFO_TOPIC")
	playerDetailTopic = utils.GetEnv("PLAYER_DETAIL_TOPIC")
)

func getConnection(topic string) *kafka.Conn {
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaHost, topic, partition)
	if err != nil {
		logger.Panicf("Kafka init connection error %s", err)
	}
	err = conn.SetWriteDeadline(time.Now().Add(1 * time.Hour))
	if err != nil {
		logger.Panicf("SetWriteDeadline error %s", err)
	}
	return conn
}

func SimplePlayerInfo() *kafka.Conn {
	return getConnection(simplePlayerTopic)
}

func StandingsInfo() *kafka.Conn {
	return getConnection(standingInfoTopic)
}

func PlayerDetailInfo() *kafka.Conn {
	return getConnection(playerDetailTopic)
}
