package service

import (
	"archive/zip"
	"encoding/csv"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/storage/psql"
	"log"
	"sync/atomic"
)

// ParseCSV reads .zip file with Binance candles data, read CSV files, convert raws to Candle struct and
// add this candle to database.
func ParseCSV(symbol1, timeframe1 string, nRaws, nDupRaws *uint64, path string, pgdb *pg.DB) error {
	var n, ndup uint64
	z, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer func(z *zip.ReadCloser) {
		err = z.Close()
		if err != nil {
			log.Println("zip_parsecsv.go line 67:", err)
		}
	}(z)
	for _, f := range z.File {
		r, _ := f.Open()
		c, _ := csv.NewReader(r).ReadAll()
		for i := range c {
			candle := ConvertBCtoCandleStruct(symbol1, timeframe1, ConvertRawToStruct(c[i]))
			inserted, err := psql.CreateCandleCheckForExists(pgdb, &candle)
			if !inserted && err == nil {
				ndup++
				continue
			} else if err != nil {
				return err
			}
			n++
		}
	}
	atomic.AddUint64(nRaws, n)
	atomic.AddUint64(nDupRaws, ndup)
	return nil
}
