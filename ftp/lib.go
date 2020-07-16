package ftp

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jlaffaye/ftp"
)

// connect make a connection for ftp server
func connect() (*ftp.ServerConn, error) {
	var ftpHostConf, ftpPortConf, ftpUserConf, ftpPassConf orm.SystemConfig
	if err := ftpHostConf.GetConfig(orm.SystemConfigFtpHostKey); err != nil {
		return nil, &FTPError{Message: "没有找到FTP服务器Host配置", OriginErr: err}
	}

	if err := ftpPortConf.GetConfig(orm.SystemConfigFtpPortKey); err != nil {
		return nil, &FTPError{Message: "没有找到FTP服务器Port配置", OriginErr: err}
	}

	if err := ftpUserConf.GetConfig(orm.SystemConfigFtpUsernameKey); err != nil {
		return nil, &FTPError{Message: "没有找到FTP服务器登录账号", OriginErr: err}
	}

	if err := ftpPassConf.GetConfig(orm.SystemConfigFtpPasswordKey); err != nil {
		return nil, &FTPError{Message: "没有找到FTP服务器登录密码", OriginErr: err}
	}

	ftpConn, err := ftp.Dial(fmt.Sprintf("%v:%v", ftpHostConf.Value, ftpPortConf.Value), ftp.DialWithTimeout(4*time.Second))
	if err != nil {
		return nil, &FTPError{
			Message:   fmt.Sprintf("连接FTP服务器%s:%s失败", ftpHostConf.Value, ftpPortConf.Value),
			OriginErr: err,
		}
	}
	err = ftpConn.Login(ftpUserConf.Value, ftpPassConf.Value)
	if err != nil {
		return nil, &FTPError{
			Message:   fmt.Sprintf("登录FTP服务器%s:%s失败", ftpHostConf.Value, ftpPortConf.Value),
			OriginErr: err,
		}
	}
	ftpConn.NoOp()
	return ftpConn, nil
}

// RemoveFile 删除Ftp指定文件
func RemoveFile(path string) error {
	ftpConn, err := connect()
	if err != nil {
		return err
	}
	defer ftpConn.Quit()
	if err := ftpConn.Delete(path); err != nil {
		return &FTPError{
			Message:   fmt.Sprintf("remove ftp file(%s) failed: %v", path, err),
			OriginErr: nil,
		}
	}

	return nil
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

// GetDeepFilePath return current workspace file list
// 获取Ftp服务器上指定文件夹下的所有文件，返回路径列表
func GetDeepFilePath(path string) ([]string, error) {
	ftpConn, err := connect()
	if err != nil {
		return nil, err
	}
	defer ftpConn.Quit()
	return deepGetEntries(ftpConn, path)
}

func deepGetEntries(conn *ftp.ServerConn, path string) ([]string, error) {
	var paths []string
	entries, err := conn.List(path)
	if err != nil {
		return paths, &FTPError{
			Message:   fmt.Sprintf("获取路径%s下文件列表失败: %v", path, err),
			OriginErr: err,
		}
	}

	for _, v := range entries {
		dp := filepath.Join(path, v.Name)
		fmt.Printf("path: %s\n", dp)
		switch v.Type {
		case ftp.EntryTypeFile:
			if strings.Contains(v.Name, ".xlsx") {
				paths = append(paths, dp)
			}
		case ftp.EntryTypeFolder:
			deepPaths, err := deepGetEntries(conn, dp)
			if err != nil {
				return paths, &FTPError{
					Message:   fmt.Sprintf("获取路径%s下文件列表失败: %v", dp, err),
					OriginErr: err,
				}
			}

			paths = append(paths, deepPaths...)
		}
	}

	return paths, nil
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
