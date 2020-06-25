package utils

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"
)

var baseId string = ""

func init() {
	host, _ := os.Hostname()
	baseId = fmt.Sprintf("%x", md5.Sum([]byte(host)))
	baseId = baseId[0:6]
}

func NewId() string {
	return fmt.Sprintf("%v%v", baseId, time.Now().UnixNano())
}
