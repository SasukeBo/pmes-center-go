package test

import (
	"fmt"
	"testing"
	"time"
)

func TestUtil(t *testing.T) {
	date := time.Now()
	fmt.Println(date.UTC())
	fmt.Println(date.Year())
	fmt.Println(int(date.Month()))
	fmt.Println(date.Day())
}
