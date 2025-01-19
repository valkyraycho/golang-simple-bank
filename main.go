package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/valkyraycho/bank/api"
	db "github.com/valkyraycho/bank/db/sqlc"
	"github.com/valkyraycho/bank/utils"
)

func main() {
	cfg, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}
	conn, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	server := api.NewServer(db.NewStore(conn))
	err = server.Start(cfg.ServerAddress)
	if err != nil {
		log.Fatal("cannot start the server: ", err)
	}
}
