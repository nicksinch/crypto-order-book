package trading_pairs

import (
	"slices"
	"sort"
	"strconv"
)

type Bid [2]string // price level to be updated; quantity
type Ask [2]string // price level to be updated; quantity

type askOffer struct {
	price    float64
	quantity string
}

type bidOffer struct {
	price    float64
	quantity string
}

type TradingPair interface {
	Symbol() string
	UpdateAsks(asksToBeUpdated []Ask)
	UpdateBids(bidsToBeUpdated []Bid)
}

type OrderBook struct {
	asksCount int
	bidsCount int
	askIds    map[string]int
	bidIds    map[string]int
	asks      []askOffer
	bids      []bidOffer
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		asksCount: 0,
		bidsCount: 0,
		askIds:    make(map[string]int),
		bidIds:    make(map[string]int),
		asks:      make([]askOffer, 0),
		bids:      make([]bidOffer, 0),
	}
}

func (ob *OrderBook) UpdateAsks(asksToBeUpdated []Ask) error {
	for _, ask := range asksToBeUpdated {
		price := ask[0]
		quantity := ask[1]
		askId, ok := ob.askIds[price]
		if quantity == "0.000" {
			delete(ob.askIds, price)
			if askId != 0 { // if present in asks
				slices.Delete(ob.asks, askId, askId+1) // ?
			}
			return nil
		}
		if !ok {
			ob.askIds[price] = ob.asksCount
			ob.asksCount++
			floatPrice, err := strconv.ParseFloat(price, 32)
			if err != nil {
				return err
			}
			ob.asks = append(ob.asks, askOffer{floatPrice, quantity})
		}
	}
	ob.maintainAsksOrder()
	return nil
}

func (ob *OrderBook) UpdateBids(bidsToBeUpdated []Bid) error {
	for _, bid := range bidsToBeUpdated {
		price := bid[0]
		quantity := bid[1]
		bidId, ok := ob.bidIds[price]
		if quantity == "0.000" {
			delete(ob.bidIds, price)
			if bidId != 0 { // if present in bids
				slices.Delete(ob.bids, bidId, bidId+1) // ?
			}
			return nil
		}
		if !ok {
			ob.bidIds[price] = ob.bidsCount
			ob.bidsCount++
			floatPrice, err := strconv.ParseFloat(price, 32)
			if err != nil {
				return err
			}
			ob.bids = append(ob.bids, bidOffer{floatPrice, quantity})
		}
	}
	ob.maintainBidsOrder()
	return nil
}

func (ob *OrderBook) maintainAsksOrder() { // lowest ask is first
	sort.Slice(ob.asks, func(i, j int) bool {
		return ob.asks[i].price < ob.asks[j].price
	})
}

func (ob *OrderBook) maintainBidsOrder() { // highest bid is first
	sort.Slice(ob.bids, func(i, j int) bool {
		return ob.bids[i].price > ob.bids[j].price
	})
}
