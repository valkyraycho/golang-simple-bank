package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/valkyraycho/bank/api"
	db "github.com/valkyraycho/bank/db/sqlc"
)

const (
	dbDriver      = "postgres"
	dbSource      = "postgresql://postgres:postgres@localhost:5432/bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	server := api.NewServer(db.NewStore(conn))
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start the server: ", err)
	}
}
