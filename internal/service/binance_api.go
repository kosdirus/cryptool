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

func BinanceAPISchedule(pgdb *pg.DB, BinanceAPIrun *uint32) {
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
	go BinanceAPI(pgdb, BinanceAPIrun)
	ticker := time.NewTicker(30 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				BinanceAPI(pgdb, BinanceAPIrun)
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

func getMergedSliceOnce(pgdb *pg.DB, mergedSlice *[]mergedST) {
	allLastCandles, err := psql.GetAllLastCandles(pgdb)
	if err != nil {
		log.Println("getMergedSliceOnce error:  ", err)
		return
	}
	mx := sync.Mutex{}
	for _, a := range allLastCandles {
		for openTime := a.OpenTime; time.Now().UnixMilli()-openTime > a.TimeframeInt*60000*2; openTime = openTime + a.TimeframeInt*60000 {
			if err != nil {
				log.Println("binance_api.go line 41:", err, a.Coin, a.Timeframe)
				continue
			}
			mx.Lock()
			*mergedSlice = append(*mergedSlice, mergedST{
				symb:     a.Coin,
				tf:       a.Timeframe,
				tfint:    a.TimeframeInt,
				opentime: openTime + a.TimeframeInt*60000,
			})
			mx.Unlock()
		}
	}
}

func coordinateBinanceOneAPI(length int, ch chan<- struct{}, donech <-chan struct{}, resch <-chan struct{}) {
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

	var mergedSlice []mergedST
	tmerged := time.Now()
	getMergedSliceOnce(pgdb, &mergedSlice)

	log.Println("Time spent on merge slice:", time.Since(tmerged))
	tbin := time.Now()
	log.Println("START OF CALLING GOROUTINES", tbin.Format(time.StampMicro))
	wg.Add(len(mergedSlice))

	//log.Println(mergedSlice)
	tformerger := time.Now()
	for _, ms := range mergedSlice {
		go func(ms mergedST) {
			BinanceOneAPI(ms, &wg, candlech, ch, resch)
		}(ms)
		atomic.AddUint64(&nCandles, 1)

	}
	log.Println("Time for calling all goroutines: ", time.Since(tformerger), len(mergedSlice))
	go coordinateBinanceOneAPI(len(mergedSlice), ch, donech, resch)

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

func BinanceOneAPI(ms mergedST, wg *sync.WaitGroup, candlech chan<- core.Candle, ch <-chan struct{}, resch chan<- struct{}) error {
	defer wg.Done()
	url := fmt.Sprintf("https://www.binance.com/api/v3/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d", ms.symb, ms.tf, ms.opentime, ms.opentime+(ms.tfint*60000-1))
	<-ch
	t := time.Now()
	resp, err := http.Get(url)
	tt := time.Since(t)
	resch <- struct{}{}

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
	var value [][]string
	err = json.Unmarshal(body, &value)

	if len(value) != 0 {
		c := ConvertBCtoCandleStruct(ms.symb, ms.tf, ConvertRawToStruct(value[0]))

		candlech <- c

		//log.Println("BinanceOneAPI ", ms.symb, ms.tf, time.UnixMilli(ms.opentime).UTC().Format("2006.01.02 15:04"), "time for GET",
		//	tt, "Time now:", time.Now().Format(time.StampMicro))
	}
	return nil
}

func createCandleCheckForExistsInternal(db *pg.DB, candlech chan core.Candle) {
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
