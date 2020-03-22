package ftpclient

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jlaffaye/ftp"
)

var ftpConn *ftp.ServerConn

func init() {
	connect()
}

// connect make a connection for ftp server
func connect() error {
	var err error
	ftphostConf := orm.GetSystemConfig("ftp_host")
	if ftphostConf == nil {
		return &FTPError{Message: "没有找到FTP服务器Host配置"}
	}

	ftpportConf := orm.GetSystemConfig("ftp_port")
	if ftpportConf == nil {
		return &FTPError{Message: "没有找到FTP服务器Port配置"}
	}

	ftpuserConf := orm.GetSystemConfig("ftp_username")
	if ftpuserConf == nil {
		return &FTPError{Message: "没有找到FTP服务器登录账号"}
	}

	ftppassConf := orm.GetSystemConfig("ftp_password")
	if ftppassConf == nil {
		return &FTPError{Message: "没有找到FTP服务器登录密码"}
	}

	ftpConn, err = ftp.Dial(fmt.Sprintf("%v:%v", ftphostConf.Value, ftpportConf.Value), ftp.DialWithTimeout(4*time.Second))
	if err != nil {
		return &FTPError{
			Message:   fmt.Sprintf("连接FTP服务器%s:%s失败", ftphostConf.Value, ftpportConf.Value),
			OriginErr: err,
		}
	}
	err = ftpConn.Login(ftpuserConf.Value, ftppassConf.Value)
	if err != nil {
		return &FTPError{
			Message:   fmt.Sprintf("登录FTP服务器%s:%s失败", ftphostConf.Value, ftpportConf.Value),
			OriginErr: err,
		}
	}
	ftpConn.NoOp()
	return nil
}

// ReadFile read file by path
func ReadFile(path string) ([]byte, error) {
	if ftpConn == nil {
		err := connect()
		if err != nil {
			return nil, err
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
		return nil, &FTPError{
			Message:   fmt.Sprintf("读取文件%s失败", path),
			OriginErr: err,
		}
	}

	buf, err := ioutil.ReadAll(res)
	if err != nil {
		log.Printf("[ReadFile] ioutil.ReadAll response failed: %v\n", err)
		return nil, &FTPError{
			Message:   fmt.Sprintf("读取文件%s失败", path),
			OriginErr: err,
		}
	}

	return buf, nil
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
	if err != nil {
		return entries, &FTPError{
			Message:   fmt.Sprintf("获取路径%s下文件列表失败", path),
			OriginErr: err,
		}
	}

	return entries, nil
}

// FTPError _
type FTPError struct {
	Message   string
	OriginErr error
}

// Error _
func (e *FTPError) Error() string {
	return e.Message
}

// Logger _
func (e *FTPError) Logger() {
	log.Printf("%s, originErr: %v", e.Message, e.OriginErr)
}
