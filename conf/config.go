package conf

import (
	"fmt"
)

var (
	// FtpServerHost server host
	FtpServerHost = "localhost"
	// FtpServerPort server port
	FtpServerPort = "2121"
	// FtpServerUser ftp user name
	FtpServerUser = "admin"
	// FtpServerPass ftp user password
	FtpServerPass = "123456"
	// DBdns db dns
	DBdns = "root:123456@tcp(localhost:4476)/ftpviewer?charset=utf8&parseTime=True&loc=Local"
)

// FtpAddress ftp server address
func FtpAddress() string {
	return fmt.Sprintf("%s:%s", FtpServerHost, FtpServerPort)
}
