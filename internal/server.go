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

	http.HandleFunc("GET /{tableName}", validateTableNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("tableName")

		data, err := fetchTableData(db, tableName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	}))

	// http.HandleFunc("POST /{tableName}", validateTableNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
	// 	tableName := r.PathValue("tableName")

	// 	data, err := insertRowAndReturnData(db, tableName, r)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	json.NewEncoder(w).Encode(data)
	// }))

	http.HandleFunc("GET /{tableName}/{primaryKey}", validateTableNameMiddleware(func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("tableName")
		primaryKeyValue := r.PathValue("primaryKey")

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
	}))
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

// func insertRowAndReturnData(db *sql.DB, tableName string, r *http.Request) (map[string]interface{}, error) {
// 	var columnValues map[string]interface{}
// 	if err := json.NewDecoder(r.Body).Decode(&columnValues); err != nil {
// 		return nil, fmt.Errorf("error decoding request body: %v", err)
// 	}

// 	columns := []string{}
// 	placeholders := []string{}
// 	values := []interface{}{}
// 	for col, val := range columnValues {
// 		columns = append(columns, col)
// 		placeholders = append(placeholders, "?")
// 		values = append(values, val)
// 	}
// 	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
// 		tableName,
// 		strings.Join(columns, ", "),
// 		strings.Join(placeholders, ", "),
// 	)

// 	result, err := db.Exec(query, values...)
// 	if err != nil {
// 		return nil, fmt.Errorf("error inserting data: %v", err)
// 	}
// 	lastInsertId, err := result.LastInsertId()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to retrieve last insert ID: %v", err)
// 	}

// 	return fetchInsertedRowData(db, tableName, lastInsertId)

// }

// func fetchInsertedRowData(db *sql.DB, tableName string, lastInsertId int64) (map[string]interface{}, error) {
// 	// Need to call GetPrimaryKeyValue again instead???
// 	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tableName)
// 	row, err := db.Query(query, lastInsertId)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch inserted row: %v", err)
// 	}
// 	defer row.Close()

// 	if !row.Next() {
// 		return nil, sql.ErrNoRows
// 	}

// 	columns, err := row.Columns()
// 	if err != nil {
// 		return nil, err
// 	}
// 	values := make([]interface{}, len(columns))
// 	valuePtrs := make([]interface{}, len(columns))
// 	for i := range values {
// 		valuePtrs[i] = &values[i]
// 	}

// 	if err := row.Scan(valuePtrs...); err != nil {
// 		return nil, err
// 	}

// 	rowData := make(map[string]interface{})
// 	for i, col := range columns {
// 		val := *valuePtrs[i].(*interface{})
// 		rowData[col] = val
// 	}

// 	return rowData, nil
// }

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

func validateTableNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("tableName")

		if !isValidTableName(tableName) {
			http.Error(w, "Invalid table name", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	}
}
