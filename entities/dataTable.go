package entities

type DataTable struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Title        string      `json:"title"`
	TableHeader  []string    `json:"header"`
	TableContent interface{} `json:"content"`
	IsEmpty      bool        `json:"isEmpty"`
}

func (p *DataTable) SetStatus() {
	if p.TableContent == "" {
		p.IsEmpty = true
		return
	}
	p.IsEmpty = false
}
