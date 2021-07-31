package use_case

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/levigross/grequests"
	"golang.org/x/net/http2"
	"my-web-cralwer/entities"
	"my-web-cralwer/utils"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CrawlerInfo struct {
	PlayerID string                 `json:"playerId"`
	Data     map[string]interface{} `json:"data"`
}

var re = regexp.MustCompile(`[ "/]`)
var re1 = regexp.MustCompile(`[ \n"]`)
var re2 = regexp.MustCompile(`[']`)
var wg sync.WaitGroup

const (
	pitchingTitle string = "投球成績"
	battingTitle  string = "打擊成績"
	battleTitle   string = "對戰成績"
	defenceTitle  string = "守備成績"
	followTitle   string = "逐場成績表"
)

type request struct {
	PlayerId      string
	URI           string
	Headers       *map[string]string
	DefendStation string
}

type jsonRes struct {
	Success      string
	Type         string
	BattingScore string
	DefenceScore string
	FighterScore string
	PitchScore   string
	FollowScore  string
}

type PlayerDetailCrawler struct {
	Ch      chan *CrawlerInfo
	Done    chan bool
	Players []*entities.Player
}

func (p *PlayerDetailCrawler) DoWebCrawl() {
	var headersChan chan map[string]string
	defer close(p.Ch)
	defer close(p.Done)
	//defer close(headersChan)
	logger.Infof("players count %d", len(p.Players))
	count := len(p.Players)

	headersChan = getToken()
	headers := <-headersChan

	wg.Add(count)
	for i := 0; i < count; i++ {
		logger.Info("player id", p.Players[i].ID)
		go p.DoWebCrawlImp(p.Players[i], headers)
		if i%5 == 0 && i != 0 {
			time.Sleep(time.Second * 3)
		}
	}
	wg.Wait()
	p.Done <- true
}

func (p *PlayerDetailCrawler) DoWebCrawlImp(player *entities.Player, headers map[string]string) {
	logger.Infoln("start web crawler")
	defer wg.Done()
	p1 := make(chan []*entities.DataTable)
	p2 := make(chan *entities.PlayerInfo)
	p3 := make(chan *entities.DataTable)
	count := make(chan int)
	end := make(chan bool)

	var playerTables []*entities.DataTable
	var playerInfo *entities.PlayerInfo

	go crawlPersonInfo(player.ID, p1, p2)
	go crawlFollowInfo(player.ID, p3)
	go func() {
		defer close(end)
		index := 0
	Loop:
		for {
			select {
			case <-count:
				index++
				if index == 3 {
					end <- true
					break Loop
				}
			}
		}
	}()
Loop:
	for {
		// Await both of these values
		// simultaneously, printing each one as it arrives.
		select {
		case pts, ok := <-p1:
			if ok {
				logger.Debugf("playerTables msg %+v", pts)
				playerTables = append(playerTables, pts...)
				count <- 1
			}
		case p, ok := <-p2:
			if ok {
				playerInfo = p
				logger.Debugf("playerInfo msg %+v", playerInfo)
				count <- 1
			}
		case playerTable, ok := <-p3:
			if ok {
				logger.Debugf("playerTable msg %+v", playerTable)
				playerTables = append(playerTables, playerTable)
				count <- 1
			}
		case <-end:
			close(count)
			break Loop
		}
	}
	m := fetchPersonRequest(player.ID, playerInfo.Position, headers, playerTables)
	pi, _ := json.Marshal(playerInfo)
	m["playerInfo"] = pi
	p.Ch <- &CrawlerInfo{
		PlayerID: player.ID,
		Data:     m,
	}
}

func getToken() chan map[string]string {
	logger.Infoln("get token start")
	ch := make(chan map[string]string)
	go func() {
		url := utils.GetEnv("BASE_URL")
		c1 := colly.NewCollector()
		c1.WithTransport(&http2.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})
		c1.OnHTML("html", func(e *colly.HTMLElement) {
			headers := make(map[string]string)
			text := e.ChildText("body > script")
			str := "{" + text[strings.Index(text, "RequestVerificationToken"):strings.Index(text, "},")] + "}"
			str = string(re1.ReplaceAll([]byte(str), []byte("")))
			str = string(re2.ReplaceAll([]byte(str), []byte("\"")))
			str = strings.Replace(str, "RequestVerificationToken", "\"requestverificationtoken\"", -1)
			str = strings.ReplaceAll(str, "'", "\"")
			logger.Infof("str %s", str)
			headers["x-requested-with"] = "XMLHttpRequest"
			headers["Content-Type"] = "application/x-www-form-urlencoded"

			err := json.Unmarshal([]byte(str), &headers)
			if err != nil {
				logger.Errorf("err %s", err)
			}
			logger.Infof("headers %+v", headers)
			ch <- headers
		})
		c1.OnError(func(response *colly.Response, err error) {
			logger.Infoln("error", err)
		})
		c1.OnRequest(func(r *colly.Request) {
			logger.Infoln("Visiting", r.URL.String())
		})
		c1.Visit(url)
	}()
	return ch
}

