package logutils

import (
	"fmt"
	"strings"

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

// GetZapLogger returns a logger created according to `log-level` and
// `log-dev` command-line switches.
//
// Note: this method defines two flags, `log-level` and `log-dev`; redefining
// them in your main app will cause panic.
func GetZapLogger(options ...LogConfigOption) *zap.Logger {
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

	level, levelErr := zap.ParseAtomicLevel(cfg.level)
	if levelErr != nil {
		level = zap.NewAtomicLevel()
	}
	config.Level = level

	l, buildErr := config.Build()
	if buildErr != nil {
		l = zap.L()
	}

	if levelErr != nil {
		valid := "'" + strings.Join([]string{
			zap.DebugLevel.String(),
			zap.InfoLevel.String(),
			zap.WarnLevel.String(),
			zap.ErrorLevel.String(),
			zap.FatalLevel.String(),
			zap.PanicLevel.String(),
		}, "', '") + "'"
		l.Warn(
			fmt.Sprintf("Invalid log-level was specified (defaulted to '%s', valid levels are: %s)", config.Level.String(), valid),
			zap.Error(levelErr),
		)
	}
	if buildErr != nil {
		l.Error("Failed to build the logger with desired configuration",
			zap.Error(buildErr),
		)
	}

	return l
}
