package main

import (
	"github.com/kosdirus/cryptool/internal/storage/psql"
	"github.com/kosdirus/cryptool/internal/transport/rest"
	"log"
)

// main is entry point to app, it calls func NewDB to connect to DB
// and then passes it as argument to function EchoApiServer that
// starts HTTP server
func main() {
	pgdb, err := psql.NewDB()
	if err != nil {
		log.Println("Error occurred while connecting to Postgres:", err)
		return
	}

	rest.EchoApiServer(pgdb)

	log.Print("Something went wrong, HTTP server didn't start.")
}
