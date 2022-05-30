package psql

import (
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/core"
	"github.com/kosdirus/cryptool/internal/storage/psql/initdata"
	"log"
	"time"
)

func CreateCandle(db *pg.DB, req *core.Candle) (*core.Candle, error) {
	_, err := db.Model(req).Insert()
	if err != nil {
		return nil, err
	}

	candle := &core.Candle{}
	//log.Println("converting_candle.go line 71")
	err = db.Model(candle).
		Where("candle.my_id = ?", req.MyID).
		Select()
	//log.Println("converting_candle.go line 75")
	return candle, err
}

func CreateCandleCheckForExists(db *pg.DB, req *core.Candle) (bool, error) {
	inserted, err := db.Model(req).
		Where("candle.my_id = ?", req.MyID).
		SelectOrInsert()
	return inserted, err
}

func GetCandleByMyID(db *pg.DB, candleMyID string) (*core.Candle, error) {
	candle := &core.Candle{}
	log.Println("converting_candle.go line 81")
	err := db.Model(candle).
		Where("candle.my_id = ?", candleMyID).
		Select()
	log.Println("converting_candle.go line 85")
	return candle, err
}

func GetCandleByTimeframeDate(db *pg.DB, timeframe string, opentime int64) ([]core.Candle, error) {
	c := &[]core.Candle{}
	err := db.Model(c).
		Where("candle.timeframe = ? AND candle.open_time = ?", timeframe, opentime).
		Select()
	log.Println("converting_candle.go getCandleByTimeframeDate. Len of this timeframe and opentime is:", len(*c), "Timeframe", timeframe, "opentime", opentime)
	return *c, err
}

func GetCandleBySymbolAndTimeframe(db *pg.DB, symbol, timeframe string) (*core.Candle, error) {
	candle := &core.Candle{}
	//log.Println("converting_candle.go line 81")
	err := db.Model(candle).
		Where("candle.coin = ? AND candle.timeframe = ?", symbol, timeframe).
		Order("open_time DESC").
		Limit(1).
		Select()
	log.Println("converting_candle.go line 85")
	return candle, err
}

type AllLastOpenTime struct {
	Coin         string `bson:"coin" json:"coin" pg:"coin,use_zero"`
	Timeframe    string `bson:"timeframe" json:"timeframe" pg:"timeframe,use_zero"`
	TimeframeInt int64  `bson:"timeframeint" json:"timeframeint" pg:"timeframeint,use_zero"`
	OpenTime     int64  `bson:"open_time" json:"open_time" pg:"open_time,use_zero"`
}

func GetAllLastCandles(db *pg.DB) ([]AllLastOpenTime, error) {
	t := time.Now()

	var all []AllLastOpenTime
	err := db.Model(&core.Candle{}).
		Column("coin").
		Column("timeframe").
		ColumnExpr("MAX(open_time) as open_time").
		Group("coin").
		Group("timeframe").
		Select(&all)
	for i := range all {
		(&all[i]).TimeframeInt = int64(initdata.TimeframeMapReverse[all[i].Timeframe])
	}
	//log.Println(all)
	log.Println("GetAllLastCandles:  Time spent: ", time.Since(t))
	return all, err
}

func UpdateCandle(db *pg.DB, req *core.Candle) (*core.Candle, error) {
	_, err := db.Model(req).
		Where("candle.my_id = ?", req.MyID).
		Update()
	if err != nil {
		return nil, err
	}

	candle := &core.Candle{}
	log.Println("converting_candle.go line 119")
	err = db.Model(candle).
		Where("candle.my_id = ?", req.MyID).
		Select()
	log.Println("converting_candle.go line 123")
	return candle, err
}

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
