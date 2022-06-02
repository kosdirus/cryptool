package psql

import (
	"errors"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/core"
	"github.com/kosdirus/cryptool/internal/storage/psql/initdata"
	"log"
	"time"
)

// CreateCandleCheckForExists inserts core.Candle to database if it's not already in database.
func CreateCandleCheckForExists(db *pg.DB, req *core.Candle) (bool, error) {
	inserted, err := db.Model(req).
		Where("candle.my_id = ?", req.MyID).
		SelectOrInsert()
	return inserted, err
}

// GetCandleByMyID returns core.Candle by given MyID (my_id).
func GetCandleByMyID(db *pg.DB, candleMyID string) (*core.Candle, error) {
	candle := &core.Candle{}
	log.Println("converting_candle.go line 81")
	err := db.Model(candle).
		Where("candle.my_id = ?", candleMyID).
		Select()
	log.Println("converting_candle.go line 85")
	return candle, err
}

// GetCandleByTimeframeDate returns core.Candle by given timeframe and open time.
func GetCandleByTimeframeDate(db *pg.DB, timeframe string, opentime int64) ([]core.Candle, error) {
	c := &[]core.Candle{}
	err := db.Model(c).
		Where("candle.timeframe = ? AND candle.open_time = ?", timeframe, opentime).
		Select()
	log.Println("converting_candle.go getCandleByTimeframeDate. Len of this timeframe and opentime is:", len(*c), "Timeframe", timeframe, "opentime", opentime)
	return *c, err
}

// GetCandleBySymbolAndTimeframe returns ONLY 1 last (by open time) core.Candle by given symbol and timeframe.
func GetCandleBySymbolAndTimeframe(db *pg.DB, symbol, timeframe string) (*core.Candle, error) {
	candle := &core.Candle{}
	err := db.Model(candle).
		Where("candle.coin = ? AND candle.timeframe = ?", symbol, timeframe).
		Order("open_time DESC").
		Limit(1).
		Select()
	return candle, err
}

// LastOpenTime struct is used for storing data about open time for certain coin and timeframe.
type LastOpenTime struct {
	Coin         string `bson:"coin" json:"coin" pg:"coin,use_zero"`
	Timeframe    string `bson:"timeframe" json:"timeframe" pg:"timeframe,use_zero"`
	TimeframeInt int64  `bson:"timeframeint" json:"timeframeint" pg:"timeframeint,use_zero"`
	OpenTime     int64  `bson:"open_time" json:"open_time" pg:"open_time,use_zero"`
}

// GetAllLastCandles returns slice of LastOpenTime - all last candles for each coin + timeframe.
func GetAllLastCandles(db *pg.DB) ([]LastOpenTime, error) {
	t := time.Now()

	var all []LastOpenTime
	err := db.Model(&core.Candle{}).
		Column("coin").
		Column("timeframe").
		ColumnExpr("MAX(open_time) as open_time").
		Group("coin").
		Group("timeframe").
		Select(&all)
	for i := range all {
		(&all[i]).TimeframeInt = int64(initdata.TimeframeMap[all[i].Timeframe])
	}
	log.Println("GetAllLastCandles:  Time spent: ", time.Since(t))
	return all, err
}

// UpdateCandle updates existing in database candle with given. Primary key is MyID.
func UpdateCandle(db *pg.DB, req *core.Candle) (bool, error) {
	result, err := db.Model(req).
		Where("candle.my_id = ?", req.MyID).
		Update()
	if err != nil {
		return false, err
	} else if result.RowsAffected() != 1 {
		return false, errors.New("some problem with UpdateCandle(), RowsAffected != 1")
	}

	return true, err
}

// DeleteCandle deletes candle from database based on given MyID field of candle.
func DeleteCandle(db *pg.DB, candleMyID string) error {
	candle := &core.Candle{
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
