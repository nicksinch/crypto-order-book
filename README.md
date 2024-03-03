# Crypto Order Book

## Run As Docker Container

```bash 
chmod +x ./build-and-run.sh
```

```bash 
./build-and-run.sh
```

## Technical Details
- Supports client scaling by making use of goroutines 
  - A separate goroutine is spawned inside a WaitGroup for each websocket connection URL (each trading-pair)
  - Each separate trading pair has it's `own order book`
- Supports many trading pairs (`BTC/USDT` and `ETH/USDT` are default)
  - To add a new trading-pair, simply provide a websocket URL with its corresponding snapshot URL in `websockets-config.json`
- The data structure of the order book is implemented as `two hash-maps`, one for the asks and one for the bids
  - Keys are the level prices and values are the quantity of the corresponding price level
  - This enables `amortized` O(1) for look-ups, adds, deletions and updates
  - Space complexity in this case is O(2n) = O(n) where n is the number of price levels in the order book
  - the hash-maps constantly grow and shrink as some of the keys are being deleted if the quantity becomes 0
- For the calculation of each indicator, the price levels for the order book are needed `in the correct sorted manner`:
  - To achieve this, after each order book update, all price levels of the two maps are taken and stored in a slice => O(n)
  - Secondly, a sort for each map is performed on the slice => O(nlogn) on average, where n is the number of bids/asks in the ob
  - Finally, best bid, best ask and 10th levels are obtained in O(1) due to the sorted arrays
  - Total time complexity to be able to calculate the 3 indicators at the desired interval is: O(m) + O(n) + O(nlogn) = O(nlogn) since n > m,
  where m is the number of asks/bids to be updated, n is the total number of asks/bids in the hash-map
  - To calculate SMA and EWMA, a goroutine is created and is listening at the desired intervals to perform the calculations
  using the already computed values from the last step

## Further Improvements
- the calculation of the indicators can be delegated to a IndicatorManager so that integration of other indicators can be done easily
- Understanding the data coming from Binance on a deeper level and based on that:
  - benchmark performance and experiment with different sorting strategies and data structures to improve speed trading-off some memory and vice versa
  - Such strategies may involve some more sophisticated data structures such as:
    - Min/Max Heaps/Fibonacci Heaps (although there will be some challenges there due to the nature of this task)
    - AVL self-balancing binary search trees to perform some of the required tasks if we observe tree balancing issues


## Resources:
- https://binance-docs.github.io/apidocs/futures/en/#partial-book-depth-streams
- https://stanford.edu/~ashlearn/RLForFinanceBook/chapter9.pdf
- https://github.com/sammchardy/python-binance/blob/master/binance/depthcache.py
- https://dev.binance.vision/t/how-to-manage-a-local-order-book/9207
- https://github.com/wisespace-io/binance-rs/issues/6
- https://github.com/binance/binance-spot-api-docs/blob/master/web-socket-streams.md#how-to-manage-a-local-order-book-correctly
- https://web.archive.org/web/20110219163448/http://howtohft.wordpress.com/2011/02/15/how-to-build-a-fast-limit-order-book/