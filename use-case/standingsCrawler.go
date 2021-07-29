package use_case

import (
	"crypto/tls"
	"github.com/gocolly/colly"
	"my-web-cralwer/entities"
	"my-web-cralwer/utils"
	"net/http"
	"regexp"
)

type StandingCrawlerInfo struct {
	Tables []*entities.RankingTable
}

type StandingsCrawler struct {
	Ch chan *StandingCrawlerInfo
}

func (s *StandingsCrawler) DoWebCrawl() {
	defer close(s.Ch)
	endpoint := utils.GetEnv("STANDINGS_TARGET_ENDPOINT")
	c := colly.NewCollector()
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})
	crawlerInfo := &StandingCrawlerInfo{}
	// Find and visit all links
	c.OnHTML(`div[class=RecordTableWrap]`, func(e *colly.HTMLElement) {
		if len(crawlerInfo.Tables) != 0 {
			return
		}
		re := regexp.MustCompile(`[ \n\t\r"]`)
		// contact the table title
		text := e.ChildText("div.record_table_caption")
		rt := &entities.RankingTable{
			Title: text,
			Type:  "standings",
		}
		index := 0
		// find main content element
		e.ForEach("table > tbody", func(_ int, el *colly.HTMLElement) {
			index++
			tableHeader := []string{}
			// find each tables title
			el.ForEach("tr > th", func(_ int, el *colly.HTMLElement) {
				text := string(re.ReplaceAll([]byte(el.Text), []byte("")))
				// logger.Infoln(string(text))
				tableHeader = append(tableHeader, text)
				rt.TableHeader = tableHeader
			})
			// find table content
			tableContent := [][]string{}
			// find each tables content
			el.ForEach("tr", func(_ int, el *colly.HTMLElement) {
				tc := []string{}
				el.ForEach("td", func(i int, el *colly.HTMLElement) {
					if i > 0 {
						text := string(re.ReplaceAll([]byte(el.Text), []byte("")))
						//logger.Infof("aaaaa --- %s --- %d", text, i)
						tc = append(tc, text)
					} else {
						t := el.ChildText("div.sticky_wrap > div.team-w-trophy > a")
						text := string(re.ReplaceAll([]byte(t), []byte("")))
						logger.Debugf("aaaaa --- %s --- %d", text, i)
						tc = append(tc, text)
					}
				})
				if len(tc) > 0 {
					tableContent = append(tableContent, tc)
				}
			})
			rt.TableContent = tableContent
		})
		crawlerInfo.Tables = append(crawlerInfo.Tables, rt)
	})
	c.OnError(func(response *colly.Response, err error) {
		logger.Infoln("error", err)
		s.Ch <- nil
	})
	c.OnRequest(func(r *colly.Request) {
		logger.Infoln("Visiting", r.URL.String())
	})
	c.Visit(endpoint)
	s.Ch <- crawlerInfo
}
