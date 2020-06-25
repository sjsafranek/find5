package database

import (
	"strings"

	"github.com/sjsafranek/logger"
)

func checkError(err error) error {
	if strings.Contains(err.Error(), "driver.Value type <nil>") {
		logger.Trace("No rows found")
		return nil
	}
	return err
}
