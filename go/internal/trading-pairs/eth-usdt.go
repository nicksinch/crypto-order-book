package trading_pairs

type EthUsdtPair struct {
	ob *OrderBook
}

func NewEthUsdtPair() TradingPair {
	return &EthUsdtPair{
		NewOrderBook(),
	}
}

func (p *EthUsdtPair) Symbol() string {
	return "ethusdt"
}

func (p *EthUsdtPair) UpdateAsks(asksToBeUpdated []Ask) {
	p.ob.UpdateAsks(asksToBeUpdated)
}

func (p *EthUsdtPair) UpdateBids(bidsToBeUpdated []Bid) {
	p.ob.UpdateBids(bidsToBeUpdated)
}
