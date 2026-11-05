package main

import (
	"log"
	"project/internal/api"
	"project/internal/db"
	"project/internal/server"
)

func main() {
	database, err := db.Init("demo.db")
	if err != nil {
		log.Fatal(err)

	}
	defer database.Conn.Close()

	srv := server.NewServer()

	mux := srv.Mux()

	Api := api.Init(database)

	Api.Init(mux)

	err = srv.Start()
	if err != nil {
		log.Fatal(err)
	}

}
