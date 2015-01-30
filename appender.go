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
	Close() error
}

//BaseLogAppender provides a simple struct for building log appenders.
type BaseLogAppender struct {
	level     LogLevel
	formatter LogFormatter
	mutex     sync.RWMutex
}

//SetLevel stores the level in the BaseLogAppender struct
func (appender *BaseLogAppender) SetLevel(l LogLevel) {
	appender.mutex.Lock()
	defer appender.mutex.Unlock()

	appender.level = l
}

func (appender *BaseLogAppender) checkLevel(l LogLevel) bool {
	// caller is responsible for obtaining lock
	return appender.level <= l
}

//CheckLevel tests the level in the BaseLogAppender struct
func (appender *BaseLogAppender) CheckLevel(l LogLevel) bool {
	appender.mutex.RLock()
	defer appender.mutex.RUnlock()

	return appender.checkLevel(l)
}

//SetFormatter stores the formatting function in the BaseLogAppender struct
func (appender *BaseLogAppender) SetFormatter(formatter LogFormatter) {
	appender.mutex.Lock()
	defer appender.mutex.Unlock()

	appender.formatter = formatter
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
	appender := new(NullAppender)
	appender.level = DEFAULT
	return appender
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
	appender := new(ErrorAppender)
	appender.level = DEFAULT
	return appender
}

//Log adds to the count and returns an error
func (appender *ErrorAppender) Log(record *LogRecord) error {
	atomic.AddInt64(&(appender.count), 1)
	return fmt.Errorf("error: %s", record.Message)
}

//ConsoleAppender can be used to write log records to standard
//err or standard out.
type ConsoleAppender struct {
	useStderr bool
	BaseLogAppender
}

//NewStdErrAppender creates a console appender configured to write to
//standard err.
func NewStdErrAppender() *ConsoleAppender {
	appender := new(ConsoleAppender)
	appender.level = DEFAULT
	appender.useStderr = true
	return appender
}

//NewStdOutAppender creates a console appender configured to write to
//standard out.
func NewStdOutAppender() *ConsoleAppender {
	appender := new(ConsoleAppender)
	appender.level = DEFAULT
	appender.useStderr = false
	return appender
}

//Log writes the record, if its level passes the appenders level
//to stderr or stdout
func (appender *ConsoleAppender) Log(record *LogRecord) error {
	appender.mutex.Lock()
	defer appender.mutex.Unlock()

	if !appender.checkLevel(record.Level) {
		return nil
	}

	if appender.useStderr {
		fmt.Fprintln(os.Stderr, appender.format(record))
	} else {
		fmt.Fprintln(os.Stdout, appender.format(record))
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
	appender.level = DEFAULT
	appender.LoggedMessages = make([]string, 0, 100)
	return appender
}

//Log checks the log records level and if it passes appends the record to the list
func (appender *MemoryAppender) Log(record *LogRecord) error {
	appender.mutex.Lock()
	defer appender.mutex.Unlock()

	if !appender.checkLevel(record.Level) {
		return nil
	}

	appender.LoggedMessages = append(appender.LoggedMessages, appender.format(record))
	return nil
}

//GetLoggedMessages returns the list of logged messages as strings.
func (appender *MemoryAppender) GetLoggedMessages() []string {
	appender.mutex.RLock()
	defer appender.mutex.RUnlock()

	return appender.LoggedMessages
}

//WriterAppender is a simple appender that pushes messages as bytes to a writer
type WriterAppender struct {
	BaseLogAppender
	writer io.Writer
}

//NewWriterAppender creates an appender from the specified writer.
func NewWriterAppender(writer io.Writer) *WriterAppender {
	appender := new(WriterAppender)
	appender.level = DEFAULT
	appender.writer = writer
	return appender
}

//Log checks the log record's level and then writes the formatted record
//to the writer, followed by the bytes for "\n"
func (appender *WriterAppender) Log(record *LogRecord) error {
	appender.mutex.Lock()
	defer appender.mutex.Unlock()

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
