package ftpclient

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jlaffaye/ftp"
)

// connect make a connection for ftp server
func connect() (*ftp.ServerConn, error) {
	ftphostConf := orm.GetSystemConfig("ftp_host")
	if ftphostConf == nil {
		return nil, &FTPError{Message: "没有找到FTP服务器Host配置"}
	}

	ftpportConf := orm.GetSystemConfig("ftp_port")
	if ftpportConf == nil {
		return nil, &FTPError{Message: "没有找到FTP服务器Port配置"}
	}

	ftpuserConf := orm.GetSystemConfig("ftp_username")
	if ftpuserConf == nil {
		return nil, &FTPError{Message: "没有找到FTP服务器登录账号"}
	}

	ftppassConf := orm.GetSystemConfig("ftp_password")
	if ftppassConf == nil {
		return nil, &FTPError{Message: "没有找到FTP服务器登录密码"}
	}

	ftpConn, err := ftp.Dial(fmt.Sprintf("%v:%v", ftphostConf.Value, ftpportConf.Value), ftp.DialWithTimeout(4*time.Second))
	if err != nil {
		return nil, &FTPError{
			Message:   fmt.Sprintf("连接FTP服务器%s:%s失败", ftphostConf.Value, ftpportConf.Value),
			OriginErr: err,
		}
	}
	err = ftpConn.Login(ftpuserConf.Value, ftppassConf.Value)
	if err != nil {
		return nil, &FTPError{
			Message:   fmt.Sprintf("登录FTP服务器%s:%s失败", ftphostConf.Value, ftpportConf.Value),
			OriginErr: err,
		}
	}
	ftpConn.NoOp()
	return ftpConn, nil
}

// ReadFile read file by path
func ReadFile(path string) ([]byte, error) {
	ftpConn, err := connect()
	if err != nil {
		return nil, err
	}
	defer ftpConn.Quit()
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
	ftpConn, err := connect()
	if err != nil {
		return nil, err
	}
	defer ftpConn.Quit()
	var entries []string
	entries, err = ftpConn.NameList(path)
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
