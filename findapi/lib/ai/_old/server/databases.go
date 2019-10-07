package server

import (
	"time"

	"github.com/schollz/find4/server/main/src/api"
	"github.com/schollz/find4/server/main/src/database"
)

var (
	DATABASES map[string]*database.Database
)

func OpenDatabase(family string) error {
	db_conn, err := database.Open(family, false)
	if nil != err {
		return err
	}
	DATABASES[family] = db_conn

	// control for server shutdowns and crashs
	// make sure calibration occurs on database startup
	go api.DatabaseWorker(db_conn, family)

	return nil
}

func GetDatabase(family string) (*database.Database, error) {
	if _, ok := DATABASES[family]; !ok {
		err := OpenDatabase(family)
		return DATABASES[family], err
	}
	return DATABASES[family], nil
}

func DeleteDatabase(family string) error {
	db, err := GetDatabase(family)
	if nil != err {
		return err
	}
	db.Delete()
	DATABASES[family].Close()
	delete(DATABASES, family)
	return nil
}

func init() {
	DATABASES = make(map[string]*database.Database)

	// debugging goroutine to report database write queues
	go func() {
		for {
			time.Sleep(10 * time.Second)

			c := 0
			for range DATABASES {
				c++
			}

			if 0 != c {
				logger.Debugf("%v active databases", c)
				for family := range DATABASES {
					pending := DATABASES[family].GetPending()
					if 0 != pending {
						logger.Debugf("%v requests in %v queue", pending, family)
					}
				}
			}
		}
	}()

}

// Shutdown closes databases for a graceful shutdown
func Shutdown() {
	for family := range DATABASES {
		logger.Warnf("Closing %v database", family)
		DATABASES[family].Close()
	}
}
