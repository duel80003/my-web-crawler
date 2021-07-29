package test

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"my-web-cralwer/driver"
	"my-web-cralwer/entities"
	use_case "my-web-cralwer/use-case"
	"my-web-cralwer/utils"

	"testing"
)

var simpleInfoProducer = driver.SimplePlayerInfo()

func TestPlayerCrawler(t *testing.T) {
	utils.SetLogLevel()
	ch := make(chan entities.Player)
	done := make(chan bool)
	go func() {
		for {
			select {
			case player := <-ch:
				t.Logf("player %+v", player)
			case done := <-done:
				t.Logf("cralw end. %v", done)
			}
		}

	}()
	t.Logf("cralwer players start \n")
	crawler := use_case.NewCrawler(&use_case.PlayersCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	<-done
}

func TestSendToKafka(t *testing.T) {
	utils.SetLogLevel()
	ch := make(chan entities.Player)
	done := make(chan bool)
	go func() {
		for {
			select {
			case player := <-ch:
				t.Logf("player %+v", player)
				message, _ := json.Marshal(player)
				simpleInfoProducer.WriteMessages(
					kafka.Message{Value: message})
			case done := <-done:
				t.Logf("cralw end. %v", done)
			}
		}

	}()
	t.Logf("cralwer players start \n")
	crawler := use_case.NewCrawler(&use_case.PlayersCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	<-done
}
