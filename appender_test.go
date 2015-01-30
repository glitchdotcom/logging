package logging

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path"
	"testing"
)

func TestAppenderLevel(t *testing.T) {

	logger, memory := setup()
	logger.SetLogLevel(DEBUG)
	memory.SetLevel(WARN)

	secondAppender := NewMemoryAppender()
	secondAppender.SetLevel(DEBUG)
	AddAppender(secondAppender)

	logger.Error("error")
	logger.Info("info")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 1, "Appender should filter messages.")
	assert.Equal(t, len(secondAppender.GetLoggedMessages()), 2, "Appender should work separately.")
}

func TestNullAppender(t *testing.T) {
	ClearAppenders()

	app := NewNullAppender()
	AddAppender(app)

	SetDefaultLogLevel(INFO)
	Info("one")
	Debug("two")

	WaitForIncoming()
	assert.Equal(t, app.Count(), 1, "Null appender should check levels appropriately")
}

func TestAppenderCheckLevel(t *testing.T) { //not sure how to test std err without subproc so this is for coverage
	ClearAppenders()

	app := NewStdErrAppender()
	AddAppender(app)

	app.SetLevel(INFO)

	assert.True(t, app.CheckLevel(ERROR), "error is allowed")
	assert.True(t, app.CheckLevel(INFO), "info is allowed")
	assert.False(t, app.CheckLevel(DEBUG), "debug is not allowed")
}

func TestStdErrAppender(t *testing.T) { //not sure how to test std err without subproc so this is for coverage
	ClearAppenders()

	app := NewStdErrAppender()
	AddAppender(app)

	SetDefaultLogLevel(INFO)
	Info("one")
	Debug("two")
}

func TestStdOutAppender(t *testing.T) { //not sure how to test std out without subproc so this is for coverage
	ClearAppenders()

	app := NewStdOutAppender()
	AddAppender(app)

	SetDefaultLogLevel(INFO)
	Info("one")
	Debug("two")
}

func TestWriterAppender(t *testing.T) {
	ClearAppenders()

	SetDefaultLogLevel(DEBUG)

	filepath := path.Join(os.TempDir(), "writerlogtest.txt")
	f, _ := os.Create(filepath)
	app := NewWriterAppender(f)
	app.SetFormatter(GetFormatter(MINIMAL))
	AddAppender(app)

	app.SetLevel(INFO)

	Info("one")
	Debug("two")

	WaitForIncoming()
	PauseLogging() // data race if we don't pause
	f.Close()

	buf := bytes.NewBuffer(nil)
	f, _ = os.Open(filepath)
	io.Copy(buf, f)
	f.Close()

	s := string(buf.Bytes())

	assert.Equal(t, s, "one\n", "File should contain a single entry for writer appender")
	RestartLogging() //don't leave logging off

}
