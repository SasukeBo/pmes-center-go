package ftpclient

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

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

