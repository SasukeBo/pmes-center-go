package orm

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/jinzhu/gorm"
)

func handleError(err error, key string, value interface{}) *errormap.Error {
	message := fmt.Sprintf("query with %s = %v failed: %v", key, value, err)
	if err == gorm.ErrRecordNotFound {
		return errormap.NewCodeOrigin(errormap.ErrorCodeObjectNotFound, message)
	}
	return errormap.NewCodeOrigin(errormap.ErrorCodeGetObjectFailed, message)
}
