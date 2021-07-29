package delivery

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"my-web-cralwer/driver"
	"my-web-cralwer/entities"
	use_case "my-web-cralwer/use-case"
	"my-web-cralwer/utils"
	"sync"
)

var (
	logger             = utils.LoggerInstance()
	simpleInfoProducer = driver.SimplePlayerInfo()
	standingsInfo      = driver.StandingsInfo()
)

func StartWebCrawler() {
	logger.Info("Start Web crawl")
	var wg sync.WaitGroup
	wg.Add(3)
	go PlayersCrawlerHandler(&wg)
	go StandingsCrawler(&wg)
	//go PlayerDetailsCrawler(&wg)
	wg.Wait()
}

func PlayersCrawlerHandler(wg *sync.WaitGroup) {
	ch := make(chan entities.Player)
	done := make(chan bool)
	go simplePlayerInfoProcess(ch, done)
	logger.Printf("cralwer players start")
	crawler := use_case.NewCrawler(&use_case.PlayersCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	wg.Done()
}

func simplePlayerInfoProcess(ch <-chan entities.Player, done <-chan bool) {
	for {
		select {
		case player := <-ch:
			logger.Printf("player %+v", player)
			message, _ := json.Marshal(player)
			simpleInfoProducer.WriteMessages(
				kafka.Message{Value: message},
			)
		case done, ok := <-done:
			if ok {
				logger.Printf("simple player info crawl end. %v", done)
				break
			}
		}
	}
}

func StandingsCrawler(wg *sync.WaitGroup) {
	ch := make(chan []*entities.RankingTable)
	done := make(chan bool)
	logger.Infof("cralwer standings list start")
	go standingsProcess(ch, done)
	crawler := use_case.NewCrawler(&use_case.StandingsCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	wg.Done()
}

func standingsProcess(ch <-chan []*entities.RankingTable, done <-chan bool) {
	for {
		select {
		case tables := <-ch:
			logger.Printf("tables %+v", tables)
			if len(tables) > 0 {
				message, _ := json.Marshal(tables[0])
				standingsInfo.WriteMessages(
					kafka.Message{Value: message},
				)
			}
		case done, ok := <-done:
			if ok {
				logger.Printf("simple player info crawl end. %v", done)
				break
			}
		}
	}
}

//func PlayerDetailsCrawler(topWg *sync.WaitGroup) {
//	logger.Infof("cralwer player details start")
//	playerStore := store.New(dbClient)
//	players, err := playerStore.GetPlayers()
//	if err != nil {
//		return
//	}
//	logger.Debugf("players %+v, %d", players, len(players))
//	ch := make(chan *use_case.CrawlerInfo)
//	var wg sync.WaitGroup
//	wg.Add(len(players))
//	go playerDetailsProcess(ch, &wg)
//	crawler := use_case.NewCrawler(&use_case.PlayerDetailCrawler{
//		Ch:      ch,
//		Players: players,
//	})
//	crawler.DoWebCrawl()
//	wg.Wait()
//	topWg.Done()
//}

//
//func playerDetailsProcess(ch <-chan *use_case.CrawlerInfo, wg *sync.WaitGroup) {
//	for crawlerInfo := range ch {
//		if crawlerInfo == nil {
//			wg.Done()
//			continue
//		}
//		ss := standingsStore.New(redisClient)
//		ss.SavePlayerDetails(crawlerInfo)
//		wg.Done()
//	}
//}
