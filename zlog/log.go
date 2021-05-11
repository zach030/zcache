package zlog

import "fmt"

func Info(string2 string) {
	fmt.Printf("[go-cache] Info:%s\n", string2)
}

func Error(string2 string) {
	fmt.Printf("[go-cache] Error:%s\n", string2)
}
