package candle

import (
	"time"
)

type BinanceCandle struct {
	OpenTime                 int64   `json:"open_time"`
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	Volume                   float64 `json:"volume"`
	CloseTime                int64   `json:"close_time"`
	QuoteAssetVolume         float64 `json:"quote_asset_volume"`
	NumberOfTrades           int64   `json:"number_of_trades"`
	TakerBuyBaseAssetVolume  float64 `json:"taker_buy_base_asset_volume"`
	TakerBuyQuoteAssetVolume float64 `json:"taker_buy_quote_asset_volume"`
	Ignore                   float64 `json:"ignore"`
}

type Candle struct {
	//ID                       primitive.ObjectID `bson:"_id"`
	ID                       int64     `bson:"id" json:"id" pg:"id"`
	MyID                     string    `bson:"my_id" json:"my_id" pg:"my_id,use_zero"`
	CoinTF                   string    `bson:"coin_tf" json:"coin_tf" pg:"coin_tf,use_zero"`
	Coin                     string    `bson:"coin" json:"coin" pg:"coin,use_zero"`
	Timeframe                string    `bson:"timeframe" json:"timeframe" pg:"timeframe,use_zero"`
	UTCOpenTime              time.Time `bson:"utc_open_time" json:"utc_open_time" pg:"utc_open_time,use_zero"`
	OpenTime                 int64     `bson:"open_time" json:"open_time" pg:"open_time,use_zero"`
	Open                     float64   `bson:"open" json:"open" pg:"open,use_zero"`
	High                     float64   `bson:"high" json:"high" pg:"high,use_zero"`
	Low                      float64   `bson:"low" json:"low" pg:"low,use_zero"`
	Close                    float64   `bson:"close" json:"close" pg:"close,use_zero"`
	Volume                   float64   `bson:"volume" json:"volume" pg:"volume,use_zero"`
	UTCCloseTime             time.Time `bson:"utc_close_time" json:"utc_close_time" pg:"utc_close_time,use_zero"`
	CloseTime                int64     `bson:"close_time" json:"close_time" pg:"close_time,use_zero"`
	QuoteAssetVolume         float64   `bson:"quote_asset_volume" json:"quote_asset_volume" pg:"quote_asset_volume,use_zero"`
	NumberOfTrades           int64     `bson:"number_of_trades" json:"number_of_trades" pg:"number_of_trades,use_zero"`
	TakerBuyBaseAssetVolume  float64   `bson:"taker_buy_base_asset_volume" json:"taker_buy_base_asset_volume" pg:"taker_buy_base_asset_volume,use_zero"`
	TakerBuyQuoteAssetVolume float64   `bson:"taker_buy_quote_asset_volume" json:"taker_buy_quote_asset_volume" pg:"taker_buy_quote_asset_volume,use_zero"`
	MA50                     float64   `bson:"ma50" json:"ma50" pg:"ma50,use_zero"`
	MA50Trend                bool      `bson:"ma50trend" json:"ma50trend" pg:"ma50trend,use_zero"`
	MA100                    float64   `bson:"ma100" json:"ma100" pg:"ma100,use_zero"`
	MA100Trend               bool      `bson:"ma100trend" json:"ma100trend" pg:"ma100trend,use_zero"`
	MA200                    float64   `bson:"ma200" json:"ma200" pg:"ma200,use_zero"`
	MA200Trend               bool      `bson:"ma200trend" json:"ma200trend" pg:"ma200trend,use_zero"`
}
