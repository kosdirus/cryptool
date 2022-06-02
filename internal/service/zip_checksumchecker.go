package service

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

// Func checkChecksum returns true if generated checksum is equal to checksum received from Binance.
func checkChecksum(s string) bool {
	if generateChecksum(s) == checkChecksumFile(s) {
		//fmt.Println("files were checked, checksum is correct")
		return true
	}
	return false
}

// Func generateChecksum returns checksum string for file with given path.
func generateChecksum(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err1 := io.Copy(hash, file); err != nil {
		log.Fatal(err1)
	}
	sumString := fmt.Sprintf("%x", hash.Sum(nil))
	return sumString
}

// Func checkChecksumFile reads checksum string from checksum file.
func checkChecksumFile(path string) string {
	checksumFile, err := os.ReadFile(path + ".CHECKSUM")
	if err != nil {
		log.Fatal(err)
	}
	var res string
	for _, val := range string(checksumFile) {
		if string(val) == " " {
			break
		}
		res += string(val)
	}
	return res
}
