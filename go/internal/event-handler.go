package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	snapshotLastUpdateId     int64
	previousEventFinalUpdate int
	tradingPairOb            *OrderBook
	snapshotUrl              string
	pairSymbol               string
	firstEventProcessed      bool
}

func InitializeHandler(tpSymbol string, snapshotUrl string) *Handler {
	snapshot, err := getDepthSnapshot(snapshotUrl)
	if err != nil {
		return nil
	}
	slog.Info("Snapshot taken", slog.String("Symbol", strings.ToUpper(tpSymbol)), slog.Int64("LastUpdateId", snapshot.LastUpdateId))
	return &Handler{
		snapshotLastUpdateId: snapshot.LastUpdateId,
		tradingPairOb:        NewOrderBook(),
		snapshotUrl:          snapshotUrl,
		firstEventProcessed:  false,
		pairSymbol:           tpSymbol,
	}
}

// HandleUpdate handles a single `depthUpdate` event maintaining the correct order book invariant
// and calculates the values needed for indicators
func (h *Handler) HandleUpdate(message []byte, bestAskChan chan float64, bestBidChan chan float64, secondsChan *time.Ticker) error {
	parseDepthUpdateEvent := depthUpdateEvent{}
	if err := json.Unmarshal(message, &parseDepthUpdateEvent); err != nil {
		return err
	}
	if parseDepthUpdateEvent.Data.FinalUpdateId < h.snapshotLastUpdateId {
		slog.Debug("Discarding event with smaller last update id...")
		return nil
	}
	if (int(parseDepthUpdateEvent.Data.FinalUpdateIdLastStream) != h.previousEventFinalUpdate) && h.firstEventProcessed {
		slog.LogAttrs(context.Background(),
			slog.LevelDebug,
			"pu of current event not equal to u of last event. Reinitializing local order book managing...\n")
		err := h.reinitializeSnapshot()
		if err != nil {
			return err
		}
		return nil
	}
	if !h.firstEventProcessed {
		h.firstEventProcessed = true
	}

	bestAsk, tenthLevelAsk, oneUnitAsk, err := h.tradingPairOb.UpdateAsks(parseDepthUpdateEvent.Data.AsksToBeUpdated)
	if err != nil {
		return err
	}
	bestBid, tenthLevelBid, oneUnitBid, err := h.tradingPairOb.UpdateBids(parseDepthUpdateEvent.Data.BidsToBeUpdated)
	if err != nil {
		return err
	}

	asset, _, _ := strings.Cut(h.pairSymbol, "usdt")
	slog.Info(fmt.Sprintf("BUY 1 %s %v | SELL 1 %s %v", strings.ToUpper(asset), oneUnitBid, strings.ToUpper(asset), oneUnitAsk))
	slog.Info("Spread between 10th order book level", slog.Float64("Value", tenthLevelAsk-tenthLevelBid), slog.String("Symbol", strings.ToUpper(h.pairSymbol)))

	select {
	case <-secondsChan.C:
		bestAskChan <- bestAsk
		bestBidChan <- bestBid
	default:
	}

	h.previousEventFinalUpdate = int(parseDepthUpdateEvent.Data.FinalUpdateId)

	return nil
}

func (h *Handler) reinitializeSnapshot() error {
	snapshot, err := getDepthSnapshot(h.snapshotUrl)
	if err != nil {
		return err
	}
	h.snapshotLastUpdateId = snapshot.LastUpdateId
	slog.Info("Snapshot taken", slog.String("Symbol", h.pairSymbol),
		slog.Int64("LastUpdateId", h.snapshotLastUpdateId))
	return nil
}

func getDepthSnapshot(snapshotUrl string) (*depthSnapshot, error) {
	resp, err := http.Get(snapshotUrl)
	if err != nil {
		slog.Error("Error getting snapshot", slog.String("snapshotUrl", snapshotUrl))
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading snapshot response", slog.String("snapshotUrl", snapshotUrl))
		return nil, err
	}
	snapshot := depthSnapshot{}
	err = json.Unmarshal(body, &snapshot)
	if err != nil {
		slog.Error("Error unmarshalling depth snapshot", slog.String("snapshotUrl", snapshotUrl))
		return nil, err
	}
	return &snapshot, nil
}
