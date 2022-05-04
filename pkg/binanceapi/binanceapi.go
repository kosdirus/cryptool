package binanceapi

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/cmd/interal/candle"
	"github.com/kosdirus/cryptool/cmd/interal/symbol"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func BinanceAPISchedule(pgdb *pg.DB) {
	log.Println("binance API schedule started")
	for time.Now().Minute()%30 != 0 {
		time.Sleep(5 * time.Second)
	}
	switch time.Now().Second() {
	case 0:
		time.Sleep(4 * time.Second)
	case 1:
		time.Sleep(3 * time.Second)
	case 2:
		time.Sleep(2 * time.Second)
	case 3:
		time.Sleep(1 * time.Second)
	}
	go BinanceAPI(pgdb)
	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				BinanceAPI(pgdb)
			}
		}
	}()
}

func BinanceAPI(pgdb *pg.DB) {
	fmt.Println("BinanceAPI start", time.Now().UTC())
	tt := time.Now()
	var nCoins, nCandles uint64
	var goroutines int
	if os.Getenv("ENV") == "DIGITAL" {
		goroutines = 7
		log.Println(runtime.NumCPU(), "CPUs")
	} else {
		goroutines = 5
	}
	totalCoins := len(symbol.SymbolList)
	lastGoroutine := goroutines - 1
	stride := totalCoins / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)

	/*var mergedSlice []string
	for _, v := range symbol.SymbolList[start:end] {
		for k1, v1 := range symbol.Timeframe {

		}
	}*/

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			start := g * stride
			end := start + stride
			if g == lastGoroutine {
				end = totalCoins
			}

			for _, v := range symbol.SymbolList[start:end] {
				for k1, v1 := range symbol.TimeframeMap {
					for openTime, err := candle.GetLastCandle(pgdb, v, v1); time.Now().UnixMilli()-openTime > int64(k1*60000*2); {
						if err != nil {
							log.Println("binanceapi.go line 41:", err, v, v1)
							continue
						}
						openTime = openTime + int64(k1*60000)

						err = BinanceOneAPI(pgdb, v, v1, int64(k1), openTime)
						if err != nil {
							log.Println("binanceapi.go line 47:", err, v, v1, openTime)
							continue
						}
						atomic.AddUint64(&nCandles, 1)
					}
				}
				atomic.AddUint64(&nCoins, 1)
			}
			wg.Done()
		}(g)
	}
	wg.Wait()

	fmt.Println("BinanceAPI finish at", time.Now().UTC(), ".", nCoins, "coins (trade pairs) checked and", nCandles/nCoins, "timeframes updated. Total number of fetched candles:",
		nCandles, ". Time spent:", time.Since(tt))
	fmt.Println("CheckDBIntegrity started.")
	go CheckDBIntegrity(pgdb)
}

func BinanceOneAPI(pgdb *pg.DB, symbol, timeframe string, k1, openTime int64) error {

	url := fmt.Sprintf("https://www.binance.com/api/v3/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d", symbol, timeframe, openTime, openTime+(k1*60000-1))
	t := time.Now()
	resp, err := http.Get(url)
	tt := time.Since(t)
	if err != nil && resp.StatusCode != 429 {
		return err
	} else if resp.StatusCode == 429 {
		fmt.Println("response code 429, sleeping 4mins")
		time.Sleep(4 * time.Minute)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var test candle.TestAPIStruct
	err = json.Unmarshal(body, &test)

	if len(test) != 0 {
		c := candle.ConvertAPItoCandleStruct(symbol, timeframe, test[0])

		_, err = candle.CreateCandleCheckForExists(pgdb, &c)
		if err != nil {
			return err
		}

		log.Println("BinanceOneAPI ", symbol, timeframe, time.UnixMilli(openTime).UTC(), "time for GET", tt)
	}

	return nil
}
