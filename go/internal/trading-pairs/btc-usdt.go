package trading_pairs

type BtcUsdtPair struct {
	ob *OrderBook
}

func NewBtcUsdtPair() TradingPair {
	return &BtcUsdtPair{
		NewOrderBook(),
	}
}

func (p *BtcUsdtPair) Symbol() string {
	return "btcusdt"
}

func (p *BtcUsdtPair) UpdateAsks(asksToBeUpdated []Ask) {
	p.ob.UpdateAsks(asksToBeUpdated)
}

func (p *BtcUsdtPair) UpdateBids(bidsToBeUpdated []Bid) {
	p.ob.UpdateBids(bidsToBeUpdated)
}
