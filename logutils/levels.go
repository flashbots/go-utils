package logutils

import "go.uber.org/zap"

var Levels = []string{
	zap.DebugLevel.String(),
	zap.InfoLevel.String(),
	zap.WarnLevel.String(),
	zap.ErrorLevel.String(),
	zap.DPanicLevel.String(),
	zap.PanicLevel.String(),
	zap.FatalLevel.String(),
}