func crawlPersonInfo(playerId string, ch1 chan<- []*entities.DataTable, ch2 chan<- *entities.PlayerInfo) {
	defer close(ch2)
	defer close(ch1)
	url := utils.GetEnv("PLAYER_DETAILS_TARGET_ENDPOINT")
	endPoint := fmt.Sprintf("%s/person?acnt=%s", url, playerId)
	playerInfo := &entities.PlayerInfo{}
	var playerTables []*entities.DataTable
	var c = colly.NewCollector()

	c.WithTransport(&http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})
	c.OnHTML("div.PlayerBrief > div > dl", func(e *colly.HTMLElement) {
		playerInfo.Team = e.ChildText("dt > div.team")
		playerInfo.Number = e.ChildText("dt > div.name > span.number")
		playerInfo.Birthday = e.ChildText("dd.born > div.desc")
		playerInfo.FirstGame = e.ChildText("dd.debut > div.desc")
		playerInfo.Type = e.ChildText("dd.b_t > div.desc")
		playerInfo.Position = e.ChildText("dd.pos > div.desc")
		playerInfo.ID = playerId
		playerInfo.SetPlayerType()

		name := e.ChildText("dt > div.name")
		playerInfo.Name = strings.Replace(name, playerInfo.Number, "", -1)

		heightAndWeight := e.ChildText("dd.ht_wt > div.desc")
		str := strings.Split(heightAndWeight, "/")
		playerInfo.Height = re.ReplaceAllString(str[0], "")
		playerInfo.Weight = re.ReplaceAllString(str[1], "")
	})

	c.OnHTML("div[id=bindVue]", func(e *colly.HTMLElement) {
		playerType := playerInfo.GetPlayerType()
		e.ForEach("div.DistTitle > h3", func(i int, el *colly.HTMLElement) {
			if ok := isValidIndex(true, playerType, i); !ok {
				return
			}
			playerTable := &entities.DataTable{}
			playerTable.Title = el.Text
			playerTable.Type = getType(el.Text)
			playerTables = append(playerTables, playerTable)
		})
		index := 0
		e.ForEach("div.RecordTableOuter > div.RecordTable > table > tbody ", func(i int, el *colly.HTMLElement) {
			if ok := isValidIndex(false, playerType, i); !ok {
				return
			}
			var tableHeader []string
			el.ForEach("tr > th", func(i int, el *colly.HTMLElement) {
				text := el.Text
				tableHeader = append(tableHeader, re1.ReplaceAllString(text, ""))
			})
			playerTables[index].TableHeader = tableHeader
			index++
		})
	})

	c.OnError(func(response *colly.Response, err error) {
		logger.Infoln("error", err)
	})
	c.OnRequest(func(r *colly.Request) {
		logger.Infoln("Visiting", r.URL.String())
	})
	c.Visit(endPoint)
	ch1 <- playerTables
	ch2 <- playerInfo
}

