package logic

import (
	"fmt"
	"testing"
	"time"

	// "github.com/SasukeBo/ftpviewer/test"
)

func TestIsMaterialExist(t *testing.T) {
	fmt.Printf("material 1765 exist: %v\n", IsMaterialExist("1766"))
}

func TestFetchMaterialDatas(t *testing.T) {
	t1, _ := time.Parse(time.RFC3339, "2020-03-01T00:00:00+08:00")
	t2, _ := time.Parse(time.RFC3339, "2020-03-20T00:00:00+08:00")
	FetchMaterialDatas("1765", &t1, &t2)
}

func TestRMSError(t *testing.T) {
	data := []float64{200, 50, 100, 200}
	r := RMSError(data)
	if r != 75 {
		t.Fatalf("expect 75, but got %v", r)
	}
}
