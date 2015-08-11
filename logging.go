//Package logging for Go takes a slightly different approach based on our experience with real world logging.
//
//A default logger provides an easy path to logging.
//Named loggers provide a way to set levels by package.
//Tags provide a way to set levels across concepts, perpendicular to logger names.
//Default log levels and default tag levels are independent of the default logger.
//All loggers share the same appenders - but appenders can be associated with a level which is unrelated to tags.
//Each logger has an optional buffer, that will be flushd whenever its level/tags change.
//This buffer contains un-passed messages. So that it is possible to configure the system to capture messages and replay them later.
//Replayed messages are tagged and have a double time stamp.
//A default appender is initialized to send log messages to stderr.
package logging

import (
	"container/ring"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"strings"
)

//Logger is the interface for the objects that are the target of logging messages. Logging methods
//imply a level. For example, Info() implies a level of LogLevel.INFO.
type Logger interface {
	PanicWithTagsf(tags []string, fmt string, args ...interface{})
	PanicWithTags(tags []string, args ...interface{})
	Panicf(fmt string, args ...interface{})
	Panic(args ...interface{})

	ErrorWithTagsf(tags []string, fmt string, args ...interface{})
	ErrorWithTags(tags []string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Error(args ...interface{})

	WarnWithTagsf(tags []string, fmt string, args ...interface{})
	WarnWithTags(tags []string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Warn(args ...interface{})

	InfoWithTagsf(tags []string, fmt string, args ...interface{})
	InfoWithTags(tags []string, args ...interface{})
	Infof(fmt string, args ...interface{})
	Info(args ...interface{})

	DebugWithTagsf(tags []string, fmt string, args ...interface{})
	DebugWithTags(tags []string, args ...interface{})
	Debugf(fmt string, args ...interface{})
	Debug(args ...interface{})

	VerboseWithTagsf(tags []string, fmt string, args ...interface{})
	Verbosef(fmt string, args ...interface{})

	SetLogLevel(l LogLevel)
	SetTagLevel(tag string, l LogLevel)
	CheckLevel(l LogLevel, tags []string) bool

	SetBufferLength(length int)
}

const (
	stopped = iota
	paused
	running
)

//logMutex is a global lock for protecting all global state in package
var logMutex = new(sync.RWMutex)

//defaultLogger is provided for most logging situations
var defaultLogger *LoggerImpl

//The default format is used to determine how appenders without a custom format log their messages
var defaultFormatter = GetFormatter(FULL)

//Loggers share the appenders
var appenders = make([]LogAppender, 0)

//The package maintains a map of named loggers
var loggers = make(map[string]*LoggerImpl)
var incomingChannel = make(chan *LogRecord, 2048)
var stateChannel = make(chan int, 0)
var waiter = new(sync.WaitGroup)
var logged uint64
var processed uint64
var logErrors chan<- error
var enableVerbose int32

func init() {
	defaultLogger = new(LoggerImpl)
	defaultLogger.name = "_default"
	defaultLogger.level = INFO
	defaultLogger.SetBufferLength(0)

	AddAppender(NewStdErrAppender())
	AdaptStandardLogging(INFO, nil)

	go processIncoming()
}

//LogRecord is the type used in the logging buffer
type LogRecord struct {
	//Time is the time that the log record is being appended, can be
	//different from Original if the record was buffered
	Time time.Time
	//Original is the original time for the log record
	Original time.Time
	//Level is the level the record was logged at
	Level LogLevel
	//Tags are the custom tags assigned to the record when it was logged
	Tags []string
	//Message is the actual log message
	Message string
	//Logger is the logger associated with this log record, if any
	Logger *LoggerImpl
}

//LoggerImpl stores the data for a logger.
//A Logger maintains its own level, tag levels and buffer. Each logger is named.
type LoggerImpl struct {
	name      string
	level     LogLevel
	tagLevels map[string]LogLevel
	buffer    *ring.Ring
}

//PauseLogging stops all logging from being processed.
//Pause will not wait for all log messages to be processed
func PauseLogging() {
	stateChannel <- paused
}

//RestartLogging starts messages logging again
func RestartLogging() {
	stateChannel <- running
}

//StopLogging can only be called once, and completely stops the logging
//process
func StopLogging() {
	stateChannel <- stopped
	waiter.Wait()
}

func processIncoming() {
loop:
	for {
		select {
		case record := <-incomingChannel:
			processLogRecord(record)
		case newState := <-stateChannel:
			switch newState {
			case stopped:
				waiter.Done()
				break loop
			case paused: //run a sub-loop looking for a state change
			subloop:
				for {
					select {
					case state := <-stateChannel:
						switch state {
						case stopped:
							waiter.Done()
							break loop
						case running:
							break subloop
						default:
							continue
						}
					}
				}
			}
		}
	}
}

//WaitForIncoming should be used in tests or system shutdowns to make sure
//that all of the log messages pushed into the logging channel are processed
//and appended appropriately.
func WaitForIncoming() {
	runtime.Gosched() //start by giving the other go routines a chance to run
	for {
		if atomic.LoadUint64(&processed) != atomic.LoadUint64(&logged) {
			time.Sleep(2 * time.Millisecond)
		} else {
			return
		}
	}
}

//Waits for the specific log to be processed
func WaitForProcessed(logNum uint64) {
	runtime.Gosched() //start by giving the other go routines a chance to run
	for {
		if logNum != atomic.LoadUint64(&processed) {
			time.Sleep(1 * time.Millisecond)
		} else {
			return
		}
	}
}

//CaptureLoggingErrors allows the logging user to provide a channel
//for capturing logging errors. Any error during the logging process, like an
//appender failing will be sent to this channel.
//By default there is no error channel.
//Logging will not block when writting to the error channel so make sure the
//channel is big enough to capture errors
func CaptureLoggingErrors(errs chan<- error) {
	logMutex.Lock()
	logErrors = errs
	logMutex.Unlock()
}

//DefaultLogger returns a logger that can be used when a named logger isn't required
func DefaultLogger() Logger {
	return defaultLogger
}

//GetLogger returns a named logger, creating it if necessary. The logger will have the default settings.
//By default the logger will use the DefaultLoggers level and tag levels
func GetLogger(name string) Logger {
	logMutex.RLock()
	logger := loggers[name]
	logMutex.RUnlock()

	if logger == nil {
		logger = new(LoggerImpl)
		logger.name = name
		logger.level = DEFAULT
		logger.SetBufferLength(defaultLogger.buffer.Len())
		logMutex.Lock()
		loggers[name] = logger
		logMutex.Unlock()
	}

	return logger
}

//EnableVerboseLogging by default verbose logging is ignored, use this
//method to allow verbose logging
func EnableVerboseLogging() {
	atomic.StoreInt32(&enableVerbose, 1)
}

//DisableVerboseLogging by default verbose logging is ignored, use this
//method to turn off verbose logging if you have enabled it
func DisableVerboseLogging() {
	atomic.StoreInt32(&enableVerbose, 0)
}

//SetDefaultLogLevel sets the default loggers log level, flushes all buffers in case messages are cleared for logging
func SetDefaultLogLevel(l LogLevel) {
	defaultLogger.SetLogLevel(l)
}

//SetDefaultTagLogLevel sets the default loggers level for the specified tag, flushes all buffers in case messages are cleared for logging..
func SetDefaultTagLogLevel(tag string, l LogLevel) {
	defaultLogger.SetTagLevel(tag, l)
}

//SetDefaultFormatter sets the default formatter used by appenders that don't have their own
func SetDefaultFormatter(formatter LogFormatter) {
	logMutex.Lock()
	defaultFormatter = formatter
	logMutex.Unlock()
}

//SetDefaultBufferLength sets the buffer length for the default logger, new loggers will use this length.
//Existing loggers with buffers are not affected, those with buffers are not effected.
func SetDefaultBufferLength(length int) {

	logMutex.Lock()
	defaultLogger.setBufferLengthImpl(length)

	for _, val := range loggers {
		if val.buffer == nil {
			val.setBufferLengthImpl(length)
		}
	}

	logMutex.Unlock()
}

//AddAppender adds a new global appender for use by all loggers. Levels can be used to restrict logging to specific appenders.
func AddAppender(appender LogAppender) {
	logMutex.Lock()
	appenders = append(appenders, appender)
	logMutex.Unlock()
}

//ClearAppenders removes all of the global appenders, mainly used during configuration.
//Will pause and restart logging
func ClearAppenders() {
	PauseLogging()
	logMutex.Lock()
	for _, appender := range appenders {
		if app, ok := appender.(ClosableAppender); ok {
			app.Close()
		}
	}
	appenders = make([]LogAppender, 0)
	logMutex.Unlock()
	RestartLogging()
}

//ClearLoggers is provided so that an application can
//completely reset its logging configuration, for example
//on a SIGHUP
func ClearLoggers() {
	PauseLogging()
	logMutex.Lock()
	loggers = make(map[string]*LoggerImpl)
	logMutex.Unlock()
	RestartLogging()
}

/*
AddTag creates a new array and adds a string to it. This insures that no
slices are shared for tags.
*/
func AddTag(tags []string, newTag string) []string {
	newTags := make([]string, 0, len(tags)+1)
	newTags = append(newTags, tags...)
	newTags = append(newTags, newTag)
	return newTags
}

//SetLogLevel sets the level of messages allowed for a logger. This level can be
//overriden for specific tags using SetTagLevel. Changing the level for a Logger
//flushes its buffer in case messages are now free to be logged. This means that
//buffered messages might be printed out of order, but will be formatted to indicate this.
func (logger *LoggerImpl) SetLogLevel(l LogLevel) {
	logMutex.Lock()
	logger.level = l

	wait := new(sync.WaitGroup)

	if logger == defaultLogger {
		flushAllLoggers(wait)
	} else {
		wait.Add(1)
		logger.flushBuffer(wait)
	}
	logMutex.Unlock()
	wait.Wait()
}

//SetTagLevel assigns a log level to a specific tag. This level can override the general
//level for a logger allowing specific log messages to slip through and be appended to the logs
func (logger *LoggerImpl) SetTagLevel(tag string, l LogLevel) {
	logMutex.Lock()
	if logger.tagLevels == nil {
		logger.tagLevels = make(map[string]LogLevel)
	}
	logger.tagLevels[tag] = l
	wait := new(sync.WaitGroup)
	if logger == defaultLogger {
		flushAllLoggers(wait)
	} else {
		wait.Add(1)
		logger.flushBuffer(wait)
	}
	logMutex.Unlock()
	wait.Wait()
}

//SetBufferLength clears the buffer and creates a new one of the specified length.
func (logger *LoggerImpl) SetBufferLength(length int) {
	logMutex.Lock()

	logger.setBufferLengthImpl(length)

	logMutex.Unlock()
}

//expects the lock
func (logger *LoggerImpl) setBufferLengthImpl(length int) {

	if length == 0 {
		logger.buffer = nil
	} else if length != logger.buffer.Len() {
		logger.buffer = ring.New(length)
	}
}

//NewLogRecord creates a log record object
func NewLogRecord(logger *LoggerImpl, level LogLevel, tags []string, message string, time time.Time, original time.Time) *LogRecord {
	record := new(LogRecord)
	record.Logger = logger
	record.Level = level
	record.Tags = tags
	record.Message = message
	record.Time = time
	record.Original = original
	return record
}

//should be called inside the logging lock,
//puts the error on the logging error channel if one is set
func logError(err error) {
	if err != nil && logErrors != nil {

		select {
		case logErrors <- err:
			//write the error
		default:
			//don't write or block
		}
	}
}

/* Check the tags for this logger, or the defaults, if any pass, then we pass */
/* Should be called inside the logging lock */
func (logger *LoggerImpl) checkTagLevel(l LogLevel, tags []string) bool {

	for _, tag := range tags {

		if logger.tagLevels != nil {
			if tagLevel, ok := logger.tagLevels[tag]; ok && tagLevel <= l {
				return true
			}
		}

		if logger != defaultLogger && defaultLogger.tagLevels != nil {
			if tagLevel, ok := defaultLogger.tagLevels[tag]; ok && tagLevel <= l {
				return true
			}
		}
	}

	return false
}

//CheckLevel tests the default logger for its permissions
func CheckLevel(l LogLevel, tags []string) bool {
	return defaultLogger.CheckLevel(l, tags)
}

//CheckLevel checks tags, then check the level on this , or the default level
func (logger *LoggerImpl) CheckLevel(l LogLevel, tags []string) bool {

	logMutex.RLock()
	defer logMutex.RUnlock()

	return logger.checkLevelWithTags(l, tags)
}

//requires the lock be acquired
func (logger *LoggerImpl) checkLevelWithTags(l LogLevel, tags []string) bool {

	if (logger.tagLevels != nil || defaultLogger.tagLevels != nil) && tags != nil {
		matchTag := logger.checkTagLevel(l, tags)
		if matchTag {
			return true //otherwise check the general level
		}
	}

	if logger.level != DEFAULT {
		return logger.level <= l
	}

	return defaultLogger.level <= l
}

//flushAllLoggers expects the logging lock to be held by the caller
func flushAllLoggers(wait *sync.WaitGroup) {
	wait.Add(len(loggers) + 1)
	for _, val := range loggers {
		val.flushBuffer(wait)
	}
	defaultLogger.flushBuffer(wait)
}

//should be called witin the lock
func logToAppenders(record *LogRecord) {
	for _, appender := range appenders {
		err := appender.Log(record)
		logError(err)
	}
}

func processLogRecord(record *LogRecord) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	logger := record.Logger
	passed := logger.checkLevelWithTags(record.Level, record.Tags)

	if passed {
		logToAppenders(record)
	} else if logger.buffer != nil && record.Level > VERBOSE {
		logger.buffer.Next().Value = record
		logger.buffer = logger.buffer.Next()
	}
	atomic.AddUint64(&processed, 1)
}

//flushBuffer expects the logging lock to be held, and does not take the lock
//should call done on the wait group when the buffer is flushed
//does not 1 to the waitgroup
func (logger *LoggerImpl) flushBuffer(wait *sync.WaitGroup) {
	if logger.buffer != nil {
		now := time.Now()
		oldBuffer := logger.buffer
		logger.buffer = ring.New(oldBuffer.Len())

		go func() {
			oldBuffer.Do(func(x interface{}) {

				if x == nil {
					return
				}

				record := x.(*LogRecord)
				record.Time = now

				atomic.AddUint64(&logged, 1)
				incomingChannel <- record
			})

			wait.Done()
		}()
	} else {
		wait.Done()
	}
}

func (logger *LoggerImpl) logwithformat(level LogLevel, tags []string, format string, args ...interface{}) uint64 {

	if level == VERBOSE && atomic.LoadInt32(&enableVerbose) != 1 {
		return 0
	}

	now := time.Now()
	msg := ""

	if format == "" {
		msg = fmt.Sprint(args...)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	if level == PANIC {
		stack := make([]byte, 10 * 1024)
		size := runtime.Stack(stack, false)
		stackStr := strings.Replace(string(stack[:size]), "\n", "\n  ", -1)
		msg = msg + "\n  " + stackStr
	}

	logRecord := NewLogRecord(logger, level, tags, msg, now, now)
	logNum := atomic.AddUint64(&logged, 1)
	incomingChannel <- logRecord

	//return the logged number to track if it was processed
	return logNum
}

func (logger *LoggerImpl) log(level LogLevel, tags []string, args ...interface{}) uint64{
	return logger.logwithformat(level, tags, "", args...)
}

//PanicWithTagsf logs a PANIC level message with the provided tags and formatted string.
func (logger *LoggerImpl) PanicWithTagsf(tags []string, format string, args ...interface{}) {
	logNum := logger.logwithformat(PANIC, tags, format, args...)

	//ensure the panic message we just logged gets processed
	WaitForProcessed(logNum)

	//flush all logs before panicking
	logMutex.Lock()
	wait := new(sync.WaitGroup)
	if logger == defaultLogger {
		flushAllLoggers(wait)
	} else {
		wait.Add(1)
		logger.flushBuffer(wait)
	}
	logMutex.Unlock()
	wait.Wait()

	//continue with the panic
	if format == "" {
		panic(fmt.Sprint(args...))
	} else {
		panic(fmt.Sprintf(format, args...))
	}
}

//PanicWithTags logs a PANIC level message with the provided tags and provided arguments joined into a string.
func (logger *LoggerImpl) PanicWithTags(tags []string, args ...interface{}) {
	logNum := logger.log(PANIC, tags, args...)

	//ensure the panic message we just logged gets processed
	WaitForProcessed(logNum)

	//flush all logs before panicking
	logMutex.Lock()
	wait := new(sync.WaitGroup)
	if logger == defaultLogger {
		flushAllLoggers(wait)
	} else {
		wait.Add(1)
		logger.flushBuffer(wait)
	}
	logMutex.Unlock()
	wait.Wait()

	//continue with panic
	panic(fmt.Sprint(args...))
}

//Panicf logs a PANIC level message with the no tags and formatted string.
func (logger *LoggerImpl) Panicf(format string, args ...interface{}) {
	logNum := logger.logwithformat(PANIC, nil, format, args...)

	//ensure the panic message we just logged gets processed
	WaitForProcessed(logNum)

	//flush all logs before panicking
	logMutex.Lock()
	wait := new(sync.WaitGroup)
	if logger == defaultLogger {
		flushAllLoggers(wait)
	} else {
		wait.Add(1)
		logger.flushBuffer(wait)
	}
	logMutex.Unlock()
	wait.Wait()

	//continue with the panic
	if format == "" {
		panic(fmt.Sprint(args...))
	} else {
		panic(fmt.Sprintf(format, args...))
	}
}

//Panic logs a PANIC level message with no tags and provided arguments joined into a string.
func (logger *LoggerImpl) Panic(args ...interface{}) {
	logNum := logger.log(PANIC, nil, args...)

	//ensure the panic message we just logged gets processed
	WaitForProcessed(logNum)

	//flush all logs before panicking
	logMutex.Lock()
	wait := new(sync.WaitGroup)
	if logger == defaultLogger {
		flushAllLoggers(wait)
	} else {
		wait.Add(1)
		logger.flushBuffer(wait)
	}
	logMutex.Unlock()
	wait.Wait()

	//continue with the panic
	panic(fmt.Sprint(args...))
}

//ErrorWithTagsf logs an ERROR level message with the provided tags and formatted string.
func (logger *LoggerImpl) ErrorWithTagsf(tags []string, fmt string, args ...interface{}) {
	logger.logwithformat(ERROR, tags, fmt, args...)
}

//ErrorWithTags logs an ERROR level message with the provided tags and provided arguments joined into a string.
func (logger *LoggerImpl) ErrorWithTags(tags []string, args ...interface{}) {
	logger.log(ERROR, tags, args...)
}

//Errorf logs an ERROR level message with the no tags and formatted string.
func (logger *LoggerImpl) Errorf(fmt string, args ...interface{}) {
	logger.logwithformat(ERROR, nil, fmt, args...)
}

//Error logs an ERROR level message with no tags and provided arguments joined into a string.
func (logger *LoggerImpl) Error(args ...interface{}) {
	logger.log(ERROR, nil, args...)
}

//WarnWithTagsf logs an WARN level message with the provided tags and formatted string.
func (logger *LoggerImpl) WarnWithTagsf(tags []string, fmt string, args ...interface{}) {
	logger.logwithformat(WARN, tags, fmt, args...)
}

//WarnWithTags logs an WARN level message with the provided tags and provided arguments joined into a string.
func (logger *LoggerImpl) WarnWithTags(tags []string, args ...interface{}) {
	logger.log(WARN, tags, args...)
}

//Warnf logs an WARN level message with the no tags and formatted string.
func (logger *LoggerImpl) Warnf(fmt string, args ...interface{}) {
	logger.logwithformat(WARN, nil, fmt, args...)
}

//Warn logs an WARN level message with no tags and provided arguments joined into a string.
func (logger *LoggerImpl) Warn(args ...interface{}) {
	logger.log(WARN, nil, args...)
}

//InfoWithTagsf logs an INFO level message with the provided tags and formatted string.
func (logger *LoggerImpl) InfoWithTagsf(tags []string, fmt string, args ...interface{}) {
	logger.logwithformat(INFO, tags, fmt, args...)
}

//InfoWithTags logs an INFO level message with the provided tags and provided arguments joined into a string.
func (logger *LoggerImpl) InfoWithTags(tags []string, args ...interface{}) {
	logger.log(INFO, tags, args...)
}

//Infof logs an INFO level message with the no tags and formatted string.
func (logger *LoggerImpl) Infof(fmt string, args ...interface{}) {
	logger.logwithformat(INFO, nil, fmt, args...)
}

//Info logs an INFO level message with no tags and provided arguments joined into a string.
func (logger *LoggerImpl) Info(args ...interface{}) {
	logger.log(INFO, nil, args...)
}

//DebugWithTagsf logs an DEBUG level message with the provided tags and formatted string.
func (logger *LoggerImpl) DebugWithTagsf(tags []string, fmt string, args ...interface{}) {
	logger.logwithformat(DEBUG, tags, fmt, args...)
}

//DebugWithTags logs an DEBUG level message with the provided tags and provided arguments joined into a string.
func (logger *LoggerImpl) DebugWithTags(tags []string, args ...interface{}) {
	logger.log(DEBUG, tags, args...)
}

//Debugf logs an DEBUG level message with the no tags and formatted string.
func (logger *LoggerImpl) Debugf(fmt string, args ...interface{}) {
	logger.logwithformat(DEBUG, nil, fmt, args...)
}

//Debug logs an DEBUG level message with no tags and provided arguments joined into a string.
func (logger *LoggerImpl) Debug(args ...interface{}) {
	logger.log(DEBUG, nil, args...)
}

//VerboseWithTagsf logs an VERBOSE level message with the provided tags and formatted string.
//Verbose messages are not buffered
func (logger *LoggerImpl) VerboseWithTagsf(tags []string, fmt string, args ...interface{}) {
	logger.logwithformat(VERBOSE, tags, fmt, args...)
}

//Verbosef logs an VERBOSE level message with the no tags and formatted string.
//Verbose messages are not buffered
func (logger *LoggerImpl) Verbosef(fmt string, args ...interface{}) {
	logger.logwithformat(VERBOSE, nil, fmt, args...)
}

//ErrorWithTagsf logs an ERROR level message with the provided tags and formatted string. Uses the default logger.
func ErrorWithTagsf(tags []string, fmt string, args ...interface{}) {
	defaultLogger.logwithformat(ERROR, tags, fmt, args...)
}

//ErrorWithTags logs an ERROR level message with the provided tags and provided arguments joined into a string. Uses the default logger.
func ErrorWithTags(tags []string, args ...interface{}) {
	defaultLogger.log(ERROR, tags, args...)
}

//Errorf logs an ERROR level message with the no tags and formatted string. Uses the default logger.
func Errorf(fmt string, args ...interface{}) {
	defaultLogger.logwithformat(ERROR, nil, fmt, args...)
}

//Error logs an ERROR level message with no tags and provided arguments joined into a string. Uses the default logger.
func Error(args ...interface{}) {
	defaultLogger.log(ERROR, nil, args...)
}

//WarnWithTagsf logs an WARN level message with the provided tags and formatted string. Uses the default logger.
func WarnWithTagsf(tags []string, fmt string, args ...interface{}) {
	defaultLogger.logwithformat(WARN, tags, fmt, args...)
}

//WarnWithTags logs an WARN level message with the provided tags and provided arguments joined into a string. Uses the default logger.
func WarnWithTags(tags []string, args ...interface{}) {
	defaultLogger.log(WARN, tags, args...)
}

//Warnf logs an WARN level message with the no tags and formatted string. Uses the default logger.
func Warnf(fmt string, args ...interface{}) {
	defaultLogger.logwithformat(WARN, nil, fmt, args...)
}

//Warn logs an WARN level message with no tags and provided arguments joined into a string. Uses the default logger.
func Warn(args ...interface{}) {
	defaultLogger.log(WARN, nil, args...)
}

//InfoWithTagsf logs an INFO level message with the provided tags and formatted string. Uses the default logger.
func InfoWithTagsf(tags []string, fmt string, args ...interface{}) {
	defaultLogger.logwithformat(INFO, tags, fmt, args...)
}

//InfoWithTags logs an INFO level message with the provided tags and provided arguments joined into a string. Uses the default logger.
func InfoWithTags(tags []string, args ...interface{}) {
	defaultLogger.log(INFO, tags, args...)
}

//Infof logs an INFO level message with the no tags and formatted string. Uses the default logger.
func Infof(fmt string, args ...interface{}) {
	defaultLogger.logwithformat(INFO, nil, fmt, args...)
}

//Info logs an INFO level message with no tags and provided arguments joined into a string. Uses the default logger.
func Info(args ...interface{}) {
	defaultLogger.log(INFO, nil, args...)
}

//DebugWithTagsf logs an DEBUG level message with the provided tags and formatted string. Uses the default logger.
func DebugWithTagsf(tags []string, fmt string, args ...interface{}) {
	defaultLogger.logwithformat(DEBUG, tags, fmt, args...)
}

//DebugWithTags logs an DEBUG level message with the provided tags and provided arguments joined into a string. Uses the default logger.
func DebugWithTags(tags []string, args ...interface{}) {
	defaultLogger.log(DEBUG, tags, args...)
}

//Debugf logs an DEBUG level message with the no tags and formatted string. Uses the default logger.
func Debugf(fmt string, args ...interface{}) {
	defaultLogger.logwithformat(DEBUG, nil, fmt, args...)
}

//Debug logs an DEBUG level message with no tags and provided arguments joined into a string. Uses the default logger.
func Debug(args ...interface{}) {
	defaultLogger.log(DEBUG, nil, args...)
}

//VerboseWithTagsf logs an VERBOSE level message with the provided tags and formatted string. Uses the default logger.
//Verbose messages are not buffered
func VerboseWithTagsf(tags []string, fmt string, args ...interface{}) {
	defaultLogger.logwithformat(VERBOSE, tags, fmt, args...)
}

//Verbosef logs an VERBOSE level message with the no tags and formatted string. Uses the default logger.
//Verbose messages are not buffered
func Verbosef(fmt string, args ...interface{}) {
	defaultLogger.logwithformat(VERBOSE, nil, fmt, args...)
}
