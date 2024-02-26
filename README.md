# Crypto Order Book

## Run As Docker Container

```bash 
chmod +x ./build-and-run.sh
```

```bash 
./build-and-run.sh
```

## Overview
- Supports client scaling by making use of goroutines. 
  - A separate goroutine is spawned inside a WaitGroup for each websocket connection URL (each trading-pair)
- Supports many trading pairs (`BTC/USDT` and `ETH/USDT` are default)
  - To add a new trading-pair, simply provide a websocket URL with its corresponding snapshot URL in `websockets-config.json`
