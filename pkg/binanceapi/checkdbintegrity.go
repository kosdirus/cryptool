package binanceapi

import (
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/cmd/interal/candle"
	"github.com/kosdirus/cryptool/cmd/interal/symbol"
	"log"
	"time"
)

func CheckDBIntegrity(pgdb *pg.DB) {
	t := time.Now()
	for _, s := range symbol.SymbolList {
		for tfi, tf := range symbol.TimeframeMap {
			getInfoByCoinAndTimeframe(pgdb, s, tf, tfi)
		}
	}
	log.Println("DB integrity check finished.", time.Since(t))
}

func getInfoByCoinAndTimeframe(pgdb *pg.DB, symbol, timeframe string, timeframeint int) {

	c := &[]candle.Candle{}
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

				err := BinanceOneAPI(pgdb, symbol, timeframe, int64(timeframeint), n1+int64(timeframeint*60000))
				if err != nil {
					log.Println("error while check db integrity and adding one api (checkdbintegrity.go line:41)", symbol, timeframe, "Err:", err)
				} else {
					log.Println("Added candle to db:", symbol, timeframe, n1+int64(timeframeint*60000))
				}
			}
		} else {
			n, n1 = n1, v.OpenTime

			n1 = v.OpenTime
			if (n-n1)/60000 != int64(timeframeint) {
				log.Println("Data is NOT ok for:", symbol, timeframe, ". List of open_time:", n, n1)

				err := BinanceOneAPI(pgdb, symbol, timeframe, int64(timeframeint), n1+int64(timeframeint*60000))
				if err != nil {
					log.Println("error while check db integrity and adding one api (checkdbintegrity.go line:58)", symbol, timeframe, "Err:", err)
				} else {
					log.Println("Added candle to db:", symbol, timeframe, n1+int64(timeframeint*60000))
				}
			}
		}
	}
}
