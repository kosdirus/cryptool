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

type mergedST struct {
	symb     string
	tf       string
	tfint    int64
	opentime int64
}

func getMergedSlice(pgdb *pg.DB, mergedSlice *[]mergedST) {
	var goroutines = 6
	if os.Getenv("ENV") == "DIGITAL" {
		goroutines = 2
	}
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

func coordinateBinanceOneAPI(length int, ch chan<- struct{}, donech <-chan struct{}, resch <-chan struct{}) {
	time.Sleep(5 * time.Second)
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
			case <-resch:
			}
		}
		ticker.Reset(62 * time.Second)
	}
	proc()

	for {
		select {
		case <-ticker.C:
			proc()
		case <-donech:
			log.Println("Donech received, coordinateBinanceOneAPI finish!!")
			return
		}
	}
}

func BinanceAPI(pgdb *pg.DB) {
	fmt.Println("BinanceAPI start", time.Now().UTC())
	candlech := make(chan candle.Candle, 15)
	ch := make(chan struct{})
	resch := make(chan struct{}, 500)
	donech := make(chan struct{}, 1)
	var wg sync.WaitGroup
	go createCandleCheckForExistsInternal(pgdb, candlech)
	tt := time.Now()
	var nCoins, nCandles uint64

	var mergedSlice []mergedST
	tmerged := time.Now()
	getMergedSlice(pgdb, &mergedSlice)
	go coordinateBinanceOneAPI(len(mergedSlice), ch, donech, resch)

	//log.Println(mergedSlice)
	log.Println("Time spent on merge slice:", time.Since(tmerged))
	tbin := time.Now()
	log.Println("START OF CALLING GOROUTINES", tbin.Format(time.StampMicro))
	wg.Add(len(mergedSlice))

	for _, ms := range mergedSlice {
		go func(ms mergedST) {
			BinanceOneAPI(ms, &wg, candlech, ch, resch)
		}(ms)
		atomic.AddUint64(&nCandles, 1)

	}

	//atomic.AddUint64(&nCandles, 1)
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

func BinanceOneAPI(ms mergedST, wg *sync.WaitGroup, candlech chan<- candle.Candle, ch <-chan struct{}, resch chan<- struct{}) error {
	defer wg.Done()
	url := fmt.Sprintf("https://www.binance.com/api/v3/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d", ms.symb, ms.tf, ms.opentime, ms.opentime+(ms.tfint*60000-1))
	<-ch
	t := time.Now()
	resp, err := http.Get(url)
	tt := time.Since(t)
	resch <- struct{}{}
	//runtime.Gosched()

	log.Println(ms, t.Format(time.StampMicro), tt)

	if err != nil && resp.StatusCode != 429 {
		return err
	} else if resp.StatusCode == 429 {
		fmt.Println("response code 429, sending pause signal to channel, sleeping 70seconds. Goroutine nums:", runtime.NumGoroutine())
		/*resp, err = http.Get(url)
		if err != nil {
			return err
		}*/
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var test candle.TestAPIStruct
	err = json.Unmarshal(body, &test)

	if len(test) != 0 {
		c := candle.ConvertAPItoCandleStruct(ms.symb, ms.tf, test[0])

		candlech <- c

		//log.Println("BinanceOneAPI ", ms.symb, ms.tf, time.UnixMilli(ms.opentime).UTC().Format("2006.01.02 15:04"), "time for GET",
		//	tt, "Time now:", time.Now().Format(time.StampMicro))
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
