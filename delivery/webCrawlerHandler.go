package delivery

import (
	"my-web-cralwer/entities"
	use_case "my-web-cralwer/use-case"
	"my-web-cralwer/utils"
	"sync"
)

var logger = utils.LoggerInstance()

func StartWebCrawler() {
	logger.Info("Start Web crawl")
	var wg sync.WaitGroup
	wg.Add(3)
	go PlayersCrawlerHandler(&wg)
	//go StandingsCrawler(&wg)
	//go PlayerDetailsCrawler(&wg)
	wg.Wait()
}

func PlayersCrawlerHandler(wg *sync.WaitGroup) {
	ch := make(chan entities.Player)
	done := make(chan bool)
	logger.Infof("cralwer players start")
	//go playersProcess(ch, done)
	crawler := use_case.NewCrawler(&use_case.PlayersCrawler{Ch: ch})
	crawler.DoWebCrawl()
	<-done
	wg.Done()
}

//func StandingsCrawler(wg *sync.WaitGroup) {
//	ch := make(chan *use_case.StandingCrawlerInfo)
//	done := make(chan bool)
//	logger.Infof("cralwer standings list start")
//	go standingsProcess(ch, done)
//	crawler := use_case.NewCrawler(&use_case.StandingsCrawler{Ch: ch})
//	crawler.DoWebCrawl()
//	<-done
//	wg.Done()
//}

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

//func playersProcess(ch <-chan []*entities.Player, done chan<- bool) {
//	defer close(done)
//	for players := range ch {
//		if len(players) == 0 {
//			done <- true
//			break
//		}
//		playerStore := store.New(dbClient)
//		playerStore.BatchUpsertWithoutUpdate(players)
//		done <- true
//	}
//}

//func standingsProcess(ch <-chan *use_case.StandingCrawlerInfo, done chan<- bool) {
//	defer close(done)
//	for info := range ch {
//		if info == nil {
//			done <- true
//			break
//		}
//		ss := standingsStore.New(redisClient)
//		ss.Save(info)
//		done <- true
//	}
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
