package internal

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// TODO: Ensure storage layout
type Handler struct {
	snapshotLastUpdateId     int64
	previousEventFinalUpdate int
	tradingPairOb            *OrderBook
	snapshotUrl              string
	pairSymbol               string
	firstEventProcessed      bool
	bestBid                  float64
	bestAsk                  float64
}

func InitializeHandler(tpSymbol string, snapshotUrl string) *Handler {
	snapshot := getDepthSnapshot(snapshotUrl)
	slog.Info("Snapshot taken", slog.String("Symbol", strings.ToUpper(tpSymbol)), slog.Int64("LastUpdateId", snapshot.LastUpdateId))
	return &Handler{
		snapshotLastUpdateId: snapshot.LastUpdateId,
		tradingPairOb:        NewOrderBook(),
		snapshotUrl:          snapshotUrl,
		firstEventProcessed:  false,
		pairSymbol:           tpSymbol,
	}
}

func (h *Handler) HandleUpdate(message []byte) error {
	parseDepthUpdateEvent := DepthUpdateEvent{}
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
		h.reinitializeSnapshot()
		return nil // TODO: Is this the correct behavior ?
	}
	if !h.firstEventProcessed {
		h.previousEventFinalUpdate = int(parseDepthUpdateEvent.Data.FinalUpdateId)
		h.firstEventProcessed = true
	}

	bestAsk, tenthLevelAsk, err := h.tradingPairOb.UpdateAsks(parseDepthUpdateEvent.Data.AsksToBeUpdated)
	if err != nil {
		return err
	}
	bestBid, tenthLevelBid, err := h.tradingPairOb.UpdateBids(parseDepthUpdateEvent.Data.BidsToBeUpdated)
	if err != nil {
		return err
	}

	slog.Info("Spread between 10th order book level", slog.Float64("Value", tenthLevelAsk-tenthLevelBid), slog.String("Symbol", strings.ToUpper(h.pairSymbol)))

	h.bestBid = bestBid
	h.bestAsk = bestAsk

	h.previousEventFinalUpdate = int(parseDepthUpdateEvent.Data.FinalUpdateId)

	return nil
}

func (h *Handler) GetBestBid() float64 {
	return h.bestBid
}

func (h *Handler) GetBestAsk() float64 {
	return h.bestAsk
}

func (h *Handler) reinitializeSnapshot() {
	h.snapshotLastUpdateId = getDepthSnapshot(h.snapshotUrl).LastUpdateId
	slog.Info("Snapshot taken", slog.String("Symbol", h.pairSymbol),
		slog.Int64("LastUpdateId", h.snapshotLastUpdateId))
}

func getDepthSnapshot(snapshotUrl string) *DepthSnapshot {
	resp, err := http.Get(snapshotUrl)
	if err != nil {
		slog.Error("Error getting snapshot", slog.String("snapshotUrl", snapshotUrl))
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading snapshot response", slog.String("snapshotUrl", snapshotUrl))
		return nil
	}
	snapshot := DepthSnapshot{}
	err = json.Unmarshal(body, &snapshot)
	if err != nil {
		slog.Error("Error unmarshalling depth snapshot", slog.String("snapshotUrl", snapshotUrl))
		return nil
	}
	return &snapshot
}
