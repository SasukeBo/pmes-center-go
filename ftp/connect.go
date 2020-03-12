package ftpconn

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"io"
	"log"
	"time"
)

var ftpConn *ftp.ServerConn

func init() {
	var err error
	ftpConn, err = ftp.Dial("192.168.8.119:21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Printf("[ftp.Dial] failed with dial ftp server:\n%v\n", err)
		return
	}
	err = ftpConn.Login("anonymous", "anonymous")
	if err != nil {
		log.Printf("[ftpConn.Login] failed:\n%v\n", err)
		return
	}
	ftpConn.NoOp()
}

func readFile(file string) io.Reader {
	if ftpConn == nil {
		return nil
	}

	r, err := ftpConn.Retr(file)
	if err != nil {
		log.Printf("[c.Retr] with file(%v) failed:\n%v\n", file, err)
		return nil
	}

	return r
}

func getList() {
	dir, err := ftpConn.CurrentDir()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("current dir is", dir)
	entries, err := ftpConn.NameList("/")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(entries)
	return
}
