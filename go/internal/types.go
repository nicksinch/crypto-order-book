package internal

import tradingpairs "crypto-order-book/go/internal/trading-pairs"

type eventData struct {
	EventType               string             `json:"e"`
	EventTime               int64              `json:"E"`
	TransactionTime         int64              `json:"T"`
	Symbol                  string             `json:"s"`
	FirstUpdateId           int64              `json:"U"`
	FinalUpdateId           int64              `json:"u"`
	FinalUpdateIdLastStream int64              `json:"pu"`
	BidsToBeUpdated         []tradingpairs.Bid `json:"b"`
	AsksToBeUpdated         []tradingpairs.Ask `json:"a"`
}

type DepthUpdateEvent struct {
	Stream string    `json:"stream"`
	Data   eventData `json:"data"`
}

type DepthSnapshot struct {
	LastUpdateId    int64              `json:"lastUpdateId"`
	EventTime       int64              `json:"E"`
	TransactionTime int64              `json:"T"`
	Bids            []tradingpairs.Bid `json:"bids"`
	Asks            []tradingpairs.Ask `json:"asks"`
}
