package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"restique/db"
)

func NewRouter(dbConn *sql.DB) *http.ServeMux {
	router := http.NewServeMux()

	// Probably move this to api/all.go or something
	router.HandleFunc("GET /{tableName}", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("tableName")

		data, err := db.GetRecords(dbConn, tableName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	return router
}
