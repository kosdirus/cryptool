package binancezip

import (
	"crypto/sha256"
	"fmt"
	"github.com/kosdirus/cryptool/cmd/interal/symbol"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// example how to call AllChecksumBC func in main func
//internal.AllChecksumBC(1)

func AllChecksumBC(ndays int) {
	tn := time.Now()
	var nFiles uint64
	goroutines := runtime.NumCPU()/2 - 1
	totalCoins := len(symbol.SymbolList)
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

			for _, v := range symbol.SymbolList[start:end] {
				for _, v1 := range symbol.Timeframe {
					t := time.Now().Add(-24 * time.Hour)
					for i := 0; i < ndays; i++ {
						k := FindPath(v, v1, t)
						if _, exists := os.Stat(k); exists == nil {
							err := checkChecksum(k)
							if err == true {
								t = t.Add(-24 * time.Hour)
								atomic.AddUint64(&nFiles, 1)
							} else {
								panic(fmt.Sprintln("ERROR:", k))
							}
						}
					}
				}
			}
			wg.Done()
		}(g)
	}
	wg.Wait()
	fmt.Println("Checked", nFiles, "file pairs. Time spent on AllChecksumBC:", time.Since(tn))
}

func checkChecksum(s string) bool {
	if generateChecksum(s) == checksumFile(s) {
		//fmt.Println("files were checked, checksum is correct")
		return true
	}
	return false
}

func generateChecksum(s string) string {
	input, err := os.Open(s)
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()

	hash := sha256.New()
	if _, err1 := io.Copy(hash, input); err != nil {
		log.Fatal(err1)
	}
	sumString := fmt.Sprintf("%x", hash.Sum(nil))
	return sumString
}

func checksumFile(s string) string {
	inputCheck, err := os.ReadFile(s + ".CHECKSUM")
	if err != nil {
		log.Fatal(err)
	}
	var res string
	for _, val := range string(inputCheck) {
		if string(val) == " " {
			break
		}
		res += string(val)
	}
	return res
}
