package util

import "fmt"

func Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}
