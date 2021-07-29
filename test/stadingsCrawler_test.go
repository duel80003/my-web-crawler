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

var standingsInfo = driver.StandingsInfo()

func TestStandingCrawler(t *testing.T) {
	utils.SetLogLevel()
	ch := make(chan []*entities.RankingTable)
	done := make(chan bool)
	go func() {
		for {
			select {
			case tables := <-ch:
				if len(tables) > 0 {
					t.Logf("player %+v", tables[0])
				}
			case done := <-done:
				t.Logf("crawl end. %v", done)
				break
			}
		}

	}()
	t.Logf("cralwer players start \n")
	crawler := use_case.NewCrawler(&use_case.StandingsCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	<-done
}

func TestSendingStandingToKafka(t *testing.T) {
	utils.SetLogLevel()
	ch := make(chan []*entities.RankingTable)
	done := make(chan bool)
	go func() {
		for {
			select {
			case tables := <-ch:
				if len(tables) > 0 {
					message, _ := json.Marshal(tables[0])
					standingsInfo.WriteMessages(
						kafka.Message{Value: message},
					)
				}
			case done := <-done:
				t.Logf("crawl end. %v", done)
				break
			}
		}

	}()
	t.Logf("cralwer players start \n")
	crawler := use_case.NewCrawler(&use_case.StandingsCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	<-done
}
