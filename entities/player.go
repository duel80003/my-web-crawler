package entities

import (
	"database/sql"
	"time"
)

type Player struct {
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
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  sql.NullTime
}
