package service

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

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
