package norm

import "github.com/leisurelicht/norm/internal/config"

const (
	Debug = config.Debug
	Info  = config.Info
	Warn  = config.Warn
	Error = config.Error
)

func SetLevel(level config.Level) {
	config.C.Level = level
}
