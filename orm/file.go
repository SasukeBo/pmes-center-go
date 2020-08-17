package orm

import (
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

const (
	XlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	// 文件存储二级目录
	DirPrivate = "priv"
	DirUpload = "post"
	DirSource = "source"
	DirCache = "cache"
)

// 文件 model，用于存储文件信息。

type File struct {
	gorm.Model
	Name        string `gorm:"not null"`              // 文件名称
	Path        string `gorm:"not null"`              // 存储路径
	Token       string `gorm:"not null;unique_index"` // 文件唯一标识
	UserID      uint   `gorm:"column:user_id"`
	Size        uint   // 文件大小
	ContentType string `gorm:"not null"` // 文件内容类型
}

func (f *File) BeforeCreate() error {
	if f.Token == "" {
		uid, err := uuid.NewRandom()
		if err != nil {
			return err
		}

		f.Token = uid.String()
	}

	return nil
}

func (f *File) GetByToken(token string) *errormap.Error {
	if err := DB.Model(f).Where("token = ?", token).First(f).Error; err != nil {
		return handleError(err, "token", token)
	}

	return nil
}

func (f *File) Get(id uint) *errormap.Error {
	if err := DB.Model(f).Where("id = ?", id).First(f).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}
