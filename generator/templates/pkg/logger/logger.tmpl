package logger

import (
	"fmt"
	"runtime"
	"time"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
)

// internal utility to format message with timestamp and file/line
func format(prefix, color string, msg string, args ...any) string {
	// Capture caller info
	_, file, line, _ := runtime.Caller(2)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	header := fmt.Sprintf("%s [%s] [%s] %s:%d ==> %s", color, timestamp, prefix, file, line, reset)
	body := fmt.Sprintf(msg, args...)
	return fmt.Sprintf("%s %s", header, body)
}

func Info(msg string, args ...any) {
	fmt.Println(format("INFO", blue, msg, args...))
}

func Success(msg string, args ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	args = append([]any{timestamp}, args...)
	fmt.Printf(green+"[%v: SUCCESS] "+reset+msg+"\n", args...)
}

func Warn(msg string, args ...any) {
	fmt.Println(format("WARN", yellow, msg, args...))
}

func Error(msg string, args ...any) {
	fmt.Println(format("ERROR", red, msg, args...))
}

func Danger(msg string, args ...any) {
	fmt.Println(format("DANGER", red, msg, args...))
}

func Debug(msg string, args ...any) {
	fmt.Println(format("DEBUG", cyan, msg, args...))
}
