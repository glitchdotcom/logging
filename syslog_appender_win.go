// +build windows

package logging

import (
	"errors"
)

type SysLogAppender struct {
	BaseLogAppender
}

func NewSysLogAppender() *SysLogAppender {
	panic(errors.New("Syslog is not supported on Windows"))
	return nil
}

func (appender *SysLogAppender) Log(record *LogRecord) error {

	if !appender.CheckLevel(record.Level) {
		return nil
	}

	return errors.New("Syslog is not supported on Windows")
}

func (appender *SysLogAppender) Close() error {
	return nil
}
