package service

import (
	"errors"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/storage/psql/initdata"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// BinanceZipToPostgres used to download CSV files from Binance for given coins and timeframes.
// It coordinates and calls all needed functions for downloading csv and checksum files and then
// convert data and store it in database.
func BinanceZipToPostgres(pgdb *pg.DB) {
	tn := time.Now()
	var nFilesDownload, nFilesChecksum, nFilesParse, nRawsParse, nDuplicateParse uint64
	var goroutines int
	if os.Getenv("ENV") == "DIGITAL" {
		goroutines = 15
	} else {
		goroutines = runtime.NumCPU() * 5
	}

	totalCoins := len(initdata.TradePairList)
	lastGoroutine := goroutines - 1
	stride := totalCoins / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			start := g * stride
			end := start + stride
			if g == lastGoroutine {
				end = totalCoins
			}

			for _, symb := range initdata.TradePairList[start:end] {
				for timeframe, days := range initdata.TimeframeDays {
					t := time.Now().Add(-24 * time.Hour)
					for i := 0; i < days; i++ {
						path := FindPath(symb, timeframe, t)
						err := FileDownload(FindURL(symb, timeframe, t), path)
						if fmt.Sprint(err) == "404" {
							break
						} else if err != nil {
							log.Println("zip_binance.go line 43:", err)
							return
						}

						err = FileDownload(FindURLChecksum(symb, timeframe, t), FindPathChecksum(symb, timeframe, t))
						if err != nil {
							log.Println("zip_binance.go line 49:", err)
							return
						}

						if _, exists := os.Stat(path); exists == nil {
							if ok := checkChecksum(path); ok == true {
								atomic.AddUint64(&nFilesChecksum, 1)
							} else {
								log.Println("Checksum error:", path)
								continue
							}
							err = ParseCSV(symb, timeframe, &nRawsParse, &nDuplicateParse, path, pgdb)
							if err != nil {
								log.Println("zip_binance.go line 64:", err, path)
								return
							}
							//w.Write([]byte(fmt.Sprintf("cvs parsed: %s\n", path)))
							atomic.AddUint64(&nFilesParse, 1)
						}
						t = t.Add(-24 * time.Hour)
						atomic.AddUint64(&nFilesDownload, 2)
					}
				}
			}
			wg.Done()
		}(g)
	}
	wg.Wait()
	log.Println("Downloaded", nFilesDownload, "files. Time spent on AllDownloadBC:", time.Since(tn))
	log.Println("Checked", nFilesChecksum, "file pairs. Time spent on AllChecksumBC:", time.Since(tn))
	log.Println("Parsed", nFilesParse, "files. Parsed", nRawsParse, "unique rows, were omitted", nDuplicateParse, "duplicate raws. "+
		"Time spent on ParseAllCSVToMongo:", time.Since(tn))
}

// FileDownload downloads file from given URL and saves it with fileName given in second argument.
func FileDownload(URL, fileName string) error {
	//check if file exists
	if _, err := os.Stat(fileName); err == nil {
		//fmt.Println("file", fileName, "already exists")
		return nil
	}

	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 && response.StatusCode != 404 {
		log.Println(response.StatusCode)
		return errors.New("received non 200 response code")
	} else if response.StatusCode == 404 {
		return errors.New("404")
	}

	//Create an empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
