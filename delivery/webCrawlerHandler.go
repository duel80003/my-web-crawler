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
	logger                   = utils.LoggerInstance()
	simpleInfoProducer       = driver.SimplePlayerInfo()
	standingsInfoProducer    = driver.StandingsInfo()
	playerDetailInfoProducer = driver.PlayerDetailInfo()
)

func StartWebCrawler() {
	logger.Info("Start Web crawl")
	var wg sync.WaitGroup
	players := make(chan []*entities.Player)
	wg.Add(3)
	go PlayersCrawlerHandler(&wg, players)
	go StandingsCrawler(&wg)
	go PlayerDetailsCrawler(&wg, players)
	wg.Wait()
}

func PlayersCrawlerHandler(wg *sync.WaitGroup, playersCh chan<- []*entities.Player) {
	ch := make(chan entities.Player)
	done := make(chan bool)
	go simplePlayerInfoProcess(ch, done, playersCh)
	logger.Printf("cralwer players start")
	crawler := use_case.NewCrawler(&use_case.PlayersCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	wg.Done()
}

func simplePlayerInfoProcess(ch <-chan entities.Player, done <-chan bool, playersCh chan<- []*entities.Player) {
	defer close(playersCh)
	var players []*entities.Player
Loop:
	for {
		select {
		case player := <-ch:
			logger.Debugf("player %+v", player)
			players = append(players, &player)
			message, _ := json.Marshal(player)
			simpleInfoProducer.WriteMessages(
				kafka.Message{Value: message},
			)
		case done, ok := <-done:
			if ok {
				playersCh <- players
				logger.Printf("simple player info crawl end. %v", done)
				break Loop
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
Loop:
	for {
		select {
		case tables := <-ch:
			logger.Debugf("tables %+v", tables)
			if len(tables) > 0 {
				message, _ := json.Marshal(tables[0])
				standingsInfoProducer.WriteMessages(
					kafka.Message{Value: message},
				)
			}
		case done, ok := <-done:
			if ok {
				logger.Infof("simple player info crawl end. %v", done)
				break Loop
			}
		}
	}
}

func PlayerDetailsCrawler(wg *sync.WaitGroup, playersCh <-chan []*entities.Player) {
	for players := range playersCh {
		if players != nil {
			logger.Infof("cralwer player details start")
			logger.Debugf("players %+v, %d", players, len(players))
			ch := make(chan *use_case.CrawlerInfo)
			done := make(chan bool)
			go playerDetailsProcess(ch, done)
			crawler := use_case.NewCrawler(&use_case.PlayerDetailCrawler{
				Ch:      ch,
				Done:    done,
				Players: players,
			})
			crawler.DoWebCrawl()
			wg.Done()
		}
	}
}

func playerDetailsProcess(ch <-chan *use_case.CrawlerInfo, done <-chan bool) {
Loop:
	for {
		select {
		case playerDetailInfo := <-ch:
			logger.Infof("send message.")
			message, _ := json.Marshal(playerDetailInfo)
			playerDetailInfoProducer.WriteMessages(
				kafka.Message{Value: message},
			)

		case done, ok := <-done:
			if ok {
				logger.Infoln("simple player info crawl end. %v", done)
				break Loop
			}
		}
	}
}
