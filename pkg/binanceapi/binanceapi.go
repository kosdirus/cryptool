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

type mergedST struct {
	symb     string
	tf       string
	tfint    int64
	opentime int64
}

func getMergedSlice(pgdb *pg.DB, mergedSlice *[]mergedST) {
	var goroutines = 6
	totalCoins := len(symbol.SymbolList)
	lastGoroutine := goroutines - 1
	stride := totalCoins / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)

	mx := sync.Mutex{}

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			start := g * stride
			end := start + stride
			if g == lastGoroutine {
				end = totalCoins
			}

			for _, s := range symbol.SymbolList[start:end] {
				for tfint, tf := range symbol.TimeframeMap {
					for openTime, err := candle.GetLastCandle(pgdb, s, tf); time.Now().UnixMilli()-openTime > int64(tfint*60000*2); openTime = openTime + int64(tfint*60000) {
						if err != nil {
							log.Println("binanceapi.go line 41:", err, s, tf)
							continue
						}
						mx.Lock()
						*mergedSlice = append(*mergedSlice, mergedST{
							symb:     s,
							tf:       tf,
							tfint:    int64(tfint),
							opentime: openTime + int64(tfint*60000),
						})
						mx.Unlock()
					}
				}
			}
			wg.Done()
		}(g)
	}
	wg.Wait()
}

func coordinateBinanceOneAPI(ch chan<- struct{}, donech <-chan struct{}) {
	ticker := time.NewTicker(61 * time.Second)
	for i := 0; i < 1190; i++ {
		ch <- struct{}{}
	}
	time.Sleep(55 * time.Second)

	for {
		select {
		case <-ticker.C:
			for i := 0; i < 1190-len(ch); i++ {

				ch <- struct{}{}
			}
		case <-donech:
			log.Println("Donech received, coordinateBinanceOneAPI finish!!")
			return
		}
	}
}

func BinanceAPI(pgdb *pg.DB) {
	fmt.Println("BinanceAPI start", time.Now().UTC())
	candlech := make(chan candle.Candle, 15)
	ch := make(chan struct{}, 1190)
	donech := make(chan struct{}, 1)
	var wg sync.WaitGroup
	go createCandleCheckForExistsInternal(pgdb, candlech)
	go coordinateBinanceOneAPI(ch, donech)
	tt := time.Now()
	var nCoins, nCandles uint64

	var mergedSlice []mergedST
	tmerged := time.Now()
	getMergedSlice(pgdb, &mergedSlice)

	log.Println(mergedSlice)
	log.Println("Time spent on merge slice:", time.Since(tmerged))
	tbin := time.Now()

	wg.Add(len(mergedSlice))

	for _, ms := range mergedSlice {
		go func() {
			BinanceOneAPI(ms, &wg, candlech, ch)
		}()
		time.Sleep(7 * time.Millisecond)
		atomic.AddUint64(&nCandles, 1)

	}

	atomic.AddUint64(&nCandles, 1)
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
	fmt.Println("BinanceAPI finish at", time.Now().UTC(), ".", nCoins, "coins (trade pairs) checked and", nCandles/nCoins, "timeframes updated. Total number of fetched candles:",
		nCandles, ". Time spent:", time.Since(tt))
	fmt.Println("CheckDBIntegrity started.")
	go CheckDBIntegrity(pgdb)
}

func BinanceOneAPI(ms mergedST, wg *sync.WaitGroup, candlech chan<- candle.Candle, ch <-chan struct{}) error {
	defer wg.Done()
	url := fmt.Sprintf("https://www.binance.com/api/v3/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d", ms.symb, ms.tf, ms.opentime, ms.opentime+(ms.tfint*60000-1))
	<-ch
	t := time.Now()
	resp, err := http.Get(url)
	tt := time.Since(t)

	log.Println(ms)

	if err != nil && resp.StatusCode != 429 {
		return err
	} else if resp.StatusCode == 429 {
		fmt.Println("response code 429, sending pause signal to channel, sleeping 70seconds. Goroutine nums:", runtime.NumGoroutine())
		resp, err = http.Get(url)
		if err != nil {
			return err
		}
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var test candle.TestAPIStruct
	err = json.Unmarshal(body, &test)

	if len(test) != 0 {
		c := candle.ConvertAPItoCandleStruct(ms.symb, ms.tf, test[0])

		candlech <- c

		log.Println("BinanceOneAPI ", ms.symb, ms.tf, time.UnixMilli(ms.opentime).UTC(), "time for GET", tt)
	}
	return nil
}

func createCandleCheckForExistsInternal(db *pg.DB, candlech chan candle.Candle) {
	for req := range candlech {
		//t := time.Now()
		//time.Sleep(30 * time.Millisecond)
		_, err := db.Model(&req).
			Where("candle.my_id = ?", req.MyID).
			SelectOrInsert()
		if err != nil {
			log.Println(err)
			return
		}
		/*if inserted {
			log.Println("Time spent on adding candle to db", req.CoinTF, req.UTCOpenTime, time.Since(t))
		} else {
			log.Println("Candle exists in DB:", req.CoinTF, req.UTCOpenTime, time.Since(t))
		}*/

	}
}
