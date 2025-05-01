package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var baseLogger *zap.Logger

// Init sets up the global logger. If debug is true, logs everything.
// Otherwise, logs only Info and above.
func Init(debug bool) {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "module",
		MessageKey:       "msg",
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		EncodeName:       modulePrefixEncoder,
		ConsoleSeparator: " ",
	}

	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	baseLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// Named creates a new logger with the specified module name.
func Named(module string) *zap.SugaredLogger {
	return baseLogger.Named(module).Sugar()
}

func modulePrefixEncoder(name string, enc zapcore.PrimitiveArrayEncoder) {
	if name != "" {
		enc.AppendString("[" + name + "]")
	}
}
