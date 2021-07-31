package test

import (
	"my-web-cralwer/delivery"
	"my-web-cralwer/entities"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestPlayersCrawlerHandler(t *testing.T) {
	players := make(chan []*entities.Player)
	wg.Add(1)
	go delivery.PlayersCrawlerHandler(&wg, players)
	wg.Wait()
}

func TestStandingsCrawler(t *testing.T) {
	wg.Add(1)
	go delivery.StandingsCrawler(&wg)
	wg.Wait()
}
