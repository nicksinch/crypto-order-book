package internal

type eventData struct {
	EventType               string `json:"e"`
	EventTime               int64  `json:"E"`
	TransactionTime         int64  `json:"T"`
	Symbol                  string `json:"s"`
	FirstUpdateId           int64  `json:"U"`
	FinalUpdateId           int64  `json:"u"`
	FinalUpdateIdLastStream int64  `json:"pu"`
	BidsToBeUpdated         []bid  `json:"b"`
	AsksToBeUpdated         []ask  `json:"a"`
}

type depthUpdateEvent struct {
	Stream string    `json:"stream"`
	Data   eventData `json:"data"`
}

type depthSnapshot struct {
	LastUpdateId    int64 `json:"lastUpdateId"`
	EventTime       int64 `json:"E"`
	TransactionTime int64 `json:"T"`
	Bids            []bid `json:"bids"`
	Asks            []ask `json:"asks"`
}
