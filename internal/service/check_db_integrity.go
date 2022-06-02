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

// CheckDBIntegrity calls check in goroutines for all trade pairs and timeframes and waiting for all to be done.
// If during function run was at least 1 added candle to database - function will call itself in recursive call.
// This behaviour is caused by the specifics of checkByCoinAndTimeframe func.
func CheckDBIntegrity(pgdb *pg.DB) {
	t := time.Now()
	log.Println("DB integrity check started: ", t.UTC())
	var total int64

	var wg sync.WaitGroup

	ch := make(chan struct{}, 7)
	for _, s := range initdata.TradePairList {
		for tfi, tf := range initdata.TimeframeMapReverse {
			wg.Add(1)
			go func(s, tf string, tfi int) {
				ch <- struct{}{}
				checkByCoinAndTimeframe(pgdb, s, tf, tfi, &total)
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
	} else {
		log.Println("ChechDBIntegrity finished.")
	}

}

// Type openTimeFromDB used for database requests to use less resources comparing to situation when
// whole Candle struct would be received
type openTimeFromDB struct {
	OpenTime int64 `bson:"open_time" json:"open_time" pg:"open_time,use_zero"`
}

// checkByCoinAndTimeframe sends request to database and receives slice of open_time int64 for given
// coin and timeframe. After receiving slice - func iterates through it and checks if difference between
// candle's open_time is equal to timeframe.
func checkByCoinAndTimeframe(pgdb *pg.DB, symbol, timeframe string, timeframeint int, total *int64) {
	c := &[]openTimeFromDB{}
	pgdb.Model(&core.Candle{}).
		Column("open_time").
		Where("coin = ? AND timeframe = ?", symbol, timeframe).
		Order("open_time DESC").
		Select(c)

	var n, n1 int64

	for i, v := range *c {
		if i == 0 {
			n = v.OpenTime
		} else if i == 1 {
			n1 = v.OpenTime

			if (n-n1)/60000 != int64(timeframeint) {
				log.Println("Data is NOT ok for:", symbol, timeframe, ". Open_time (prev, next):", n, n1)
				atomic.AddInt64(total, 1)

				err := internalBinanceOneAPI(pgdb, symbol, timeframe, int64(timeframeint), n1+int64(timeframeint*60000))
				if err != nil {
					log.Println("error while check db integrity and adding one api (check_db_integrity.go line:83)", symbol, timeframe, "Err:", err)
					return
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
					log.Println("error while check db integrity and adding one api (check_db_integrity.go line:99)", symbol, timeframe, "Err:", err)
				} else {
					log.Println("Added candle to db:", symbol, timeframe, n1+int64(timeframeint*60000))
				}
			}
		}
	}
}

// Func internalBinanceOneAPI is similar to previous versions of BinanceOneAPI - it doesn't have
// syncing through channels, just immediately request data from Binance.
// TODO same coordination logic as BinanceAPI has. Perfectly it should share the same syncing channel.
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
	var value [][]interface{}
	err = json.Unmarshal(body, &value)

	if len(value) != 0 {
		c := ConvertAPItoCandleStruct(symbol, timeframe, value[0])

		_, err = psql.CreateCandleCheckForExists(pgdb, &c)
		if err != nil {
			return err
		}

		log.Println("BinanceOneAPI ", symbol, timeframe, time.UnixMilli(openTime).UTC(), "time for GET", tt)
	}

	return nil
}
