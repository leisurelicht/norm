package config

import "sync"

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

type config struct {
	Level Level
}

var (
	initConfig sync.Once
	c          config
)

func ensureInit() {
	initConfig.Do(func() {
		c = config{Level: Info}
	})
}

func Get() config {
	ensureInit()
	return c
}

func SetLevel(level Level) {
	ensureInit()
	c.Level = level
}

func IsDebug() bool {
	return c.Level == Debug
}

func IsInfo() bool {
	return c.Level == Info
}

func IsWarn() bool {
	return c.Level == Warn
}

func IsError() bool {
	return c.Level == Error
}
