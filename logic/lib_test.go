package logic

import (
	"fmt"
	"testing"
	// "github.com/SasukeBo/ftpviewer/test"
)

func TestIsMaterialExist(t *testing.T) {
	fmt.Printf("material 1765 exist: %v\n", IsMaterialExist("1766"))
}

func TestRMSError(t *testing.T) {
	data := []float64{200, 50, 100, 200}
	r := solveRmsError(data)
	if r != 75 {
		t.Fatalf("expect 75, but got %v", r)
	}
}
