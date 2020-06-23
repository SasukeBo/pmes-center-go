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
	nameList, err := GetList("/1828")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(nameList)
}
