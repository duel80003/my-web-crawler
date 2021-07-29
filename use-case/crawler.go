package use_case

type Crawler struct {
	strategy CrawlerStrategy
}

func (c *Crawler) DoWebCrawl() {
	c.strategy.DoWebCrawl()
}
