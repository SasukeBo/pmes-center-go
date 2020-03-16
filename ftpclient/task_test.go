package ftpclient

import (
	"testing"

	"github.com/SasukeBo/ftpviewer/test"
)

func TestFetchData(t *testing.T) {
	defer test.ClearDB()
	test.SetConfig()
	fetchAndStore("./1765/1765-1-20200312-w.csv")
}
