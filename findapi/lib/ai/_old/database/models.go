package database

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sjsafranek/ligneous"
)

const DEFAULT_DATA_FOLDER = "."

// DataFolder is set to where you want each Sqlite3 database to be stored
var DataFolder = DEFAULT_DATA_FOLDER

// Database is the main structure for holding the information
// pertaining to the name of the database.
type Database struct {
	name           string
	family         string
	db             *sql.DB
	logger         *ligneous.Log
	isClosed       bool
	requestQueue   chan func(string)
	num_queries    int64
	lock           sync.RWMutex
	LastInsertTime time.Time
}
