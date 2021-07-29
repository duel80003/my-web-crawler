package use_case

import (
	"crypto/tls"
	"github.com/gocolly/colly"
	"golang.org/x/net/http2"
	"my-web-cralwer/entities"
	"my-web-cralwer/utils"
	"net/url"
	"regexp"
	"strings"
)

var logger = utils.LoggerInstance()

func NewCrawler(strategy CrawlerStrategy) *Crawler {
	return &Crawler{
		strategy: strategy,
	}
}

type PlayersCrawler struct {
	Ch chan entities.Player
	Done chan bool
}

func (p *PlayersCrawler) DoWebCrawl() {
	defer close(p.Ch)
	defer close(p.Done)
	logger.Info("start web crawl")
	var c = colly.NewCollector()
	var re = regexp.MustCompile(`[ *◎#]`)
	var logger = utils.LoggerInstance()
	basicURL := utils.GetEnv("TARGET_ENDPOINT")
	c.WithTransport(&http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})
	c.OnHTML("div.PlayersList> dl > dd > a", func(e *colly.HTMLElement) {
		name := string(re.ReplaceAll([]byte(e.Text), []byte("")))
		href := e.Attr("href")
		str := strings.Split(href, "?")
		m, _ := url.ParseQuery(str[1])
		playerID := m.Get("acnt")
		p.Ch <- entities.Player{
			Name: name,
			ID:   playerID,
		}
	})
	c.OnError(func(response *colly.Response, err error) {
		logger.Errorf("c.OnError: %s", err)
	})
	c.OnRequest(func(r *colly.Request) {
		logger.Infoln("Visiting", r.URL.String())
	})
	c.Visit(basicURL)
	p.Done <- true
}
