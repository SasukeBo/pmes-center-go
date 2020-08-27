package ftp

import (
	"fmt"
	"testing"
)

func TestReadFile(t *testing.T) {
	content, err := ReadFile("./testfile")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(content)
}

func TestGetList(t *testing.T) {
	nameList, err := GetDeepFilePath("/1828-BZ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(nameList)
}

func TestRemoveFile(t *testing.T) {
	err := RemoveFile("/1828/def/ghi/1828-EDAC_E568_1-20200427-b.xls")
	if err != nil {
		fmt.Println(err)
	}
}
