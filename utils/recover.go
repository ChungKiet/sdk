package utils

import "fmt"

func Recover() {
	if err := recover(); err != nil {
		fmt.Println("Recovered from panic: ", err)
	}
}