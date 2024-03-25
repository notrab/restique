package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

func StartServer(dbFile string, port int) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	setupRoutes(db)

	address := fmt.Sprintf(":%d", port)

	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(db *sql.DB) {
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tables, err := fetchTables(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(tables)
	})

	http.HandleFunc("GET /{tableName}", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("tableName")

		if !isValidTableName(tableName) {
			http.Error(w, "Invalid table name", http.StatusBadRequest)
			return
		}

		data, err := fetchTableData(db, tableName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	http.HandleFunc("GET /{tableName}/{primaryKey}", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("tableName")
		primaryKeyValue := r.PathValue("primaryKey")

		if !isValidTableName(tableName) {
			http.Error(w, "Invalid table name", http.StatusBadRequest)
			return
		}

		primaryKeyColumn, err := GetPrimaryKeyColumn(db, tableName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := fetchTableRowData(db, tableName, primaryKeyColumn, primaryKeyValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})
}

func fetchTables(db *sql.DB) ([]string, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table';"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func fetchTableData(db *sql.DB, tableName string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var tableData []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)

		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				rowData[col] = string(b)
			} else {
				rowData[col] = val
			}
		}

		tableData = append(tableData, rowData)
	}

	return tableData, nil
}

func fetchTableRowData(db *sql.DB, tableName, primaryKeyColumn, primaryKeyValue string) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", tableName, primaryKeyColumn)
	rows, err := db.Query(query, primaryKeyValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	rowData := make(map[string]interface{})
	for i, col := range columns {
		var val interface{}
		b, ok := values[i].([]byte)
		if ok {
			val = string(b)
		} else {
			val = values[i]
		}
		rowData[col] = val
	}

	return rowData, nil
}

func isValidTableName(name string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", name)
	return match
}

func GetPrimaryKeyColumn(db *sql.DB, tableName string) (string, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var (
		cid        int
		name       string
		ctype      string
		notnull    int
		dflt_value *string
		pk         int
	)
	for rows.Next() {
		err = rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk)
		if err != nil {
			return "", err
		}
		if pk == 1 {
			return name, nil
		}
	}
	return "", fmt.Errorf("no primary key found for table %s", tableName)
}
