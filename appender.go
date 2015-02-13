package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

//LogAppender is used to push log records to a destination like a file
type LogAppender interface {
	//Log takes a record and should append it
	Log(record *LogRecord) error
	//SetLevel should remember the level assigned to this appender and check
	//it to filter incoming records
	SetLevel(l LogLevel)
	//SetFormatter should remember the formatting function that this
	//appender should use to generate strings from LogRecords
	SetFormatter(formatter LogFormatter)
}

//ClosableAppender defines an optional single method for appenders that
//need to be closed when they will not be used anymore. An example
//is a file appender.
type ClosableAppender interface {
	LogAppender
	io.Closer
}

//BaseLogAppender provides a simple struct for building log appenders.
type BaseLogAppender struct {
	m         sync.RWMutex
	level     LogLevel
	formatter LogFormatter
}

//SetLevel stores the level in the BaseLogAppender struct
func (appender *BaseLogAppender) SetLevel(l LogLevel) {
	appender.m.Lock()
	appender.level = l
	appender.m.Unlock()
}

func (appender *BaseLogAppender) checkLevel(l LogLevel) bool {
	// caller is responsible for obtaining lock
	return appender.level <= l
}

//CheckLevel tests the level in the BaseLogAppender struct
func (appender *BaseLogAppender) CheckLevel(l LogLevel) bool {
	appender.m.RLock()
	defer appender.m.RUnlock()

	return appender.checkLevel(l)
}

//SetFormatter stores the formatting function in the BaseLogAppender struct
func (appender *BaseLogAppender) SetFormatter(formatter LogFormatter) {
	appender.m.Lock()
	appender.formatter = formatter
	appender.m.Unlock()
}

func (appender *BaseLogAppender) format(record *LogRecord) string {
	// caller is responsible for obtaining lock
	formatter := appender.formatter

	if formatter == nil {
		formatter = defaultFormatter
	}

	return formatter(record.Level, record.Tags, record.Message, record.Time, record.Original)
}

//NullAppender is a simple log appender that just counts the number of log messages
type NullAppender struct {
	BaseLogAppender
	count int64
}

//NewNullAppender creates a null appender
func NewNullAppender() *NullAppender {
	return &NullAppender{}
}

//Log adds 1 to the count
func (appender *NullAppender) Log(record *LogRecord) error {
	atomic.AddInt64(&(appender.count), 1)
	return nil
}

//Count returns the count of messages logged
func (appender *NullAppender) Count() int64 {
	return atomic.LoadInt64(&(appender.count))
}

//ErrorAppender is provided for testing and will generate an error
//when asked to log a message, it will also maintain a count
type ErrorAppender struct {
	NullAppender
}

//NewErrorAppender creates an ErrorAppender
func NewErrorAppender() *ErrorAppender {
	return &ErrorAppender{}
}

//Log adds to the count and returns an error
func (appender *ErrorAppender) Log(record *LogRecord) error {
	atomic.AddInt64(&(appender.count), 1)
	return fmt.Errorf("error: %s", record.Message)
}

//ConsoleAppender can be used to write log records to standard
//err or standard out.
type ConsoleAppender struct {
	useStdout bool
	BaseLogAppender
}

//NewStdErrAppender creates a console appender configured to write to
//standard err.
func NewStdErrAppender() *ConsoleAppender {
	return &ConsoleAppender{}
}

//NewStdOutAppender creates a console appender configured to write to
//standard out.
func NewStdOutAppender() *ConsoleAppender {
	return &ConsoleAppender{useStdout: true}
}

//Log writes the record, if its level passes the appenders level
//to stderr or stdout
func (appender *ConsoleAppender) Log(record *LogRecord) error {
	appender.m.Lock()
	defer appender.m.Unlock()

	if !appender.checkLevel(record.Level) {
		return nil
	}

	if appender.useStdout {
		fmt.Fprintln(os.Stdout, appender.format(record))
	} else {
		fmt.Fprintln(os.Stderr, appender.format(record))
	}
	return nil
}

//MemoryAppender is useful for testing and keeps a list of logged messages
type MemoryAppender struct {
	BaseLogAppender
	//LoggedMesssages is the list of messages that have been logged to this appender
	LoggedMessages []string
}

//NewMemoryAppender creates a new empty memory appender
func NewMemoryAppender() *MemoryAppender {
	appender := new(MemoryAppender)
	appender.LoggedMessages = make([]string, 0, 100)
	return appender
}

//Log checks the log records level and if it passes appends the record to the list
func (appender *MemoryAppender) Log(record *LogRecord) error {
	appender.m.Lock()
	defer appender.m.Unlock()

	if !appender.checkLevel(record.Level) {
		return nil
	}

	appender.LoggedMessages = append(appender.LoggedMessages, appender.format(record))
	return nil
}

//GetLoggedMessages returns the list of logged messages as strings.
func (appender *MemoryAppender) GetLoggedMessages() []string {
	appender.m.RLock()
	defer appender.m.RUnlock()

	return appender.LoggedMessages
}

//WriterAppender is a simple appender that pushes messages as bytes to a writer
type WriterAppender struct {
	BaseLogAppender
	writer io.Writer
}

//NewWriterAppender creates an appender from the specified writer.
func NewWriterAppender(writer io.Writer) *WriterAppender {
	return &WriterAppender{writer: writer}
}

//Log checks the log record's level and then writes the formatted record
//to the writer, followed by the bytes for "\n"
func (appender *WriterAppender) Log(record *LogRecord) error {
	appender.m.Lock()
	defer appender.m.Unlock()

	if !appender.checkLevel(record.Level) {
		return nil
	}

	if appender.writer != nil {
		_, err := appender.writer.Write([]byte(appender.format(record)))
		_, err = appender.writer.Write([]byte("\n"))
		return err
	}

	return nil
}
