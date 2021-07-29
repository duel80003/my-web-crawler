package entities

type PlayerType string

func (p PlayerType) Str() string {
	return string(p)
}

const (
	pitcher PlayerType = "pitcher"
	batter  PlayerType = "batter"
)

type PlayerInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Number     string `json:"number"`
	Team       string `json:"team"`
	Position   string `json:"position"`
	Type       string `json:"type"`
	Height     string `json:"height"`
	Weight     string `json:"weight"`
	Birthday   string `json:"birthday"`
	FirstGame  string `json:"FirstGame"`
	PlayerType string `json:"playerType"`
}

func (p PlayerInfo) isPitcher() bool {
	return p.PlayerType == pitcher.Str()
}

func (p *PlayerInfo) GetKey() string {
	if p.PlayerType == pitcher.Str() {
		return "pitching"
	}
	return "batting"
}

func (p *PlayerInfo) SetPlayerType() {
	if p.Position == "投手" {
		p.PlayerType = pitcher.Str()
		return
	}
	p.PlayerType = batter.Str()
}

func (p *PlayerInfo) GetPlayerType() string {
	return p.PlayerType
}
