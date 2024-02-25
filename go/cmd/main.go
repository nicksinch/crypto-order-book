package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"crypto-order-book/go/internal"

	"github.com/gorilla/websocket"
)

const (
	btcUsdtWsEndpoint   = "wss://fstream.binance.com/stream?streams=btcusdt@depth@100ms"
	ethUsdtWsEndpoint   = "wss://fstream.binance.com/stream?streams=ethusdt@depth@100ms"
	depthEventsJsonPath = "/Users/nickkirov/GolandProjects/reflect-playground/cmd/order-book/binance-result.json"
	prefixFStreamApi    = "wss://fstream.binance.com/stream?streams="
)

// connectToWebsocket connects to a single trading-pair websocket url
func connectToWebsocket(u string, shutdown chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("connecting to %s", u)
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

	eventHandler = internal.InitializeHandler(tpSymbol)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			err = eventHandler.HandleUpdate(message)
			if err != nil {
				slog.Error("Error handling order book update", slog.Any("error", err.Error()))
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
			log.Printf("%v sending a unsolicited pong frame.", time.Now().Format(time.RFC3339))
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
	endpoints := []string{btcUsdtWsEndpoint, ethUsdtWsEndpoint}

	shutdown := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		log.Println("interrupt")
		close(shutdown)
	}()

	var wg sync.WaitGroup
	for _, u := range endpoints {
		// where elements are URLs
		// of endpoints to connect to.
		wg.Add(1)
		go connectToWebsocket(u, shutdown, &wg)
	}
	wg.Wait()
	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("Error running crypto order book service", slog.String("error", err.Error()))
	}
}
