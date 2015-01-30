// +build !windows

package logging

import (
	"log/syslog"
)

//SysLogAppender is the logging appender for appending to the syslog service
type SysLogAppender struct {
	BaseLogAppender
	syslogger *syslog.Writer
}

/*
NewSysLogAppender creates a sys log appender
*/
func NewSysLogAppender() *SysLogAppender {
	appender := new(SysLogAppender)
	appender.level = DEFAULT
	return appender
}

/*
Log adds a record to the sys log
*/
func (appender *SysLogAppender) Log(record *LogRecord) error {

	if !appender.CheckLevel(record.Level) {
		return nil
	}

	if appender.syslogger == nil {
		logWriter, e := syslog.New(syslog.LOG_DEBUG, "")

		if e == nil {
			appender.syslogger = logWriter
		} else {
			return e
		}
	}

	if appender.syslogger != nil {

		formatted := appender.format(record)

		switch record.Level {
		case DEBUG:
			return appender.syslogger.Debug(formatted)
		case INFO:
			return appender.syslogger.Info(formatted)
		case WARN:
			return appender.syslogger.Warning(formatted)
		case ERROR:
			return appender.syslogger.Err(formatted)
		default:
			return appender.syslogger.Debug(formatted)
		}
	}

	return nil
}

//Close shuts down the syslog connection
func (appender *SysLogAppender) Close() error {

	if appender.syslogger != nil {
		return appender.syslogger.Close()
	}
	return nil
}
