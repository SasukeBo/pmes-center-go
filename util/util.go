package util

import (
	"crypto/md5"
	"fmt"
)

// Encrypt _
func Encrypt(origin string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(origin)))
}
