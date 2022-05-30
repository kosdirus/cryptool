package service

import (
	"fmt"
	"log"
	"os"
	"time"
)

func FindPath(symbol, timeframe string, t time.Time) string {
	y, m, d := t.Date()
	res := fmt.Sprintf("./tooldata/%s/%s/%s-%s-%d-%.2d-%.2d.zip", symbol,
		timeframe, symbol, timeframe, y, int(m), d)

	// If there is no such directory - it's created
	s := fmt.Sprintf("./tooldata/%s/%s/", symbol, timeframe)
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

func FindURL(symbol, timeframe string, t time.Time) string {
	//"https://data.binance.vision/data/spot/daily/klines/TWTUSDT/15m/TWTUSDT-15m-2022-03-04.zip"
	y, m, d := t.Date()
	return fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/%s/%s-%s-%d-%.2d-%.2d.zip", symbol,
		timeframe, symbol, timeframe, y, int(m), d)
}

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

func FindURLChecksum(symbol, timeframe string, t time.Time) string {
	//"https://data.binance.vision/data/spot/daily/klines/TWTUSDT/15m/TWTUSDT-15m-2022-03-04.zip"
	y, m, d := t.Date()
	return fmt.Sprintf("https://data.binance.vision/data/spot/daily/klines/%s/%s/%s-%s-%d-%.2d-%.2d.zip.CHECKSUM", symbol,
		timeframe, symbol, timeframe, y, int(m), d)
}
