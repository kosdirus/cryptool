package main

import (
	"fmt"
	"github.com/kosdirus/cryptool/pkg/api"
	"github.com/kosdirus/cryptool/pkg/binanceapi"
	"github.com/kosdirus/cryptool/pkg/db"
	"log"
	"net/http"
	"os"
)

func main() {
	pgdb, err := db.NewDB()
	if err != nil {
		panic(err)
	}

	go binanceapi.BinanceAPISchedule(pgdb)

	router := api.NewAPI(pgdb)

	log.Print("we're up and running!")
	port := "80"
	if os.Getenv("ENV") == "LOCAL" {
		port = os.Getenv("IPORT")
	}
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		log.Println("error from router", err)
	}
	log.Print("we're up and running! 2")

}
