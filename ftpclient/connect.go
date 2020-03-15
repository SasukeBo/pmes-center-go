package ftpclient

import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jlaffaye/ftp"
	"io/ioutil"
	"log"
	"time"
)

var ftpConn *ftp.ServerConn

func init() {
	connect()
}

// connect make a connection for ftp server
func connect() error {
	var err error
	ftphostConf := orm.GetSystemConfigCache("ftp_host")
	if ftphostConf == nil {
		return errors.New("没有找到FTP服务器Host配置")
	}

	ftpportConf := orm.GetSystemConfigCache("ftp_port")
	if ftpportConf == nil {
		return errors.New("没有找到FTP服务器Port配置")
	}

	ftpuserConf := orm.GetSystemConfigCache("ftp_username")
	if ftpuserConf == nil {
		return errors.New("没有找到FTP服务器登录账号")
	}

	ftppassConf := orm.GetSystemConfigCache("ftp_password")
	if ftppassConf == nil {
		return errors.New("没有找到FTP服务器登录密码")
	}

	ftpConn, err = ftp.Dial(fmt.Sprintf("%v:%v", ftphostConf.Value, ftpportConf.Value), ftp.DialWithTimeout(4*time.Second))
	if err != nil {
		return err
	}
	err = ftpConn.Login(ftpuserConf.Value, ftppassConf.Value)
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
