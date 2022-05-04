package candle

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"log"
	"strconv"
	"time"
)

func ConvertRawToStruct(c []string) (bc BinanceCandle) {
	for i, v := range c {
		switch i {
		case 0:
			bc.OpenTime, _ = strconv.ParseInt(v, 10, 64)
		case 1:
			bc.Open, _ = strconv.ParseFloat(v, 64)
		case 2:
			bc.High, _ = strconv.ParseFloat(v, 64)
		case 3:
			bc.Low, _ = strconv.ParseFloat(v, 64)
		case 4:
			bc.Close, _ = strconv.ParseFloat(v, 64)
		case 5:
			bc.Volume, _ = strconv.ParseFloat(v, 64)
		case 6:
			bc.CloseTime, _ = strconv.ParseInt(v, 10, 64)
		case 7:
			bc.QuoteAssetVolume, _ = strconv.ParseFloat(v, 64)
		case 8:
			bc.NumberOfTrades, _ = strconv.ParseInt(v, 10, 64)
		case 9:
			bc.TakerBuyBaseAssetVolume, _ = strconv.ParseFloat(v, 64)
		case 10:
			bc.TakerBuyQuoteAssetVolume, _ = strconv.ParseFloat(v, 64)
		case 11:
			bc.Ignore, _ = strconv.ParseFloat(v, 64)
		}
	}
	return bc
}

func ConvertBCtoCandleStruct(symbol, timeframe string, bc BinanceCandle) (c Candle) {
	c.CoinTF = symbol + timeframe
	c.MyID = c.CoinTF + strconv.FormatInt(bc.OpenTime, 10)
	c.Coin = symbol
	c.Timeframe = timeframe
	c.OpenTime = bc.OpenTime
	c.UTCOpenTime = time.UnixMilli(c.OpenTime).UTC()
	c.Open = bc.Open
	c.High = bc.High
	c.Low = bc.Low
	c.Close = bc.Close
	c.Volume = bc.Volume
	c.CloseTime = bc.CloseTime
	c.UTCCloseTime = time.UnixMilli(c.CloseTime).UTC()
	c.QuoteAssetVolume = bc.QuoteAssetVolume
	c.NumberOfTrades = bc.NumberOfTrades
	c.TakerBuyBaseAssetVolume = bc.TakerBuyBaseAssetVolume
	c.TakerBuyQuoteAssetVolume = bc.TakerBuyQuoteAssetVolume
	return c
}

func CreateCandle(db *pg.DB, req *Candle) (*Candle, error) {
	_, err := db.Model(req).Insert()
	if err != nil {
		return nil, err
	}

	candle := &Candle{}
	//log.Println("converting.go line 71")
	err = db.Model(candle).
		Where("candle.my_id = ?", req.MyID).
		Select()
	//log.Println("converting.go line 75")
	return candle, err
}

func CreateCandleCheckForExists(db *pg.DB, req *Candle) (bool, error) {
	inserted, err := db.Model(req).
		Where("candle.my_id = ?", req.MyID).
		SelectOrInsert()
	return inserted, err
}

func GetCandleByMyID(db *pg.DB, candleMyID string) (*Candle, error) {
	candle := &Candle{}
	log.Println("converting.go line 81")
	err := db.Model(candle).
		Where("candle.my_id = ?", candleMyID).
		Select()
	log.Println("converting.go line 85")
	return candle, err
}

func GetCandleByTimeframeDate(db *pg.DB, timeframe string, opentime int64) ([]Candle, error) {
	c := &[]Candle{}
	err := db.Model(c).
		Where("candle.timeframe = ? AND candle.open_time = ?", timeframe, opentime).
		Select()
	log.Println("converting.go getCandleByTimeframeDate. Len of this timeframe and opentime is:", len(*c), "Timeframe", timeframe, "opentime", opentime)
	return *c, err
}

func GetCandleBySymbolAndTimeframe(db *pg.DB, symbol, timeframe string) (*Candle, error) {
	candle := &Candle{}
	log.Println("converting.go line 81")
	err := db.Model(candle).
		Where("candle.coin = ? AND candle.timeframe = ?", symbol, timeframe).
		Order("open_time DESC").
		Limit(1).
		Select()
	log.Println("converting.go line 85")
	return candle, err
}

func GetLastCandle(db *pg.DB, coin, timeframe string) (int64, error) {
	candle := &Candle{}
	//log.Println("converting.go line 91")
	err := db.Model(candle).
		Where("candle.coin = ?", coin).
		Where("candle.timeframe = ?", timeframe).
		Order("open_time DESC").
		Limit(1).
		Select()
	//log.Println("converting.go line 97")
	return candle.OpenTime, err
}

func GetCandles(db *pg.DB) ([]*Candle, error) {
	candles := make([]*Candle, 0)
	log.Println("converting.go line 103")
	err := db.Model(&candles).
		Select()
	log.Println("converting.go line 106")
	return candles, err
}

func UpdateCandle(db *pg.DB, req *Candle) (*Candle, error) {
	_, err := db.Model(req).
		Where("candle.my_id = ?", req.MyID).
		Update()
	if err != nil {
		return nil, err
	}

	candle := &Candle{}
	log.Println("converting.go line 119")
	err = db.Model(candle).
		Where("candle.my_id = ?", req.MyID).
		Select()
	log.Println("converting.go line 123")
	return candle, err
}

func DeleteCandle(db *pg.DB, candleMyID string) error {
	candle := &Candle{
		MyID: candleMyID,
	}

	err := db.Model(candle).
		Where("candle.my_id = ?", candle.MyID).
		Select()
	if err != nil {
		return err
	}

	_, err = db.Model(candle).
		Where("candle.my_id = ?", candleMyID).
		Delete()

	return err
}

func ConvertAPItoCandleStruct(symbol, timeframe string, bc TestAPIexample) (c Candle) {
	c.CoinTF = symbol + timeframe
	c.MyID = c.CoinTF + fmt.Sprintf("%.f", bc[0])
	c.Coin = symbol
	c.Timeframe = timeframe
	c.OpenTime, _ = strconv.ParseInt(fmt.Sprintf("%.f", bc[0]), 10, 64)
	c.UTCOpenTime = time.UnixMilli(c.OpenTime).UTC()
	c.Open, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[1]), 64)
	c.High, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[2]), 64)
	c.Low, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[3]), 64)
	c.Close, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[4]), 64)
	c.Volume, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[5]), 64)
	c.CloseTime, _ = strconv.ParseInt(fmt.Sprintf("%.f", bc[6]), 10, 64)
	c.UTCCloseTime = time.UnixMilli(c.CloseTime).UTC()
	c.QuoteAssetVolume, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[7]), 64)
	c.NumberOfTrades, _ = strconv.ParseInt(fmt.Sprintf("%.f", bc[8]), 10, 64)
	c.TakerBuyBaseAssetVolume, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[9]), 64)
	c.TakerBuyQuoteAssetVolume, _ = strconv.ParseFloat(fmt.Sprintf("%s", bc[10]), 64)
	return c
}

type TestAPIStruct [][]interface{}

type TestAPIexample []interface{}
