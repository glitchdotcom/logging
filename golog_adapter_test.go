package logging

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestGoLogAdapter(t *testing.T) {

	memory := NewMemoryAppender()
	ClearAppenders()
	AddAppender(memory)

	AdaptStandardLogging(ERROR, []string{})

	SetDefaultLogLevel(WARN)
	SetDefaultBufferLength(10)

	log.Print("error")
	log.Print("warn")
	log.Print("info")
	log.Print("debug")

	WaitForIncoming()
	assert.Equal(t, len(memory.GetLoggedMessages()), 4, "All messages at error should log with warn level.")
}
