package db

import (
	"database/sql"
	"fmt"
	"regexp"
)

func GetRecords(db *sql.DB, tableName string) ([]map[string]interface{}, error) {
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRowsToMap(rows)
}

func scanRowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
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

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

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

func isValidTableName(tableName string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", tableName)
	return match
}
