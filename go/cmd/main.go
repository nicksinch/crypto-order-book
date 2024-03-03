package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"crypto-order-book/go/config"
	"crypto-order-book/go/internal"

	"github.com/gorilla/websocket"
)

const (
	prefixFStreamApi = "wss://fstream.binance.com/stream?streams="
)

// subscribeToTradingPairWs connects to a single trading-pair websocket url and handles event updates
func subscribeToTradingPairWs(u string, s string, shutdown chan struct{}, depth string, wg *sync.WaitGroup) {
	defer wg.Done()

	slog.Info(fmt.Sprintf("connecting to %s", u))
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	var eventHandler *internal.Handler

	tpSymbolAfter, ok := strings.CutPrefix(u, prefixFStreamApi)
	if !ok {
		slog.Error("Couldn't cut prefix to take trading pair symbol")
	}

	tpSymbol, ok := strings.CutSuffix(tpSymbolAfter, "@depth@"+depth) // depth ms is configurable
	if !ok {
		slog.Error("Couldn't cut suffix to take trading pair symbol")
	}

	eventHandler = internal.InitializeHandler(tpSymbol, s)

	bestAskChan := make(chan float64, 1)
	bestBidChan := make(chan float64, 1)
	defer close(bestBidChan)
	defer close(bestAskChan)

	tickerMidPoint := time.NewTicker(time.Second)
	defer tickerMidPoint.Stop()

	go func(bestAskChan chan float64, bestBidChan chan float64, secondsChan *time.Ticker) {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				slog.Error("error reading websocket message:", slog.Any("error", err))
				return
			}
			err = eventHandler.HandleUpdate(message, bestAskChan, bestBidChan, secondsChan)
			if err != nil {
				slog.Error("Error handling order book update", slog.Any("error", err.Error()))
			}
		}
	}(bestAskChan, bestBidChan, tickerMidPoint)

	tickerSmaEwma := time.NewTicker(60 * time.Second)
	defer tickerSmaEwma.Stop()

	// a goroutine to calculate the SMA and EWMA indicators as per the desired intervals
	go func() {
		simpleMovingAverageSum := 0.0
		secondsCount := 0

		// the alpha value for the exponentially moving average indicator
		alphaEwma := 2/time.Minute.Seconds() + 1

		// the aggregation of the ewma value for each second (will be printed when 60 seconds are reached)
		var ewmaValue float64

		previousEwmaValue := 0.0
		for {
			select {
			case <-tickerMidPoint.C:
				bestAsk := <-bestAskChan
				bestBid := <-bestBidChan
				midPointPrice := (bestAsk - bestBid) / 2
				simpleMovingAverageSum += midPointPrice
				secondsCount++
				ewmaValue += (alphaEwma * midPointPrice) + (1-alphaEwma)*previousEwmaValue
				previousEwmaValue = ewmaValue
			case <-tickerSmaEwma.C:
				slog.Info("Simple Moving Average (SMA) for the last 60 seconds: ", slog.Float64("SMA", simpleMovingAverageSum/60), slog.String("Symbol", strings.ToUpper(tpSymbol)))
				simpleMovingAverageSum = 0.0
				secondsCount = 0
				slog.Info("EWMA value for the last 60 seconds: ", slog.Float64("EWMA", ewmaValue), slog.String("Symbol", strings.ToUpper(tpSymbol)))
				ewmaValue = 0.0
				previousEwmaValue = 0.0
			}
		}
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			slog.Debug(fmt.Sprintf("%v sending a pong frame.", time.Now().Format(time.RFC3339)))
			err := c.WriteMessage(websocket.PongMessage, nil)
			if err != nil {
				slog.Error("error sending pong message:", slog.Any("error", err))
				return
			}
		case <-shutdown:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				slog.Error("error closing socket:", slog.Any("error", err))
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func run() error {
	wsConfig, err := config.InitializeConfig()
	if err != nil {
		return err
	}
	if len(wsConfig.Pairs) != len(wsConfig.Snapshots) {
		return errors.New("each websocket url needs to have a snapshot url")
	}

	wsEndpoints := wsConfig.Pairs
	snapshotUrls := wsConfig.Snapshots
	depth := wsConfig.Depth

	shutdown := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		slog.Info("interrupt...")
		close(shutdown)
	}()

	var wg sync.WaitGroup
	for idx, u := range wsEndpoints {
		// where elements are URLs
		// of endpoints to connect to.
		wg.Add(1)
		go subscribeToTradingPairWs(u, snapshotUrls[idx], shutdown, depth, &wg)
	}
	wg.Wait()
	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("Error running crypto order book service", slog.Any("error", err))
	}
}
