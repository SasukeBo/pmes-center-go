package ftpclient

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func TestGenerateCSVFile(t *testing.T) {
	header := "序号,日期时间,料号,设备号,尺寸编号1,尺寸编号2,尺寸编号3,产品编号\n"
	row1 := "/,/,/,（上限）,20.03,15.03,9.03,/\n"
	row2 := "/,/,/,（下限）,19.97,14.97,8.97,/\n"
	content := fmt.Sprintf("%s%s%s", header, row1, row2)

	for i := 1; i <= 10000; i++ {
		content = content + fmt.Sprintf(
			"%d,%v,1765,1,20.011,14.991,9.021,%d\n",
			i,
			time.Now().Format("2020/03/10 14:50:28"),
			i,
		)
	}

	err := ioutil.WriteFile("/Users/wangbob/workspace/ftpserver/devicedata-test.csv", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}
