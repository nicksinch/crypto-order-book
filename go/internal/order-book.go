package internal

import (
	"sort"
	"strconv"
)

type bid [2]string // price level to be updated; quantity
type ask [2]string // price level to be updated; quantity

type OrderBook struct {
	asks map[string]float64
	bids map[string]float64
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		asks: make(map[string]float64),
		bids: make(map[string]float64),
	}
}

// UpdateAsks updates the asks map with the new update events and returns the current best ask along with 10th level ask
func (ob *OrderBook) UpdateAsks(asksToBeUpdated []ask) (float64, float64, error) {
	for _, ask := range asksToBeUpdated {
		price := ask[0]
		quantity := ask[1]
		if quantity == "0.000" {
			delete(ob.asks, price)
			continue
		}
		_, ok := ob.asks[price]
		if !ok {
			floatQuantity, err := strconv.ParseFloat(quantity, 32)
			if err != nil {
				return 0, 0, err
			}
			ob.asks[price] = floatQuantity
		}
	}
	pricesSorted, err := sortPrices(ob.asks, false)
	if err != nil {
		return 0, 0, err
	}
	bestAsk := (*pricesSorted)[0]
	var tenthLevelAsk float64
	if len(*pricesSorted) > 9 {
		tenthLevelAsk = (*pricesSorted)[9]
	}

	return bestAsk, tenthLevelAsk, nil
}

// UpdateBids updates the bids map with the new update events and returns the current best bid along with 10th level bid
func (ob *OrderBook) UpdateBids(bidsToBeUpdated []bid) (float64, float64, error) {
	for _, bid := range bidsToBeUpdated {
		price := bid[0]
		quantity := bid[1]
		if quantity == "0.000" {
			delete(ob.bids, price)
			continue
		}
		_, ok := ob.bids[price]
		if !ok {
			floatQuantity, err := strconv.ParseFloat(quantity, 32)
			if err != nil {
				return 0, 0, err
			}
			ob.bids[price] = floatQuantity
		}
	}
	pricesSorted, err := sortPrices(ob.bids, true)
	if err != nil {
		return 0, 0, err
	}
	bestBid := (*pricesSorted)[0]
	var tenthLevelBid float64
	if len(*pricesSorted) > 9 {
		tenthLevelBid = (*pricesSorted)[9]
	}

	return bestBid, tenthLevelBid, nil
}

// sortPrices sorts ask/bid map by key (prices). bidOrAsk value is true for bid, false for ask
func sortPrices(prices map[string]float64, bidOrAsk bool) (*[]float64, error) {
	keys := make([]float64, 0, len(prices))

	for price, _ := range prices {
		floatPrice, err := strconv.ParseFloat(price, 32)
		if err != nil {
			return nil, err
		}
		keys = append(keys, floatPrice)
	}
	if bidOrAsk {
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] > keys[j]
		})
	} else {
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})
	}
	return &keys, nil
}
