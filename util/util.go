package util

import (
	"crypto/md5"
	"fmt"
	"github.com/SasukeBo/log"
	"time"
)

// Encrypt _
func Encrypt(origin string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(origin)))
}

// NowDateStr _
func NowDateStr() string {
	var tStr = time.Now().String()
	return tStr[:10]
}

func Includes(items []string, target string) bool {
	for _, e := range items {
		if e == target {
			return true
		}
	}

	return false
}

func DebugTime(t time.Time, name string) time.Time {
	nt := time.Now()
	log.Info("[Debug] %s spend %v", name, nt.Sub(t))
	return nt
}
