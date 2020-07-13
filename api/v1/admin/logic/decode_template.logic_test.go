package logic

import (
	"fmt"
	"testing"
)

func TestIndexColumnCodeConvert(t *testing.T) {
	t.Run("test convert column code to index", func(t *testing.T) {
		var columnCode = "ABC"
		result := parseIndexFromColumnCode(columnCode)
		fmt.Println(result)

		columnCode = "Z"
		result = parseIndexFromColumnCode(columnCode)
		fmt.Println(result)

		columnCode = "AB"
		result = parseIndexFromColumnCode(columnCode)
		fmt.Println(result)
	})

	t.Run("test convert index to column code", func(t *testing.T) {
		var columnCode = "ABCDEFG"
		result := parseIndexFromColumnCode(columnCode)
		code := parseColumnCodeFromIndex(result)
		if code != columnCode {
			t.Errorf("expect %s but got %s\n", columnCode, code)
		}

		var columnCode2 = "Z"
		result2 := parseIndexFromColumnCode(columnCode2)
		code2 := parseColumnCodeFromIndex(result2)
		if code2 != columnCode2 {
			t.Errorf("expect %s but got %s\n", columnCode2, code2)
		}
	})
}
