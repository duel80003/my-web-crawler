package main

import (
	"github.com/jasonlvhit/gocron"
	"my-web-cralwer/delivery"
	"my-web-cralwer/utils"
)

var logger = utils.LoggerInstance()


func main() {
	utils.SetLogLevel()
	err := gocron.Every(1).Day().At("15:30").Do(delivery.StartWebCrawler)
	if err != nil {
		logger.Errorf("task error: %s", err)
	}
	<-gocron.Start()
}
