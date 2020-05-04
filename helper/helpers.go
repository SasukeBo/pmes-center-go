package helper

import "fmt"

func Debugf(k string, v interface{}) {
	fmt.Println("[DEBUG]———————————————————————————————————————————")
	fmt.Printf(k+": %v\n", v)
	fmt.Println("——————————————————————————————————————————————————")
}
