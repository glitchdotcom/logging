package logging

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestRollingAppender(t *testing.T) {

	filepath := path.Join(os.TempDir(), "appendtest")
	app := NewRollingFileAppender(filepath, "log", int64(2048), 5)
	app.SetFormatter(GetFormatter(MINIMAL))

	memoryAppender := NewMemoryAppender()
	memoryAppender.SetFormatter(GetFormatter(MINIMAL))

	ClearAppenders()
	AddAppender(app)
	AddAppender(memoryAppender)

	SetDefaultLogLevel(INFO)

	for i := 0; i < 2548; i++ {
		Warn("1")
		Debug("1") //none of these should get logged
	}

	WaitForIncoming()
	ClearAppenders() //will close the rolling log appender

	assert.Equal(t, len(memoryAppender.GetLoggedMessages()), 2548, "should have logged all the messages")

	pathOne := fmt.Sprintf("%s.log", filepath)
	info, err := os.Stat(pathOne)
	assert.Nil(t, err, "Stat should be able to find the log file")
	assert.Equal(t, info.Size(), 1000, "new file should have 1000 bytes, 500 1's and new lines")

	pathTwo := fmt.Sprintf("%s.1.log", filepath)
	info, err = os.Stat(pathTwo)
	assert.Nil(t, err, "Stat should be able to find the rolled log file")
	assert.Equal(t, info.Size(), 2048, "rolled file should have 2048 bytes")
}

func TestRollingAppenderOneFile(t *testing.T) {

	filepath := path.Join(os.TempDir(), "appendtest")
	app := NewRollingFileAppender(filepath, "log", int64(2048), 1)
	app.SetFormatter(GetFormatter(MINIMAL))

	memoryAppender := NewMemoryAppender()
	memoryAppender.SetFormatter(GetFormatter(MINIMAL))

	pathOne := fmt.Sprintf("%s.log", filepath)
	err := os.Remove(pathOne) //we won't roll so make sure we start with nothing

	if err != nil && !os.IsNotExist(err) {
		assert.Nil(t, err, "Should be able to delete")
	}

	ClearAppenders()
	AddAppender(app)
	AddAppender(memoryAppender)

	SetDefaultLogLevel(INFO)

	for i := 0; i < 2548; i++ {
		Warn("1")
		Debug("1") //none of these should get logged
	}

	WaitForIncoming()
	ClearAppenders() //will close the rolling log appender

	assert.Equal(t, len(memoryAppender.GetLoggedMessages()), 2548, "should have logged all the messages")

	pathOne = fmt.Sprintf("%s.log", filepath)
	info, err := os.Stat(pathOne)
	assert.Nil(t, err, "Stat should be able to find the log file")
	assert.Equal(t, info.Size(), 2548*2, "new file should have all the data, since there isn't any rolling")
}

func TestRollingAppenderNew(t *testing.T) {

	filepath := path.Join(os.TempDir(), "appendtest")
	app := NewRollingFileAppender(filepath, "log", int64(100), -1)

	assert.Equal(t, app.maxFiles, 1, "max files defaults to 1")
	assert.Equal(t, app.maxFileSize, 1024, "max filesize defaults to 1024")
	assert.Equal(t, app.currentFileName(), fmt.Sprintf("%s.%s", filepath, "log"), "current file name is always prefix.suffix")
}
