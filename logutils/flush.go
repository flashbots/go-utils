package logutils

import (
	"errors"
	"html"
	"io/fs"
	"log"
	"time"

	"go.uber.org/zap"
)

// FlushZap triggers the Sync() on the logger. In case of an error, it will log it
// using the logger from standard library.
func FlushZap(logger *zap.Logger) {
	if err := logger.Sync(); err != nil {
		// Workaround for `inappropriate ioctl for device` or `invalid argument` errors
		// See: https://github.com/uber-go/zap/issues/880#issuecomment-731261906
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			if pathErr.Path == "/dev/stderr" && pathErr.Op == "sync" {
				return
			}
		}
		log.Printf(
			"{\"level\":\"error\",\"ts\":\"%s\",\"msg\":\"Failed to sync the logger\",\"error\":\"%s\"}\n",
			time.Now().Format(time.RFC3339),
			html.EscapeString(err.Error()),
		)
	}
}
