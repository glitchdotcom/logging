package logging

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var count = 1

func setup() (Logger, *MemoryAppender) {

	SetDefaultLogLevel(INFO)
	defaultLogger.SetBufferLength(0)
	memoryAppender := NewMemoryAppender()
	memoryAppender.SetFormatter(GetFormatter(MINIMAL))
	logger := GetLogger(fmt.Sprintf("testLogger-%d", count))
	count++
	WaitForIncoming()
	ClearAppenders()
	AddAppender(memoryAppender)
	return logger, memoryAppender
}

func TestNamedLoggers(t *testing.T) {
	logger := GetLogger("named-logger")
	logger2 := GetLogger("named-logger")

	assert.True(t, logger == logger2, "named loggers should be the same")

	ClearLoggers()

	logger2 = GetLogger("named-logger")
	assert.False(t, logger == logger2, "named loggers should change when cleared")
}

func TestAddTag(t *testing.T) {
	t.Parallel()

	tags := []string{"one"}
	assert.Equal(t, AddTag(tags, "two"), []string{"one", "two"}, "Add tag should append to the array")
}

func TestClearAppenders(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(DEBUG)

	logger.Error("error")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Removed Appender should receive old messages.")

	secondAppender := NewMemoryAppender()
	ClearAppenders()
	AddAppender(secondAppender)

	logger.Error("error")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Removed Appender shouldn't receive new messages.")
	assert.Equal(t, len(secondAppender.GetLoggedMessages()), 1, "New Appender should only receive new messages.")
}

func TestLevelFilteringError(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(ERROR)

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Only messages at level ERROR should be logged.")
}

func TestLevelFilteringDebug(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(DEBUG)

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")
	logger.Verbosef("verbose")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 4, "All messages at level DEBUG should be logged.")
}

func TestLevelFilteringVerbose(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(VERBOSE)
	SetDefaultLogLevel(VERBOSE)

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")
	logger.Verbosef("verbose")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 4, "All messages at level DEBUG should be logged, verbose isn't on.")

	EnableVerboseLogging()
	logger.Verbosef("verbose")
	logger.VerboseWithTagsf([]string{"tag"}, "verbose")
	Verbosef("verbose")
	VerboseWithTagsf([]string{"tag"}, "verbose")
	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 8, "verbose is on.")

	DisableVerboseLogging()
	logger.Verbosef("verbose")
	logger.VerboseWithTagsf([]string{"tag"}, "verbose")
	Verbosef("verbose")
	VerboseWithTagsf([]string{"tag"}, "verbose")
	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 8, "verbose is off.")
}

func TestTagLevelFilteringDebug(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(ERROR)
	logger.SetTagLevel("tag", DEBUG)

	tags := []string{"tag"}

	logger.ErrorWithTags(tags, "error")
	logger.WarnWithTags(tags, "warn")
	logger.InfoWithTags(tags, "info")
	logger.DebugWithTags(tags, "debug")

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 5, "All messages with the tag should be logged.")
}

func TestDefaultTagLevelFilteringDebug(t *testing.T) {

	_, memory := setup()
	SetDefaultLogLevel(ERROR)
	SetDefaultTagLogLevel("tag", DEBUG)

	tags := []string{"tag"}

	ErrorWithTags(tags, "error")
	WarnWithTags(tags, "warn")
	InfoWithTags(tags, "info")
	DebugWithTags(tags, "debug")

	Error("error")
	Warn("warn")
	Info("info")
	Debug("debug")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 5, "All messages with the tag should be logged.")
}

func TestBuffer(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(ERROR)
	logger.SetBufferLength(10)

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Only messages at level ERROR should be logged.")

	logger.SetLogLevel(WARN)

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 2, "Buffered messages should be logged after level change.")

	logger.SetLogLevel(DEBUG)

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 4, "Buffered messages should have been saved, then logged after level change.")
}

func TestBufferLength(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(ERROR)
	logger.SetBufferLength(2)

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Only messages at level ERROR should be logged.")

	logger.SetLogLevel(DEBUG)

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 3, "Buffer should have only had 2 messages.")
}

