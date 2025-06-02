// Package log using for logging SD event
package log

// Logger struct, defines user log functions
var Logger struct {
	Debugf func(string, ...any)
	Infof  func(string, ...any)
	Errorf func(string, ...any)
	Debug  func(msg string)
}

// Debug prints message with DEBUG severity
func Debug(msg string) {
	if Logger.Debugf == nil {
		return
	}
	Logger.Debug(msg)
}

// Debugf prints formatted message with DEBUG severity
func Debugf(format string, v ...any) {
	if Logger.Debugf == nil {
		return
	}
	Logger.Debugf(format, v...)
}

// Errorf prints formatted message with ERROR severity
func Errorf(str string, v ...any) {
	if Logger.Errorf == nil {
		return
	}
	Logger.Errorf(str, v...)
}

// Infof prints formatted message with INFO severity
func Infof(str string, v ...any) {
	if Logger.Infof == nil {
		return
	}
	Logger.Infof(str, v...)
}
