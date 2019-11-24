package utils

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// String2Int64 :nodoc:
func String2Int64(str string) int64 {
	val, err := strconv.Atoi(str)
	if err != nil {
		log.Error(err)
		return 0
	}

	return int64(val)
}

// Int64ToBytes :nodoc:
func Int64ToBytes(n int64) []byte {
	return []byte(fmt.Sprint(n))
}
