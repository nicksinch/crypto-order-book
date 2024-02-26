package internal

import (
	"sort"
	"strconv"
)

type bid [2]string // price level to be updated; quantity
type ask [2]string // price level to be updated; quantity

type priceLevel struct {
	price    float64
	quantity float64
}

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

// UpdateAsks updates the asks map with the new update events and returns
// the current best ask, 10th level ask and the ask price for 1 unit
func (ob *OrderBook) UpdateAsks(asksToBeUpdated []ask) (float64, float64, float64, error) {
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
				return 0, 0, 0, err
			}
			ob.asks[price] = floatQuantity
		}
	}
	pricesSorted, err := sortPrices(ob.asks, false)
	if err != nil {
		return 0, 0, 0, err
	}
	bestAsk := (*pricesSorted)[0].price
	var tenthLevelAsk float64
	if len(*pricesSorted) > 9 {
		tenthLevelAsk = (*pricesSorted)[9].price
	}

	askOneUnit := calculatePriceOneUnit(pricesSorted)

	return bestAsk, tenthLevelAsk, askOneUnit, nil
}

// UpdateBids updates the bids map with the new update events and returns
// the current best bid, 10th level bid and the bid price for 1 unit
func (ob *OrderBook) UpdateBids(bidsToBeUpdated []bid) (float64, float64, float64, error) {
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
				return 0, 0, 0, err
			}
			ob.bids[price] = floatQuantity
		}
	}
	pricesSorted, err := sortPrices(ob.bids, true)
	if err != nil {
		return 0, 0, 0, err
	}
	bestBid := (*pricesSorted)[0].price
	var tenthLevelBid float64
	if len(*pricesSorted) > 9 {
		tenthLevelBid = (*pricesSorted)[9].price
	}

	bidOneUnit := calculatePriceOneUnit(pricesSorted)

	return bestBid, tenthLevelBid, bidOneUnit, nil
}

// sortPrices sorts ask/bid map by key (prices). bidOrAsk value is true for bid, false for ask.
// Returns a slice of priceLevel struct which also contains the quantity for the corresponding price
func sortPrices(bidsOrAsks map[string]float64, bidOrAsk bool) (*[]priceLevel, error) {
	priceLevelsSorted := make([]priceLevel, 0, len(bidsOrAsks))

	for price, quantity := range bidsOrAsks {
		floatPrice, err := strconv.ParseFloat(price, 32)
		if err != nil {
			return nil, err
		}
		priceLevelsSorted = append(priceLevelsSorted, priceLevel{floatPrice, quantity})
	}
	if bidOrAsk {
		sort.Slice(priceLevelsSorted, func(i, j int) bool {
			return priceLevelsSorted[i].price > priceLevelsSorted[j].price
		})
	} else {
		sort.Slice(priceLevelsSorted, func(i, j int) bool {
			return priceLevelsSorted[i].price < priceLevelsSorted[j].price
		})
	}
	return &priceLevelsSorted, nil
}

// calculatePriceOneUnit calculates the price to buy or sell 1 unit of the underlying symbol
func calculatePriceOneUnit(priceLevels *[]priceLevel) float64 {
	var totalPrice float64
	var totalQuantity float64
	offersCount := 0.0
	for _, level := range *priceLevels {
		totalPrice += level.price
		totalQuantity += level.quantity
		offersCount += 1
		if totalQuantity >= 1.00 {
			return totalPrice / offersCount
		}
	}
	return totalPrice / offersCount
}
