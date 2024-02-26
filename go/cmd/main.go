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

// connectToWebsocket connects to a single trading-pair websocket url
func connectToWebsocket(u string, s string, shutdown chan struct{}, wg *sync.WaitGroup) {
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

	tpSymbol, ok := strings.CutSuffix(tpSymbolAfter, "@depth@100ms")
	if !ok {
		slog.Error("Couldn't cut suffix to take trading pair symbol")
	}

	eventHandler = internal.InitializeHandler(tpSymbol, s)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			err = eventHandler.HandleUpdate(message)
			if err != nil {
				slog.Error("Error handling order book update", slog.Any("error", err.Error()))
			}
		}
	}()

	tickerMidPoint := time.NewTicker(time.Second)
	defer tickerMidPoint.Stop()

	tickerSma := time.NewTicker(60 * time.Second)
	defer tickerSma.Stop()

	simpleMovingAverageSum := 0.0
	secondsCount := 0

	go func() {
		for {
			select {
			case <-tickerMidPoint.C:
				bestAsk := eventHandler.GetBestAsk()
				bestBid := eventHandler.GetBestBid()
				midPointPrice := (bestAsk - bestBid) / 2
				simpleMovingAverageSum += midPointPrice
				secondsCount++
			case <-tickerSma.C:
				slog.Info("Simple Moving Average (SMA) for the last 60 seconds: ", slog.Float64("SMA", simpleMovingAverageSum/60), slog.String("Symbol", strings.ToUpper(tpSymbol)))
				simpleMovingAverageSum = 0.0
				secondsCount = 0
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
			slog.Debug(fmt.Sprintf("%v sending a unsolicited pong frame.", time.Now().Format(time.RFC3339)))
			err := c.WriteMessage(websocket.PongMessage, nil)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-shutdown:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
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
		go connectToWebsocket(u, snapshotUrls[idx], shutdown, &wg)
	}
	wg.Wait()
	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("Error running crypto order book service", slog.Any("error", err))
	}
}
