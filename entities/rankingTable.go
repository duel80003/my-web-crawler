package entities

type RankingTable struct {
	Type         string     `json:"type"`
	Title        string     `json:"title"`
	TableHeader  []string   `json:"header"`
	TableContent [][]string `json:"content"`
}