func crawlFollowInfo(playerId string, ch chan<- *entities.DataTable) {
	defer close(ch)
	url := utils.GetEnv("PLAYER_DETAILS_TARGET_ENDPOINT")
	playerTable := &entities.DataTable{}
	var c = colly.NewCollector()
	var pos string

	c.WithTransport(&http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnHTML("div.PlayerBrief > div > dl", func(e *colly.HTMLElement) {
		pos = e.ChildText("dd.pos > div.desc")
	})

	c.OnHTML("div[id=bindVue]", func(e *colly.HTMLElement) {
		title := e.ChildText("div.DistTitle > h3")
		playerTable.Title = title
		playerTable.Type = getType(title)
		logger.Infoln("player pos", pos)
		e.ForEach("div.RecordTableOuter > div.RecordTable > table ", func(i int, el *colly.HTMLElement) {
			var tableHeader []string
			if pos == "投手" && i == 1 {
				el.ForEach("tr > th", func(i int, el *colly.HTMLElement) {
					text := el.Text
					tableHeader = append(tableHeader, re1.ReplaceAllString(text, ""))
				})
				playerTable.TableHeader = tableHeader

			} else if pos != "投手" && i == 0 {
				el.ForEach("tr > th", func(i int, el *colly.HTMLElement) {
					text := el.Text
					tableHeader = append(tableHeader, re1.ReplaceAllString(text, ""))
				})
				playerTable.TableHeader = tableHeader
			}
		})
	})
	c.OnError(func(response *colly.Response, err error) {
		logger.Infoln("error", err)
	})
	c.OnRequest(func(r *colly.Request) {
		logger.Infoln("Visiting", r.URL.String())
	})
	endPoint := fmt.Sprintf("%s/follow?acnt=%s", url, playerId)
	c.Visit(endPoint)
	ch <- playerTable
}

func isValidIndex(isTableHeader bool, playerType string, index int) bool {
	var x [3]int
	if isTableHeader {
		// the page put all record together,
		// header index 0, 2, 3 is batter header
		if playerType == "batter" {
			x = [3]int{0, 2, 3}
		} else {
			x = [3]int{1, 2, 3}
		}
	} else {
		// the page put all record together,
		// header index 0, 2, 4 is batter header
		if playerType == "batter" {
			x = [3]int{0, 2, 4}
		} else {
			x = [3]int{1, 2, 3}
		}
	}
	for _, v := range x {
		if index == v {
			return true
		}
	}
	return false
}

func getType(text string) string {
	switch text {
	case battingTitle:
		return "batting"
	case pitchingTitle:
		return "pitch"
	case battleTitle:
		return "fight"
	case defenceTitle:
		return "defence"
	case followTitle:
		return "follow"
	default:
		return ""
	}
}

func fetchPersonRequest(playerId, position string, headers map[string]string, playerTables []*entities.DataTable) map[string]interface{} {
	logger.Infoln("fetchPersonRequest start, playerTables count", len(playerTables))
	result := make(map[string]interface{})
	requestChan := make(chan *jsonRes)
	end := make(chan bool)
	count := make(chan int)
	const total = 5
	uri := []string{
		"getbattingscore",
		"getdefencescore",
		"getfighterscore",
		"getpitchscore",
		"getfollowscore",
	}

	go func() {
		defer close(count)
		defer close(end)
		index := 0
	Loop:
		for {
			select {
			case <-count:
				index++
				if index >= total {
					end <- true
					break Loop
				}
			}
		}
	}()

	for i := 0; i < len(playerTables); i++ {
		p := playerTables[i]
		result[p.Type] = p
	}

	for i := 0; i < total; i++ {
		r := &request{
			PlayerId:      playerId,
			URI:           uri[i],
			Headers:       &headers,
			DefendStation: position,
		}
		go playerDetailsRequest(r, requestChan)
	}
Loop:
	for {
		select {
		case res := <-requestChan:
			switch res.Type {
			case "getbattingscore":
				p, ok := result["batting"]
				if ok {
					x := p.(*entities.DataTable)
					x.TableContent = res.BattingScore
					x.ID = playerId
					x.SetStatus()
					d, _ := json.Marshal(x)
					result["batting"] = d
				}
				count <- 1
			case "getdefencescore":
				p, ok := result["defence"]
				if ok {
					x := p.(*entities.DataTable)
					x.TableContent = res.DefenceScore
					x.ID = playerId
					x.SetStatus()
					d, _ := json.Marshal(x)
					result["defence"] = d
				}
				count <- 1
			case "getfighterscore":
				p, ok := result["fight"]
				if ok {
					x := p.(*entities.DataTable)
					x.TableContent = res.FighterScore
					x.ID = playerId
					x.SetStatus()
					d, _ := json.Marshal(x)
					result["fight"] = d
				}
				count <- 1
			case "getpitchscore":
				p, ok := result["pitch"]
				if ok {
					x := p.(*entities.DataTable)
					x.TableContent = res.PitchScore
					x.ID = playerId
					x.SetStatus()
					d, _ := json.Marshal(x)
					result["pitch"] = d
				}
				count <- 1
			case "getfollowscore":
				p, ok := result["follow"]
				if ok {
					x := p.(*entities.DataTable)
					x.TableContent = res.FollowScore
					x.ID = playerId
					x.SetStatus()
					d, _ := json.Marshal(x)
					logger.Debugf("getfollowscore %+v", res.FollowScore)
					result["follow"] = d
				}
				count <- 1
			case "failure":
			default:
				logger.Errorf("unKnow jsonRes type %s", res.Type)
			}
		case <-end:
			break Loop
		default:

		}
	}
	return result
}

func playerDetailsRequest(req *request, ch chan<- *jsonRes) {
	url := utils.GetEnv("PLAYER_DETAILS_TARGET_ENDPOINT")
	endpoint := fmt.Sprintf("%s/%s", url, req.URI)
	logger.Infoln("target endpoint:", endpoint)
	data := make(map[string]string)
	data["acnt"] = req.PlayerId
	data["kindCode"] = "A"

	if req.URI == "getfighterscore" {
		data["defendStation"] = req.DefendStation
	}
	if req.URI == "getfollowscore" {
		data["defendStation"] = req.DefendStation
		data["year"] = strconv.Itoa(time.Now().Year())
	}
	logger.Infof("requestd ata: %+v", data)

	t := &jsonRes{}
	t.Type = req.URI
	resp, err := grequests.Post(endpoint, &grequests.RequestOptions{
		Headers:            *req.Headers,
		Data:               data,
		InsecureSkipVerify: true,
	})

	defer resp.Close()

	if err != nil {
		logger.Errorf("%s error: %s", req.URI, err)
		t.Type = "failure"
		ch <- t
		return
	}
	logger.Debug("is ok ", resp.Ok)
	bytes := resp.Bytes()
	logger.Debugf("uri:%s string: %s", req.URI, string(bytes))
	_ = json.Unmarshal(bytes, &t)
	ch <- t
}
