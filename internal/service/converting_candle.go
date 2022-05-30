package service

import (
	"github.com/kosdirus/cryptool/internal/core"
	"strconv"
	"time"
)

func ConvertRawToStruct(c []string) (bc core.BinanceCandle) {
	bc.OpenTime, _ = strconv.ParseInt(c[0], 10, 64)
	bc.Open, _ = strconv.ParseFloat(c[1], 64)
	bc.High, _ = strconv.ParseFloat(c[2], 64)
	bc.Low, _ = strconv.ParseFloat(c[3], 64)
	bc.Close, _ = strconv.ParseFloat(c[4], 64)
	bc.Volume, _ = strconv.ParseFloat(c[5], 64)
	bc.CloseTime, _ = strconv.ParseInt(c[6], 10, 64)
	bc.QuoteAssetVolume, _ = strconv.ParseFloat(c[7], 64)
	bc.NumberOfTrades, _ = strconv.ParseInt(c[8], 10, 64)
	bc.TakerBuyBaseAssetVolume, _ = strconv.ParseFloat(c[9], 64)
	bc.TakerBuyQuoteAssetVolume, _ = strconv.ParseFloat(c[10], 64)
	bc.Ignore, _ = strconv.ParseFloat(c[11], 64)
	return bc
}

func ConvertBCtoCandleStruct(symbol, timeframe string, bc core.BinanceCandle) (c core.Candle) {
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
