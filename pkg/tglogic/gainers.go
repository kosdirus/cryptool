package tglogic

import (
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/cmd/interal/candle"
	"github.com/kosdirus/cryptool/cmd/interal/symbol"
	"log"
	"time"
)

/*
	1. send request through telegram like "sdd 20220403" or "sdu 20220430" for StrongDuringDowntrend and StrongDuringUptrend respectively
		1.1. receive high(sdd) or low(sdu) for 12h candle by date(!) and close value of last 30m closed candle for all candles
		1.2. calculate gain/loss in percentage for each symbol(coin). Formula = (priceNow - priceHistory)/priceHistory*100
		1.3. store values in ?map? to sort them later
		1.4. send values to kynselbot service (maybe it would be good for beginning to limit 50 symbols in output)
	2. receive list of symbols sorted by high gain (for StrongDuringUptrend) or less loss (for StrongDuringDowntrend)
		2.1. paint over with green positive values and with red negative ones
		2.2. send message to user

*/

func StrongDuringDowntrend(pgdb *pg.DB, HighTime string) map[string]float64 {
	t, err := time.Parse("20060102", HighTime)
	if err != nil {
		log.Println("Error during StrongDuringDowntrend func (time Parse):", err)
		return nil
	}
	tint := t.UnixMilli()
	candleMap1 := candleHandlerSDD(pgdb, "1d", tint)

	tint1 := time.Now().UnixMilli()
	candleMap2 := candleHandlerSDD(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	for len(candleMap2) != len(symbol.SymbolList) {
		time.Sleep(20 * time.Second)
		candleMap2 = candleHandlerSDD(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	}
	finalCandleMap := make(map[string]float64)
	for s := range candleMap1 {
		finalCandleMap[s] = (candleMap2[s] - candleMap1[s]) / candleMap1[s] * 100
	}
	log.Println(len(finalCandleMap), len(candleMap1), len(candleMap2))

	return finalCandleMap
}

func StrongDuringUptrend(pgdb *pg.DB, LowTime string) map[string]float64 {
	t, err := time.Parse("20060102", LowTime)
	if err != nil {
		log.Println("Error during StrongDuringDowntrend func (time Parse):", err)
		return nil
	}
	tint := t.UnixMilli()
	candleMap1 := candleHandlerSDU(pgdb, "1d", tint)

	tint1 := time.Now().UnixMilli()
	candleMap2 := candleHandlerSDU(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	for len(candleMap2) != len(symbol.SymbolList) {
		time.Sleep(20 * time.Second)
		candleMap2 = candleHandlerSDU(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	}
	finalCandleMap := make(map[string]float64)
	for s := range candleMap1 {
		finalCandleMap[s] = (candleMap2[s] - candleMap1[s]) / candleMap1[s] * 100
	}
	log.Println(len(finalCandleMap), len(candleMap1), len(candleMap2))

	return finalCandleMap
}

func candleHandlerSDD(pgdb *pg.DB, timeframe string, tint int64) map[string]float64 {
	candleList, err := candle.GetCandleByTimeframeDate(pgdb, timeframe, tint)
	if err != nil {
		log.Println("Error during StrongDuringDowntrend func (DB request):", err)
		return nil
	}
	candleMap := make(map[string]float64)
	for _, v := range candleList {
		candleMap[v.Coin] = v.High
	}

	return candleMap
}

func candleHandlerSDU(pgdb *pg.DB, timeframe string, tint int64) map[string]float64 {
	candleList, err := candle.GetCandleByTimeframeDate(pgdb, timeframe, tint)
	if err != nil {
		log.Println("Error during StrongDuringUptrend func (DB request):", err)
		return nil
	}
	candleMap := make(map[string]float64)
	for _, v := range candleList {
		candleMap[v.Coin] = v.Low
	}

	return candleMap
}
