package server

import (
	"fmt"
	"log"
	"net/http"
	"restique/api"
	"restique/db"
)

func StartServer(dbFile string, port int) {
	dbConn, err := db.InitializeDB(dbFile)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	router := api.NewRouter(dbConn)

	address := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", address)

	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
