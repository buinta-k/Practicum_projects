package main

import (
	"log"
	"project/pkg/api"
	"project/pkg/db"
	"project/pkg/server"
)

func main() {
	DB, err := db.Init("scheduler.db")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Conn.Close()

	api := api.NewApi(DB)
	srv := server.NewServer()
	mux := srv.Mu()
	api.Init(mux)

	if err := srv.Start(); err != nil {
		DB.Conn.Close()
		log.Fatal(err)
	}
}
