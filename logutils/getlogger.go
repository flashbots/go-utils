package logutils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerConfig struct {
	devMode bool
	level   string
}

// LogConfigOption allows to fine-tune the configuration of the logger.
type LogConfigOption = func(*loggerConfig)

// LogDevMode tells the logger to work in the development mode.
func LogDevMode(devMode bool) LogConfigOption {
	return func(lc *loggerConfig) {
		lc.devMode = devMode
	}
}

// LogLevel sets the desired level of logging.
func LogLevel(level string) LogConfigOption {
	return func(lc *loggerConfig) {
		lc.level = level
	}
}

// GetZapLogger returns a logger created according to the provided options. In
// case if anything goes wrong (for example if the log-level string can not be
// parsed) it will return a logger (with configuration that is closest possible
// to the desired one) and an error.
func GetZapLogger(options ...LogConfigOption) (*zap.Logger, error) {
	cfg := &loggerConfig{
		devMode: false,
		level:   zap.InfoLevel.String(),
	}

	for _, o := range options {
		o(cfg)
	}

	var config zap.Config
	if cfg.devMode {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Test the logger build as per the configuration we have so far
	// (we want to know if anything is wrong as early as possible)
	basicLogger, err := config.Build()
	if err != nil {
		return zap.L(), err // Return global logger for MustGetZapLogger sake
	}

	// Parse the log level
	level, err := zap.ParseAtomicLevel(cfg.level)
	if err != nil {
		return basicLogger, err // basicLogger is there already, so let's use it
	}
	config.Level = level

	// Build the final config of the logger
	finalLogger, err := config.Build()
	if err != nil {
		return basicLogger, err
	}

	return finalLogger, nil
}

// MustGetZapLogger is guaranteed to return a logger with configuration as close
// as possible to the desired one. Any errors encountered in the process will be
// logged as warnings with the resulting logger.
func MustGetZapLogger(options ...LogConfigOption) *zap.Logger {
	l, err := GetZapLogger(options...)
	if l == nil {
		l = zap.L()
	}
	if err != nil {
		l.Warn("Error while building the logger", zap.Error(err))
	}
	return l
}
