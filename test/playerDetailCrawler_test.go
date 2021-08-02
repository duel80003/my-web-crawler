package test

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"my-web-cralwer/driver"
	"my-web-cralwer/entities"
	use_case "my-web-cralwer/use-case"
	"testing"
)

var playerDetailInfoProducer = driver.PlayerDetailInfo()

func TestPlayerDetailCrawler(t *testing.T) {
	players := []*entities.Player{
		//&entities.Player{
		//	Name: "吳哲源",
		//	ID:   "0000001515",
		//},
		&entities.Player{
			Name: "王威晨",
			ID:   "0000000929",
		},
	}

	t.Log("crawler player details start")
	t.Logf("players %+v, %d", players, len(players))
	ch := make(chan *use_case.CrawlerInfo)
	done := make(chan bool)
	go func() {
		for {
			select {
			case playerDetailInfo, ok := <-ch:
				if ok {
					t.Logf("player info %+v", playerDetailInfo)
					message, _ := json.Marshal(playerDetailInfo)
					playerDetailInfoProducer.WriteMessages(
						kafka.Message{Value: message},
					)
				}
			case done, ok := <-done:
				if ok {
					t.Logf("player details info crawl end. %v", done)
					break
				}
			}
		}
	}()

	crawler := use_case.NewCrawler(&use_case.PlayerDetailCrawler{
		Ch:      ch,
		Done:    done,
		Players: players,
	})
	crawler.DoWebCrawl()
}
