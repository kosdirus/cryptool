package service

import (
	"fmt"
	"log"
	"os"
	"time"
)

// FindPath returns path for .zip file to be stored locally for given coin, timeframe and time.
func FindPath(coin, timeframe string, t time.Time) string {
	y, m, d := t.Date()
	res := fmt.Sprintf("./tooldata/%s/%s/%s-%s-%d-%.2d-%.2d.zip", coin,
		timeframe, coin, timeframe, y, int(m), d)

	// If there is no such directory - it's created
	s := fmt.Sprintf("./tooldata/%s/%s/", coin, timeframe)
	file, err := os.Open(s)
	if err != nil {
		err1 := os.MkdirAll(s, 0755)
		if err1 != nil {
			log.Println(err1)
		}
	}
	defer file.Close()
	return res
}

// FindURL returns .zip file URL for given coin, timeframe and time to download from Binance.
func FindURL(symbol, timeframe string, t time.Time) string {
	//"https://data.binance.vision/data/spot/daily/klines/TWTUSDT/15m/TWTUSDT-15m-2022-03-04.zip"
	y, m, d := t.Date()
	return fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/%s/%s-%s-%d-%.2d-%.2d.zip", symbol,
		timeframe, symbol, timeframe, y, int(m), d)
}

// FindPathChecksum returns path for .zip.CHECKSUM file to be stored locally for given coin, timeframe and time.
func FindPathChecksum(symbol, timeframe string, t time.Time) string {
	y, m, d := t.Date()
	res := fmt.Sprintf("./tooldata/%s/%s/%s-%s-%d-%.2d-%.2d.zip.CHECKSUM", symbol,
		timeframe, symbol, timeframe, y, int(m), d)

	// If there is no such directory - it's created
	s := fmt.Sprintf("./tooldata/%s/%s/", symbol, timeframe)
	file, err := os.Open(s)
	if err != nil {
		err1 := os.MkdirAll(s, 0755)
		if err1 != nil {
			fmt.Println(err1)
		}
	}
	defer file.Close()
	return res
}

// FindURLChecksum returns .zip.CHECKSUM file URL for given coin, timeframe and time to download from Binance.
func FindURLChecksum(symbol, timeframe string, t time.Time) string {
	//"https://data.binance.vision/data/spot/daily/klines/TWTUSDT/15m/TWTUSDT-15m-2022-03-04.zip.CHECKSUM"
	y, m, d := t.Date()
	return fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/%s/%s-%s-%d-%.2d-%.2d.zip.CHECKSUM", symbol,
		timeframe, symbol, timeframe, y, int(m), d)
}
