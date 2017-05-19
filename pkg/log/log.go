package log

import (
	"fmt"
	"strings"
)

func Debug(formate string, args ...interface{}) {
	v := fmt.Sprintf(formate, args...)
	p(v)
}

func Info(formate string, args ...interface{}) {
	v := fmt.Sprintf(formate, args...)
	p(v)
}

func Warn(formate string, args ...interface{}) {
	v := fmt.Sprintf(formate, args...)
	p(v)
}

func Error(formate string, args ...interface{}) {
	v := fmt.Sprintf(formate, args...)
	p(v)
}

func p(value string) {
	v := strings.Trim(value, "\n")
	fmt.Println(v)
}


