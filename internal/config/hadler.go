package config

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

var C = config{
	Level: Debug,
}

func IsDebug() bool {
	return C.Level == Debug
}

func IsInfo() bool {
	return C.Level == Info
}

func IsWarn() bool {
	return C.Level == Warn
}

func IsError() bool {
	return C.Level == Error
}
