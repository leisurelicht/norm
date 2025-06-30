package logger

import (
	"log"

	"github.com/leisurelicht/norm/internal/config"
)

// 自定义日志函数
func Debugf(format string, v ...any) {
	if config.IsDebug() {
		log.Printf("[DEBUG] "+format+"\n", v...)
	}
}

func Infof(format string, v ...any) {
	if config.IsInfo() {
		log.Printf("[INFO] "+format+"\n", v...)
	}
}

func Warnf(format string, v ...any) {
	if config.IsWarn() {
		log.Printf("[WARN] "+format+"\n", v...)
	}
}

func Errorf(format string, v ...any) {
	if config.IsError() {
		log.Printf("[ERROR] "+format+"\n", v...)
	}
}
