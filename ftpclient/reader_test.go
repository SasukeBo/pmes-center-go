package ftpclient

import (
	"fmt"
	"io/ioutil"
	"regexp"
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
			timeFormatReverse(time.Now()),
			i,
		)
	}

	err := ioutil.WriteFile("/Users/sasuke/workspace/ftpdata/1765/1765-1-20200312-w.csv", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

var timePatternReverse = `(\d{4})-(\d{2})-(\d{2}) (\d{2}:\d{2}:\d{2})`

func TestTimeFormat(t *testing.T) {
	fmt.Println(timeFormatReverse(time.Now()))
}

func timeFormatReverse(t time.Time) string {
	r := regexp.MustCompile(timePatternReverse)
	re := r.FindAllStringSubmatch(t.String(), -1)
	if len(re) > 0 {
		return fmt.Sprintf("%s/%s/%s %s", re[0][1], re[0][2], re[0][3], re[0][4])
	}
	return ""
}

func TestDecode(t *testing.T) {
	datas, err := ReadFile("./1765/1765-1-20200312-w.csv")
	if err != nil {
		t.Fatal(err)
	}

	var cd CSVDecoder
	if err := cd.Decode([]byte(datas)); err != nil {
		t.Fatal(err)
	}

	fmt.Println("Header:")
	fmt.Println(cd.Headers)

	fmt.Println("Limits:")
	fmt.Println(cd.Limits)

	fmt.Println("Rows:")
	for i := 0; i < 10; i++ {
		fmt.Println(cd.Rows[i])
	}
	fmt.Println("...")
}
