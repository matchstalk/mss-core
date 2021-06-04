package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/matchstalk/mss-core/logger"
)

// Options 参数配置
type Options struct {
	logger.Options
	Formatter logrus.Formatter
	Hooks     logrus.LevelHooks
	// Flag for whether to log caller info (off by default)
	ReportCaller bool
	// Exit Function to call when FatalLevel log
	ExitFunc func(int)
}

type formatterKey struct{}

// WithFormatter set formatter
func WithFormatter(f logrus.Formatter) logger.Option {
	return logger.SetOption(formatterKey{}, f)
}

type hooksKey struct{}

// WithLevelHooks set hook
func WithLevelHooks(hooks logrus.LevelHooks) logger.Option {
	return logger.SetOption(hooksKey{}, hooks)
}

type reportCallerKey struct{}

// ReportCaller warning to use this option. because logrus doest not open CallerDepth option
// this will only print this package
func ReportCaller() logger.Option {
	return logger.SetOption(reportCallerKey{}, true)
}

type exitKey struct{}

func WithExitFunc(exit func(int)) logger.Option {
	return logger.SetOption(exitKey{}, exit)
}

type logrusLoggerKey struct{}

func WithLogger(l logrus.StdLogger) logger.Option {
	return logger.SetOption(logrusLoggerKey{}, l)
}
