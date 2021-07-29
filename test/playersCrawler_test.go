package test

import (
	"my-web-cralwer/entities"
	use_case "my-web-cralwer/use-case"
	"testing"
)

func TestPlayerCrawler (t *testing.T) {
	ch := make(chan entities.Player)
	done := make(chan bool)
	go func() {
		for {
			select {
				case player := <- ch: {
					t.Logf("player %+v", player)
				}
				case done := <-done: {
					t.Logf("cralw end. %v", done)
				}
			}
		}

	}()
	t.Logf("cralwer players start \n")
	crawler := use_case.NewCrawler(&use_case.PlayersCrawler{Ch: ch, Done: done})
	crawler.DoWebCrawl()
	<-done
}