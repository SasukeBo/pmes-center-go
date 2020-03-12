package conf

import (
	"fmt"
)

var (
	FtpServerHost = "localhost"
	FtpServerPort = "2121"
	FtpServerUser = "admin"
	FtpServerPass = "123456"
)

func FtpAddress() string {
	return fmt.Sprintf("%s:%s", FtpServerHost, FtpServerPort)
}
