package models

type RatesDTO struct {
	Timestamp int64   `json:"timestamp"`
	AskPrice  float32 `json:"asks"`
	BidPrice  float32 `json:"bids"`
}
