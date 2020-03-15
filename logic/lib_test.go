package logic

import (
	"fmt"
	"testing"
	// "github.com/SasukeBo/ftpviewer/test"
)

func TestIsMaterialExist(t *testing.T) {
	fmt.Printf("material 1765 exist: %v\n", IsMaterialExist("1765"))
}
