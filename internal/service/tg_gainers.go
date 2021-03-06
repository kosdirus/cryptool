package service

import (
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/storage/psql"
	"github.com/kosdirus/cryptool/internal/storage/psql/initdata"
	"log"
	"time"
	"unicode"
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

// Func isInt checks given string if it consists only of int numbers.
func isInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// Func checkTimeFormat checks string if it has proper time format for request.
// Proper string may be "20220531" (8 bytes) or "220531" (6 bytes) and can contain only numbers.
func checkTimeFormat(receivedTime string) (time.Time, bool) {
	var t time.Time
	var err error

	if len(receivedTime) == 6 && isInt(receivedTime) {
		t, err = time.Parse("20060102", "20"+receivedTime)
		if err != nil {
			log.Println("Error during StrongDuringDowntrend func (time Parse):", err)
			return t, false
		}
	} else if len(receivedTime) == 8 && isInt(receivedTime) {
		t, err = time.Parse("20060102", receivedTime)
		if err != nil {
			log.Println("Error during StrongDuringDowntrend func (time Parse):", err)
			return t, false
		}
	} else {
		log.Println("date is not valid, try use one of next formats: ")
		return t, false
	}

	return t, true
}

// StrongDuringDowntrend receives string which is highTime, then compare "high" value for
// 1d timeframe candle (for given date in highTime argument) and "close" value for last 30m closed candle.
//
// It's advised to use this func on downtrend (if now price for trade pairs are lower comparing to time provided
// in argument) - to see change of price in percentage and find strong and most gained coins during downtrend.
func StrongDuringDowntrend(pgdb *pg.DB, highTime string) map[string]float64 {
	t, ok := checkTimeFormat(highTime)
	if !ok {
		return nil
	}

	tint := t.UnixMilli()
	candleMap1 := candleHandlerSDD(pgdb, "1d", tint)

	tint1 := time.Now().UnixMilli()
	candleMap2 := candleHandlerSDD(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	var i int
	for i = 0; len(candleMap2) != len(initdata.TradePairList) && i < 10; i++ {
		time.Sleep(20 * time.Second)
		candleMap2 = candleHandlerSDD(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	}
	if i == 10 {
		return nil
	}

	finalCandleMap := make(map[string]float64)
	for s := range candleMap1 {
		finalCandleMap[s] = (candleMap2[s] - candleMap1[s]) / candleMap1[s] * 100
	}
	log.Println(len(finalCandleMap), len(candleMap1), len(candleMap2))

	return finalCandleMap
}

// Func candleHandlerSDD used to send requests to database and return map required for comparison in next steps.
func candleHandlerSDD(pgdb *pg.DB, timeframe string, tint int64) map[string]float64 {
	candleList, err := psql.GetCandleByTimeframeDate(pgdb, timeframe, tint)
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

// StrongDuringUptrend receives string which is lowTime, then compare "low" value for
// 1d timeframe candle (for given date in lowTime argument) and "close" value for last 30m closed candle.
//
// It's advised to use this func on uptrend (if now price for trade pairs are higher comparing to time provided
// in argument) - to see change of price in percentage and find strong and most gained coins during uptrend.
func StrongDuringUptrend(pgdb *pg.DB, lowTime string) map[string]float64 {
	t, ok := checkTimeFormat(lowTime)
	if !ok {
		return nil
	}

	tint := t.UnixMilli()
	candleMap1 := candleHandlerSDU(pgdb, "1d", tint)

	tint1 := time.Now().UnixMilli()
	candleMap2 := candleHandlerSDU(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	var i int
	for ; len(candleMap2) != len(initdata.TradePairList) && i < 4; i++ {
		time.Sleep(20 * time.Second)
		candleMap2 = candleHandlerSDU(pgdb, "30m", tint1-tint1%(30*60000)-30*60000)
	}
	if i == 4 {
		return nil
	}

	finalCandleMap := make(map[string]float64)
	for s := range candleMap1 {
		finalCandleMap[s] = (candleMap2[s] - candleMap1[s]) / candleMap1[s] * 100
	}
	log.Println(len(finalCandleMap), len(candleMap1), len(candleMap2))

	return finalCandleMap
}

// Func candleHandlerSDU used to send requests to database and return map required for comparison in next steps.
func candleHandlerSDU(pgdb *pg.DB, timeframe string, tint int64) map[string]float64 {
	candleList, err := psql.GetCandleByTimeframeDate(pgdb, timeframe, tint)
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
