package service

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/core"
	"github.com/kosdirus/cryptool/internal/storage/psql"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// BinanceAPISchedule used to coordinate app to run BinanceAPI function every 30 minutes.
// If time is not hh:30 or hh:00 at the moment of starting app - it will wait and then will
// start new ticker to run API collection func every 30 minutes.
//
// Such period of 30 minutes is used because in this app the lowest timeframe used for candlesticks
// received from Binance is 30 minutes. All timeframes you can find in initdata.TimeframeMap.
func BinanceAPISchedule(pgdb *pg.DB, BinanceAPIrun *uint32) {
	log.Println("Binance API schedule started:")
	for time.Now().Minute()%30 != 0 {
		time.Sleep(4 * time.Second)
	}
	switch time.Now().Second() {
	case 0:
		time.Sleep(3 * time.Second)
	case 1:
		time.Sleep(2 * time.Second)
	case 2:
		time.Sleep(1 * time.Second)
	}

	go BinanceAPI(pgdb, BinanceAPIrun)
	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		for range ticker.C {
			BinanceAPI(pgdb, BinanceAPIrun)
		}
	}()
}

// Struct merged contains fields needed to generate http request to Binance API.
// (look at BinanceOneAPI)
type merged struct {
	symb     string
	tf       string
	tfint    int64
	opentime int64
}

// mergeSlice checks the most recent candlesticks in database and generate all missing
// candlesticks between last candlestick (for each trade pair and timeframe) and the point
// in time when the function is executed (now).
//
// Be careful using it after long pause (e.g. week) of app running, because it may lead
// to calling dozens of thousands of goroutines.
func mergeSlice(pgdb *pg.DB) (mergedSlice []merged, err error) {
	allLastCandles, err := psql.GetAllLastCandles(pgdb)
	if err != nil {
		log.Println("mergeSlice error:  ", err)
		return nil, err
	}
	mx := sync.Mutex{}
	for _, a := range allLastCandles {
		for openTime := a.OpenTime; time.Now().UnixMilli()-openTime > a.TimeframeInt*60000*2; openTime = openTime +
			a.TimeframeInt*60000 {
			if err != nil {
				log.Println("binance_api.go line 73:", err, a.Coin, a.Timeframe)
				continue
			}
			mx.Lock()
			mergedSlice = append(mergedSlice, merged{
				symb:     a.Coin,
				tf:       a.Timeframe,
				tfint:    a.TimeframeInt,
				opentime: openTime + a.TimeframeInt*60000,
			})
			mx.Unlock()
		}
	}

	return mergedSlice, nil
}

// Func coordinateBinanceOneAPI receives length of merged slice (total number of candles to receive through API)
// from BinanceAPI and through ch channel allows to make requests a certain amount of times per minute. After
// receiving through resCh channel responses from goroutines about success http requests, it resets ticker
// time to call next batch of requests after this time.
//
// Binance provides with limits of 1200 request weight every minute from IP address (request
// for 1 candle has weight of 1).
func coordinateBinanceOneAPI(length int, ch chan<- struct{}, doneCh <-chan struct{}, resCh <-chan struct{}) {
	ticker := time.NewTicker(3 * time.Minute)
	const lenglimit = 1190
	var leng int
	proc := func() {
		if length >= lenglimit {
			leng = lenglimit
			length -= lenglimit
		} else {
			leng = length
			length = 0
		}
		for i := 0; i < leng; i++ {
			select {
			case ch <- struct{}{}:
			}
			time.Sleep(3 * time.Millisecond)
		}
		for i := 0; i < leng; i++ {
			select {
			case <-resCh:
			}
		}
		ticker.Reset(61 * time.Second)
	}
	proc()

	for {
		select {
		case <-ticker.C:
			proc()
		case <-doneCh:
			log.Println("doneCh received, coordinateBinanceOneAPI finished!")
			return
		}
	}
}

