package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"restique/server"

	"github.com/spf13/cobra"
)

var dbFile string
var port int

var rootCmd = &cobra.Command{
	Use:   "restique [SQLite file]",
	Short: "A REST API server for SQLite databases",
	Long:  `restique is a REST API server for interfacing with SQLite databases.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbFile = args[0]
		if !isValidExtension(dbFile) {
			fmt.Fprintf(os.Stderr, "Invalid file extension for database file: %s\n", dbFile)
			os.Exit(1)
		}
		server.StartServer(dbFile, port)
	},
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "Port on which the server will run")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isValidExtension(fileName string) bool {
	allowedExtensions := []string{"sqlite", "sqlite3", "db", "db3", "s3db", "sl3"}
	ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return true
		}
	}
	return false
}
