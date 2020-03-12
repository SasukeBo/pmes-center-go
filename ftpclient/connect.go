package ftpclient

import (
	"github.com/SasukeBo/ftpviewer/conf"
	"github.com/jlaffaye/ftp"
	"io/ioutil"
	"log"
	"time"
)

var ftpConn *ftp.ServerConn

func init() {
	err := connect()
	if err != nil {
		panic(err)
	}
}

// connect make a connection for ftp server
func connect() error {
	var err error
	ftpConn, err = ftp.Dial(conf.FtpAddress(), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	err = ftpConn.Login(conf.FtpServerUser, conf.FtpServerPass)
	if err != nil {
		return err
	}
	ftpConn.NoOp()
	return nil
}

// ReadFile read file by path
func ReadFile(path string) (string, error) {
	if ftpConn == nil {
		err := connect()
		if err != nil {
			return "", err
		}
	}
	res, err := ftpConn.Retr(path)
	defer func() {
		if res != nil {
			res.Close()
		}
	}()
	if err != nil {
		log.Printf("[c.Retr] with file(%v) failed:\n%v\n", path, err)
		return "", err
	}

	buf, err := ioutil.ReadAll(res)
	if err != nil {
		log.Printf("[ReadFile] ioutil.ReadAll response failed: %v\n", err)
		return "", err
	}

	return string(buf), nil
}

// GetList return current workspace file list
func GetList(path string) ([]string, error) {
	var entries []string
	if ftpConn == nil {
		err := connect()
		if err != nil {
			return entries, err
		}
	}

	entries, err := ftpConn.NameList(path)
	return entries, err
}