// BinanceAPI generates list (slice) of candles needed to be downloaded and added to database.
// It checks for other BinanceAPI instance is running and returns if so. It calls goroutines
// for every value in mergerSlice and waits for all candles to be added to database. After this will
// happen it will clean and close channels and will call CheckDBIntegrity in new goroutine to check
// whole database for missing candles.
func BinanceAPI(pgdb *pg.DB, BinanceAPIrun *uint32) {
	if atomic.LoadUint32(BinanceAPIrun) == 1 {
		log.Println("Another BinanceAPI instance is running.")
		return
	} else {
		atomic.SwapUint32(BinanceAPIrun, 1)
	}
	defer atomic.SwapUint32(BinanceAPIrun, 0)

	fmt.Println("BinanceAPI start", time.Now().UTC())
	candlech := make(chan core.Candle, 15)
	ch := make(chan struct{})
	resch := make(chan struct{}, 500)
	donech := make(chan struct{}, 1)
	var wg sync.WaitGroup
	go createCandleCheckForExistsInternal(pgdb, candlech)
	tt := time.Now()
	var nCoins, nCandles uint64

	tmerged := time.Now()
	mergedSlice, err := mergeSlice(pgdb)
	if err != nil {
		log.Println(err)
	}

	log.Println("Time spent on merge slice:", time.Since(tmerged))
	tbin := time.Now()
	log.Println("START OF CALLING GOROUTINES", tbin.Format(time.StampMicro))
	wg.Add(len(mergedSlice))

	tformerger := time.Now()
	for _, ms := range mergedSlice {
		go func(ms merged) {
			BinanceOneAPI(ms, &wg, candlech, ch, resch)
		}(ms)
		atomic.AddUint64(&nCandles, 1)

	}
	log.Println("Time for calling all goroutines: ", time.Since(tformerger), len(mergedSlice))
	go coordinateBinanceOneAPI(len(mergedSlice), ch, donech, resch)

	atomic.AddUint64(&nCoins, 1)

	wg.Wait()
	log.Println("time for all binance api", time.Since(tbin))
	close(candlech)
	donech <- struct{}{}
	close(donech)
	for len(ch) != 0 {
		<-ch
	}
	close(ch)
	fmt.Println("BinanceAPI finish at", time.Now().UTC(), ".", nCoins, "coins (trade pairs) checked and",
		nCandles/nCoins, "timeframes updated. Total number of fetched candles:", nCandles,
		". Time spent:", time.Since(tt))
	fmt.Println("CheckDBIntegrity started.")
	go CheckDBIntegrity(pgdb)
}

// BinanceOneAPI called in goroutine with specific trade pair and timeframe to request candle from Binance. It waits
// signal from ch channel to send request to Binance, and after receiving response - it sends signal to resCh to
// inform coordinating func that request is finished.
//
// Func reads response body, unmarshal json and convert it to internal core.Candle struct, and lastly sends this candle
// via candleCh channel to func that is responsible for adding this candle to the database.
func BinanceOneAPI(ms merged, wg *sync.WaitGroup, candleCh chan<- core.Candle, allowCh <-chan struct{}, resCh chan<- struct{}) error {
	defer wg.Done()
	url := fmt.Sprintf("https://www.binance.com/api/v3/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d",
		ms.symb, ms.tf, ms.opentime, ms.opentime+(ms.tfint*60000-1))
	<-allowCh
	t := time.Now()
	resp, err := http.Get(url)
	tt := time.Since(t)
	resCh <- struct{}{}

	log.Println(ms, t.Format(time.StampMicro), tt)

	if err != nil && resp.StatusCode != 429 {
		return err
	} else if resp.StatusCode == 429 {
		fmt.Println("response code 429, sending pause signal to channel, sleeping 70seconds. Goroutine nums:",
			runtime.NumGoroutine())
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var value [][]interface{}
	err = json.Unmarshal(body, &value)
	if err != nil {
		log.Println(err)
	}

	if len(value) != 0 { // && value[0][0] != 0 (open_time in value should not be 0)
		c := ConvertAPItoCandleStruct(ms.symb, ms.tf, value[0])
		candleCh <- c
	}
	return nil
}

// createCandleCheckForExistsInternal is responsible for receiving candles via channel and add them to
// postgres database.
func createCandleCheckForExistsInternal(db *pg.DB, candleCh <-chan core.Candle) {
	for req := range candleCh {
		_, err := db.Model(&req).
			Where("candle.my_id = ?", req.MyID).
			SelectOrInsert()
		if err != nil {
			log.Println(err)
			return
		}
	}
}
