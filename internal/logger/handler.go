package logger

import (
	"log"

	"github.com/leisurelicht/norm/internal/config"
)

// 自定义日志函数
func Debug(format string, v ...any) {
	if config.IsDebug() {
		log.Printf("[DEBUG] "+format+"\n", v...)
	}
}

func Info(format string, v ...any) {
	if config.IsInfo() {
		log.Printf("[INFO] "+format+"\n", v...)
	}
}

func Warn(format string, v ...any) {
	if config.IsWarn() {
		log.Printf("[WARN] "+format+"\n", v...)
	}
}

func Error(format string, v ...any) {
	if config.IsError() {
		log.Printf("[ERROR] "+format+"\n", v...)
	}
}
