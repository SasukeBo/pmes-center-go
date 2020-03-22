package ftpclient

import (
	"testing"

	"github.com/SasukeBo/ftpviewer/test"
)

func TestFetchData(t *testing.T) {
	test.ClearDB()
	test.SetConfig()
	test.SetMaterial("1828", 1)
	fetchAndStore("./1828/1828-1-20200116-b.xlsx")
}