func TestFormatMethods(t *testing.T) {
	logger, memory := setup()
	logger.SetLogLevel(DEBUG)

	logger.Error("one")
	logger.Errorf("%v %v", "one", 1)

	logger.Warn("one")
	logger.Warnf("%v %v", "one", 1)

	logger.Info("one")
	logger.Infof("%v %v", "one", 1)

	logger.Debug("one")
	logger.Debugf("%v %v", "one", 1)

	WaitForIncoming()
	assert.Equal(t, memory.GetLoggedMessages()[0], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[1], "one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[2], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[3], "one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[4], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[5], "one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[6], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[7], "one 1", "messages should be formatted")
}

func TestFormatMethodsWithTags(t *testing.T) {
	logger, memory := setup()
	memory.SetFormatter(GetFormatter(MINIMALTAGGED))
	logger.SetLogLevel(DEBUG)

	tags := []string{"blit", "blat"}

	logger.ErrorWithTags(tags, "one")
	logger.ErrorWithTagsf(tags, "%v %v", "one", 1)

	logger.WarnWithTags(tags, "one")
	logger.WarnWithTagsf(tags, "%v %v", "one", 1)

	logger.InfoWithTags(tags, "one")
	logger.InfoWithTagsf(tags, "%v %v", "one", 1)

	logger.DebugWithTags(tags, "one")
	logger.DebugWithTagsf(tags, "%v %v", "one", 1)

	WaitForIncoming()
	assert.Equal(t, memory.GetLoggedMessages()[0], "[ERROR] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[1], "[ERROR] [blit blat] one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[2], "[WARN] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[3], "[WARN] [blit blat] one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[4], "[INFO] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[5], "[INFO] [blit blat] one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[6], "[DEBUG] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[7], "[DEBUG] [blit blat] one 1", "messages should be formatted")
}

func TestPanicLogging(t *testing.T) {
	logger, memory := setup()
	memory.SetFormatter(GetFormatter(MINIMALTAGGED))
	logger.SetLogLevel(DEBUG)

	tags := []string{"blit", "blat"}

	defer func() {
		rec := recover()
		assert.NotNil(t, rec, "panic message")
		assert.Equal(t, "panic", rec, "panic message")

		assert.Equal(t, memory.GetLoggedMessages()[0], "[ERROR] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[1], "[ERROR] [blit blat] one 1", "messages should be formatted")
		assert.Equal(t, memory.GetLoggedMessages()[2], "[WARN] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[3], "[WARN] [blit blat] one 1", "messages should be formatted")
		assert.Equal(t, memory.GetLoggedMessages()[4], "[INFO] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[5], "[INFO] [blit blat] one 1", "messages should be formatted")
		assert.Equal(t, memory.GetLoggedMessages()[6], "[DEBUG] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[7], "[DEBUG] [blit blat] one 1", "messages should be formatted")

		msgWithStack := strings.Split(memory.GetLoggedMessages()[8], "\n")

		assert.Equal(t, msgWithStack[0], "[PANIC] [blit blat] panic", "unformatted messages should be unchanged")
		assert.True(t, len(msgWithStack) > 1, "panic messages should include stack trace")
		for i := 1; i < len(msgWithStack); i++ {
			assert.Equal(t, "  ", msgWithStack[i][0:2], "every stack trace line should be indented two spaces")
		}

	}()

	logger.ErrorWithTags(tags, "one")
	logger.ErrorWithTagsf(tags, "%v %v", "one", 1)

	logger.WarnWithTags(tags, "one")
	logger.WarnWithTagsf(tags, "%v %v", "one", 1)

	logger.InfoWithTags(tags, "one")
	logger.InfoWithTagsf(tags, "%v %v", "one", 1)

	logger.DebugWithTags(tags, "one")
	logger.DebugWithTagsf(tags, "%v %v", "one", 1)

	logger.PanicWithTags(tags, "panic")
}

func TestPanicFormatLogging(t *testing.T) {
	logger, memory := setup()
	memory.SetFormatter(GetFormatter(MINIMALTAGGED))
	logger.SetLogLevel(DEBUG)

	tags := []string{"blit", "blat"}

	defer func() {

		rec := recover()
		assert.NotNil(t, rec, "panic message")
		assert.Equal(t, "panic 1", rec, "panic message")

		assert.Equal(t, memory.GetLoggedMessages()[0], "[ERROR] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[1], "[ERROR] [blit blat] one 1", "messages should be formatted")
		assert.Equal(t, memory.GetLoggedMessages()[2], "[WARN] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[3], "[WARN] [blit blat] one 1", "messages should be formatted")
		assert.Equal(t, memory.GetLoggedMessages()[4], "[INFO] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[5], "[INFO] [blit blat] one 1", "messages should be formatted")
		assert.Equal(t, memory.GetLoggedMessages()[6], "[DEBUG] [blit blat] one", "unformatted messages should be unchanged")
		assert.Equal(t, memory.GetLoggedMessages()[7], "[DEBUG] [blit blat] one 1", "messages should be formatted")

		msgWithStack := strings.Split(memory.GetLoggedMessages()[8], "\n")

		assert.Equal(t, msgWithStack[0], "[PANIC] [blit blat] panic 1", "messages should be formatted")
		assert.True(t, len(msgWithStack) > 1, "panic messages should include stack trace")
		for i := 1; i < len(msgWithStack); i++ {
			assert.Equal(t, "  ", msgWithStack[i][0:2], "every stack trace line should be indented two spaces")
		}
	}()

	logger.ErrorWithTags(tags, "one")
	logger.ErrorWithTagsf(tags, "%v %v", "one", 1)

	logger.WarnWithTags(tags, "one")
	logger.WarnWithTagsf(tags, "%v %v", "one", 1)

	logger.InfoWithTags(tags, "one")
	logger.InfoWithTagsf(tags, "%v %v", "one", 1)

	logger.DebugWithTags(tags, "one")
	logger.DebugWithTagsf(tags, "%v %v", "one", 1)

	logger.PanicWithTagsf(tags, "%v %v", "panic", 1)
}

func TestFormatMethodsWithDefaultLogger(t *testing.T) {
	_, memory := setup()
	SetDefaultLogLevel(DEBUG)
	SetDefaultFormatter(GetFormatter(MINIMAL))

	Error("one")
	Errorf("%v %v", "one", 1)

	Warn("one")
	Warnf("%v %v", "one", 1)

	Info("one")
	Infof("%v %v", "one", 1)

	Debug("one")
	Debugf("%v %v", "one", 1)

	WaitForIncoming()
	assert.Equal(t, memory.GetLoggedMessages()[0], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[1], "one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[2], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[3], "one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[4], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[5], "one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[6], "one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[7], "one 1", "messages should be formatted")
}

func TestFormatMethodsWithTagsAndDefaultLogger(t *testing.T) {
	_, memory := setup()
	SetDefaultFormatter(GetFormatter(MINIMALTAGGED))
	SetDefaultLogLevel(DEBUG)

	memory = NewMemoryAppender() //will get the default formatter
	ClearAppenders()
	AddAppender(memory)

	tags := []string{"blit", "blat"}

	ErrorWithTags(tags, "one")
	ErrorWithTagsf(tags, "%v %v", "one", 1)

	WarnWithTags(tags, "one")
	WarnWithTagsf(tags, "%v %v", "one", 1)

	InfoWithTags(tags, "one")
	InfoWithTagsf(tags, "%v %v", "one", 1)

	DebugWithTags(tags, "one")
	DebugWithTagsf(tags, "%v %v", "one", 1)

	WaitForIncoming()
	assert.Equal(t, memory.GetLoggedMessages()[0], "[ERROR] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[1], "[ERROR] [blit blat] one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[2], "[WARN] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[3], "[WARN] [blit blat] one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[4], "[INFO] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[5], "[INFO] [blit blat] one 1", "messages should be formatted")
	assert.Equal(t, memory.GetLoggedMessages()[6], "[DEBUG] [blit blat] one", "unformatted messages should be unchanged")
	assert.Equal(t, memory.GetLoggedMessages()[7], "[DEBUG] [blit blat] one 1", "messages should be formatted")
}

func TestDefaultLogger(t *testing.T) {

	memory := NewMemoryAppender()
	ClearAppenders()
	AddAppender(memory)

	SetDefaultLogLevel(ERROR)
	SetDefaultBufferLength(10)

	Error("error")
	Warn("warn")
	Info("info")
	Debug("debug")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Only messages at level ERROR should be logged.")

	SetDefaultLogLevel(WARN)

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 2, "Buffered messages should be logged after level change.")

	SetDefaultLogLevel(DEBUG)

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 4, "Buffered messages should have been saved, then logged after level change.")
}

func TestWaitForIncoming(t *testing.T) {
	count := 1000
	memory := NewMemoryAppender()
	ClearAppenders()
	AddAppender(memory)

	SetDefaultLogLevel(DEBUG)

	for i := 0; i < count; i++ {
		Error("error")
	}

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), count, "All messages should be logged.")
}

func TestConcurrentLogging(t *testing.T) {
	runtime.GOMAXPROCS(8)
	count := 1000
	concur := 50
	memory := NewNullAppender()
	ClearAppenders()
	AddAppender(memory)

	SetDefaultLogLevel(DEBUG)

	waiter := sync.WaitGroup{}
	waiter.Add(concur)

	for gos := 0; gos < concur; gos++ {
		go func() {
			pauseAt := int(rand.Int31n(int32(count)))
			for i := 0; i < count; i++ {
				if pauseAt == i {
					PauseLogging()
					time.Sleep(1 * time.Millisecond)
					RestartLogging()
				}
				Error("error")
				time.Sleep(1 * time.Microsecond)
			}
			waiter.Done()
		}()
	}

	waiter.Wait()
	WaitForIncoming()
	assert.Equal(t, memory.Count(), int64(count*concur), "All messages should be logged.")
}

func TestStopStartLogging(t *testing.T) {
	memory := NewNullAppender()
	ClearAppenders()
	AddAppender(memory)

	SetDefaultLogLevel(DEBUG)

	Error("error")

	WaitForIncoming()
	assert.Equal(t, memory.Count(), int64(1), "Only messages at level ERROR should be logged.")

	PauseLogging()

	Error("error")
	assert.Equal(t, memory.Count(), int64(1), "Only messages at level ERROR should be logged.")

	RestartLogging()

	Error("error")
	WaitForIncoming()
	assert.Equal(t, memory.Count(), int64(3), "Only messages at level ERROR should be logged.")
}

func TestConfigWhileStoppedLogging(t *testing.T) {
	memory := NewMemoryAppender()
	ClearAppenders()
	AddAppender(memory)

	SetDefaultLogLevel(DEBUG)

	Error("error")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Only messages at level ERROR should be logged.")

	PauseLogging()

	memory2 := NewMemoryAppender()
	ClearAppenders()
	AddAppender(memory2)

	RestartLogging()

	Error("error")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Only old messages should be in the old log.")
	assert.Equal(t, len(memory2.GetLoggedMessages()), 1, "Only new messages should be in the new log.")
}

func TestErrorChannel(t *testing.T) {

	errors := make(chan error, 10)
	logger, _ := setup()
	logger.SetLogLevel(DEBUG)

	errorApp := NewErrorAppender()
	ClearAppenders()
	AddAppender(errorApp)

	CaptureLoggingErrors(errors)

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")

	err := <-errors
	assert.Equal(t, err.Error(), "error: error", "errors should be pushed to the channel in order.")
	err = <-errors
	assert.Equal(t, err.Error(), "error: warn", "errors should be pushed to the channel in order.")
	err = <-errors
	assert.Equal(t, err.Error(), "error: info", "errors should be pushed to the channel in order.")
	err = <-errors
	assert.Equal(t, err.Error(), "error: debug", "errors should be pushed to the channel in order.")

	WaitForIncoming()
	assert.Equal(t, errorApp.Count(), int64(4), "All messages should be logged.")
}

func TestErrorChannelWontBlock(t *testing.T) {

	errors := make(chan error)
	logger, _ := setup()
	logger.SetLogLevel(DEBUG)

	errorApp := NewErrorAppender()
	ClearAppenders()
	AddAppender(errorApp)

	CaptureLoggingErrors(errors)

	logger.Error("error")
	logger.Warn("warn")
	logger.Info("info")
	logger.Debug("debug")

	select {
	case err := <-errors:
		assert.Nil(t, err, "Errors should be empty since we don't block")
	default:
		//ok
	}

	WaitForIncoming()
	assert.Equal(t, errorApp.Count(), int64(4), "All messages should be logged.")
}
