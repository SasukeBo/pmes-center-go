package ftpclient

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
	nameList, err := GetList("/")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(nameList)
}
