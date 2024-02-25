package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	tradingpairs "crypto-order-book/go/internal/trading-pairs"
)

const (
	snapshotUrlBtcUsdt = "https://fapi.binance.com/fapi/v1/depth?symbol=BTCUSDT&limit=1000"
	snapshotUrlEthUsdt = "https://fapi.binance.com/fapi/v1/depth?symbol=ETHUSDT&limit=1000"
)

// TODO: Ensure storage layout
type Handler struct {
	snapshotLastUpdateId     int64
	previousEventFinalUpdate int
	tradingPair              tradingpairs.TradingPair
	snapshotUrl              string
	firstEventProcessed      bool
}

func InitializeHandler(tpSymbol string) *Handler {
	var snapshotUrl string
	var tradingPair tradingpairs.TradingPair
	switch tpSymbol {
	case "btcusdt":
		snapshotUrl = snapshotUrlBtcUsdt
		tradingPair = tradingpairs.NewBtcUsdtPair()
	case "ethusdt":
		snapshotUrl = snapshotUrlEthUsdt
		tradingPair = tradingpairs.NewEthUsdtPair()
	}
	snapshot := GetDepthSnapshot(snapshotUrl)
	log.Println(fmt.Sprintf("Snapshot LastUpdateId for trading pair %s = %d \n", tpSymbol, snapshot.LastUpdateId))
	return &Handler{
		snapshotLastUpdateId: snapshot.LastUpdateId,
		tradingPair:          tradingPair,
		snapshotUrl:          snapshotUrl,
		firstEventProcessed:  false,
	}
}

func (h *Handler) HandleUpdate(message []byte) error {
	parseDepthUpdateEvent := DepthUpdateEvent{}
	if err := json.Unmarshal(message, &parseDepthUpdateEvent); err != nil {
		return err
	}
	if parseDepthUpdateEvent.Data.FinalUpdateId < h.snapshotLastUpdateId {
		slog.Info(fmt.Sprintf("Discarding event with smaller last update id...\n"))
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

	//indentEventRecord, _ := json.MarshalIndent(parseDepthUpdateEvent, "", " ")
	//n, err := f.Write(append(indentEventRecord, byte(',')))
	//if err != nil {
	//	log.Fatal("Error writing bytes ", err)
	//}
	//log.Println(fmt.Sprintf("wrote %d bytes\n", n))
	//f.Sync()

	h.tradingPair.UpdateAsks(parseDepthUpdateEvent.Data.AsksToBeUpdated)
	h.tradingPair.UpdateBids(parseDepthUpdateEvent.Data.BidsToBeUpdated)

	h.previousEventFinalUpdate = int(parseDepthUpdateEvent.Data.FinalUpdateId)

	return nil
}

func (h *Handler) reinitializeSnapshot() {
	h.snapshotLastUpdateId = GetDepthSnapshot(h.snapshotUrl).LastUpdateId
	slog.Debug(fmt.Sprintf("Snapshot LastUpdateId for trading pair %s UPDATED. \n", h.tradingPair.Symbol()),
		slog.Int64("NewId", h.snapshotLastUpdateId))
}
