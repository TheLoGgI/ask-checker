package structs

type Ask struct {
	Isin   string `json:"isin"`
	Name   string `json:"name"`
	Lai    string `json:"lai"`
	Ticker string `json:"ticker"`
}
