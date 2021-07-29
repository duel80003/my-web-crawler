package test

import (
	"my-web-cralwer/delivery"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestPlayersCrawlerHandler(t *testing.T) {
	wg.Add(1)
	go delivery.PlayersCrawlerHandler(&wg)
	wg.Wait()
}

func TestStandingsCrawler(t *testing.T) {
	wg.Add(1)
	go delivery.StandingsCrawler(&wg)
	wg.Wait()
}
