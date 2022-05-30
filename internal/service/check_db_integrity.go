package service

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/core"
	"github.com/kosdirus/cryptool/internal/storage/psql"
	"github.com/kosdirus/cryptool/internal/storage/psql/initdata"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func CheckDBIntegrity(pgdb *pg.DB) {
	t := time.Now()
	log.Println("DB integrity check started: ", t.UTC())
	var total int64

	var wg sync.WaitGroup

	ch := make(chan struct{}, 5)
	for _, s := range initdata.TradePairList {
		for tfi, tf := range initdata.TimeframeMap {
			wg.Add(1)
			go func(s, tf string, tfi int) {
				ch <- struct{}{}
				getInfoByCoinAndTimeframe(pgdb, s, tf, tfi, &total)
				<-ch
				wg.Done()
			}(s, tf, tfi)

		}
	}
	wg.Wait()
	close(ch)
	log.Println("DB integrity check finished.", time.Since(t), "Total number of added candles:", total)
	if total != 0 {
		CheckDBIntegrity(pgdb)
	}
	select {}
}

func getInfoByCoinAndTimeframe(pgdb *pg.DB, symbol, timeframe string, timeframeint int, total *int64) {

	c := &[]core.Candle{}
	pgdb.Model(c).
		Where("coin = ? AND timeframe = ?", symbol, timeframe).
		Order("open_time DESC").
		Select()

	var n, n1 int64

	for i, v := range *c {
		if i == 0 {
			n = v.OpenTime
		} else if i == 1 {
			n1 = v.OpenTime

			if (n-n1)/60000 != int64(timeframeint) {
				log.Println("Data is NOT ok for:", symbol, timeframe, ". List of open_time:", n, n1)
				atomic.AddInt64(total, 1)

				err := internalBinanceOneAPI(pgdb, symbol, timeframe, int64(timeframeint), n1+int64(timeframeint*60000))
				if err != nil {
					log.Println("error while check db integrity and adding one api (check_db_integrity.go line:41)", symbol, timeframe, "Err:", err)
				} else {
					log.Println("Added candle to db:", symbol, timeframe, n1+int64(timeframeint*60000))
				}
			}
		} else {
			n, n1 = n1, v.OpenTime

			n1 = v.OpenTime
			if (n-n1)/60000 != int64(timeframeint) {
				log.Println("Data is NOT ok for:", symbol, timeframe, ". List of open_time:", n, n1)
				atomic.AddInt64(total, 1)

				err := internalBinanceOneAPI(pgdb, symbol, timeframe, int64(timeframeint), n1+int64(timeframeint*60000))
				if err != nil {
					log.Println("error while check db integrity and adding one api (check_db_integrity.go line:58)", symbol, timeframe, "Err:", err)
				} else {
					log.Println("Added candle to db:", symbol, timeframe, n1+int64(timeframeint*60000))
				}
			}
		}
	}
}

func internalBinanceOneAPI(pgdb *pg.DB, symbol, timeframe string, k1, openTime int64) error {

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
	var value [][]string
	err = json.Unmarshal(body, &value)

	if len(value) != 0 {
		c := ConvertBCtoCandleStruct(symbol, timeframe, ConvertRawToStruct(value[0]))

		_, err = psql.CreateCandleCheckForExists(pgdb, &c)
		if err != nil {
			return err
		}

		log.Println("BinanceOneAPI ", symbol, timeframe, time.UnixMilli(openTime).UTC(), "time for GET", tt)
	}

	return nil
}
