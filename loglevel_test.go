package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLevelToString(t *testing.T) {

	assert.Equal(t, DEBUG.String(), "DEBUG", "DEBUG.String() = %v, want %v", DEBUG, "DEBUG")
	assert.Equal(t, INFO.String(), "INFO", "INFO.String() = %v, want %v", INFO, "INFO")
	assert.Equal(t, WARN.String(), "WARN", "WARN.String() = %v, want %v", WARN, "WARN")
	assert.Equal(t, ERROR.String(), "ERROR", "ERROR.String() = %v, want %v", ERROR, "ERROR")
	assert.Equal(t, VERBOSE.String(), "VERBOSE", "VERBOSE.String() = %v, want %v", VERBOSE, "VERBOSE")
	assert.Equal(t, LogLevel(0).String(), "VERBOSE", "VERBOSE.String() = %v, want %v", VERBOSE, "VERBOSE")
}

func TestFromString(t *testing.T) {

	levelStrings := []string{"debug", "Debug", "warn", "Warning", "error", "INFO", "Informative", "verBose", "none"}
	levels := []LogLevel{DEBUG, DEBUG, WARN, WARN, ERROR, INFO, INFO, VERBOSE, DEFAULT}

	for i, levelString := range levelStrings {
		level := levels[i]
		assert.Equal(t, LevelFromString(levelString), level, "%v = %v, want %v", levelString, LevelFromString(levelString), level)
	}
}

func TestCheckLevel(t *testing.T) {

	logger := DefaultLogger()

	SetDefaultLogLevel(ERROR)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Error")
	assert.False(t, logger.CheckLevel(WARN, nil), "Warning should not be valid when level set to Error")
	assert.False(t, logger.CheckLevel(INFO, nil), "Info should not be valid when level set to Error")
	assert.False(t, logger.CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Error")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Error")

	SetDefaultLogLevel(WARN)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Warn")
	assert.True(t, logger.CheckLevel(WARN, nil), "Warning should be valid when level set to Warn")
	assert.False(t, logger.CheckLevel(INFO, nil), "Info should not be valid when level set to Warn")
	assert.False(t, logger.CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Warn")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Warn")

	SetDefaultLogLevel(INFO)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Info")
	assert.True(t, logger.CheckLevel(WARN, nil), "Warning should be valid when level set to Info")
	assert.True(t, logger.CheckLevel(INFO, nil), "Info should be valid when level set to Info")
	assert.False(t, logger.CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Info")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Info")

	SetDefaultLogLevel(DEBUG)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(WARN, nil), "Warning should be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(INFO, nil), "Info should be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(DEBUG, nil), "Debug should be valid when level set to Debug")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Debug")

	SetDefaultLogLevel(VERBOSE)
	assert.True(t, logger.CheckLevel(VERBOSE, nil), "Verbose should be valid when level set to Verbose")
}

func TestCheckTagLevel(t *testing.T) {

	logger := DefaultLogger()
	SetDefaultLogLevel(ERROR)

	tags := []string{"tag"}
	SetDefaultTagLogLevel("tag", ERROR)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(WARN, tags), "Warning should not be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(INFO, tags), "Info should not be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Error")

	SetDefaultTagLogLevel("tag", WARN)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Warn")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(INFO, tags), "Info should not be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Warn")

	SetDefaultTagLogLevel("tag", INFO)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Info")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Info")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should be valid when tag level set to Info")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Info")

	SetDefaultTagLogLevel("tag", DEBUG)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(DEBUG, tags), "Debug should be valid when tag level set to Debug")
}

func TestCheckLevelDefault(t *testing.T) {

	SetDefaultLogLevel(ERROR)
	assert.True(t, CheckLevel(ERROR, nil), "Error should be valid when level set to Error")
	assert.False(t, CheckLevel(WARN, nil), "Warning should not be valid when level set to Error")
	assert.False(t, CheckLevel(INFO, nil), "Info should not be valid when level set to Error")
	assert.False(t, CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Error")
	assert.False(t, CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Error")

	SetDefaultLogLevel(WARN)
	assert.True(t, CheckLevel(ERROR, nil), "Error should be valid when level set to Warn")
	assert.True(t, CheckLevel(WARN, nil), "Warning should be valid when level set to Warn")
	assert.False(t, CheckLevel(INFO, nil), "Info should not be valid when level set to Warn")
	assert.False(t, CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Warn")
	assert.False(t, CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Warn")
}

func TestCheckInstanceLevel(t *testing.T) {

	logger := GetLogger("test")

	SetDefaultLogLevel(ERROR)

	logger.SetLogLevel(ERROR)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Error")
	assert.False(t, logger.CheckLevel(WARN, nil), "Warning should not be valid when level set to Error")
	assert.False(t, logger.CheckLevel(INFO, nil), "Info should not be valid when level set to Error")
	assert.False(t, logger.CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Error")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Error")

	logger.SetLogLevel(WARN)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Warn")
	assert.True(t, logger.CheckLevel(WARN, nil), "Warning should be valid when level set to Warn")
	assert.False(t, logger.CheckLevel(INFO, nil), "Info should not be valid when level set to Warn")
	assert.False(t, logger.CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Warn")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Warn")

	logger.SetLogLevel(INFO)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Info")
	assert.True(t, logger.CheckLevel(WARN, nil), "Warning should be valid when level set to Info")
	assert.True(t, logger.CheckLevel(INFO, nil), "Info should be valid when level set to Info")
	assert.False(t, logger.CheckLevel(DEBUG, nil), "Debug should not be valid when level set to Info")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Info")

	logger.SetLogLevel(DEBUG)
	assert.True(t, logger.CheckLevel(ERROR, nil), "Error should be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(WARN, nil), "Warning should be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(INFO, nil), "Info should be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(DEBUG, nil), "Debug should be valid when level set to Debug")
	assert.False(t, logger.CheckLevel(VERBOSE, nil), "Verbose should not be valid when level set to Debug")

	logger.SetLogLevel(VERBOSE)
	assert.True(t, logger.CheckLevel(VERBOSE, nil), "Verbose should be valid when level set to Verbose")
}

func TestCheckInstanceTagLevel(t *testing.T) {

	logger := GetLogger("test2")

	SetDefaultLogLevel(ERROR)
	logger.SetLogLevel(ERROR)
	SetDefaultTagLogLevel("tag2", ERROR)

	tags := []string{"tag2"}
	logger.SetTagLevel("tag2", ERROR)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(WARN, tags), "Warning should not be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(INFO, tags), "Info should not be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(VERBOSE, tags), "Verbose should not be valid when level set to Error")

	logger.SetTagLevel("tag2", WARN)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Warn")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(INFO, tags), "Info should not be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(VERBOSE, tags), "Verbose should not be valid when level set to Warn")

	logger.SetTagLevel("tag2", INFO)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Info")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Info")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should be valid when tag level set to Info")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Info")
	assert.False(t, logger.CheckLevel(VERBOSE, tags), "Verbose should not be valid when level set to Info")

	logger.SetTagLevel("tag2", DEBUG)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(DEBUG, tags), "Debug should be valid when tag level set to Debug")
	assert.False(t, logger.CheckLevel(VERBOSE, tags), "Verbose should not be valid when level set to Debug")

	logger.SetTagLevel("tag2", VERBOSE)
	assert.True(t, logger.CheckLevel(VERBOSE, tags), "Verbose should be valid when level set to Verbose")
}

func TestCheckInstanceTagLevelMany(t *testing.T) {

	logger := GetLogger("test2")

	SetDefaultLogLevel(ERROR)
	logger.SetLogLevel(ERROR)
	SetDefaultTagLogLevel("tag2", ERROR)

	tags := []string{"tag2", "tag3", "tag4", "tag5", "tag6", "tag7"}
	logger.SetTagLevel("tag2", ERROR)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(WARN, tags), "Warning should not be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(INFO, tags), "Info should not be valid when tag level set to Error")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Error")

	logger.SetTagLevel("tag2", WARN)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Warn")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(INFO, tags), "Info should not be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Warn")

	logger.SetTagLevel("tag2", INFO)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Info")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Info")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should be valid when tag level set to Info")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Info")

	logger.SetTagLevel("tag2", DEBUG)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should be valid when tag level set to Debug")
	assert.True(t, logger.CheckLevel(DEBUG, tags), "Debug should be valid when tag level set to Debug")
}

func TestCheckInstanceTagLevelVsOther(t *testing.T) {

	logger := GetLogger("test3")

	SetDefaultLogLevel(ERROR)
	logger.SetLogLevel(ERROR)
	SetDefaultTagLogLevel("tag3", ERROR)

	GetLogger("test4").SetTagLevel("tag3", DEBUG) //Check for tag level leaking

	tags := []string{"tag3"}

	logger.SetTagLevel("tag3", WARN)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when tag level set to Warn")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(INFO, tags), "Info should not be valid when tag level set to Warn")
	assert.False(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when tag level set to Warn")
}

func TestDefaultLevelOveridesTagLevel(t *testing.T) {

	logger := GetLogger("test5")
	SetDefaultLogLevel(DEBUG)

	tags := []string{"tag5"}
	SetDefaultTagLogLevel("tag5", ERROR)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when default level set to Debug")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should not be valid when default level set to Debug")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should not be valid when default level set to Debug")
	assert.True(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when default level set to Debug")
}

func TestLevelOveridesTagLevel(t *testing.T) {

	logger := GetLogger("test5")
	SetDefaultLogLevel(ERROR)
	logger.SetLogLevel(DEBUG)

	tags := []string{"tag6"}
	SetDefaultTagLogLevel("tag6", ERROR)
	assert.True(t, logger.CheckLevel(ERROR, tags), "Error should be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(WARN, tags), "Warning should not be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(INFO, tags), "Info should not be valid when level set to Debug")
	assert.True(t, logger.CheckLevel(DEBUG, tags), "Debug should not be valid when level set to Debug")
}

func BenchmarkCheckPassingLogLevel(b *testing.B) {
	logger := GetLogger("BenchmarkCheckPassingLogLevel")
	logger.SetLogLevel(ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(ERROR, nil)
	}
}

func BenchmarkCheckFailingLogLevel(b *testing.B) {
	logger := GetLogger("BenchmarkCheckFailingLogLevel")
	logger.SetLogLevel(ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(WARN, nil)
	}
}

func BenchmarkCheckPassingTagLevel(b *testing.B) {
	logger := GetLogger("BenchmarkCheckPassingTagLevel")
	logger.SetLogLevel(ERROR)
	logger.SetLogLevel(ERROR)
	tags := []string{"tag2"}
	logger.SetTagLevel("tag2", ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(ERROR, tags)
	}
}

func BenchmarkCheckFailingTagLevel(b *testing.B) {
	logger := GetLogger("BenchmarkCheckFailingTagLevel")
	logger.SetLogLevel(ERROR)
	tags := []string{"tag2"}
	logger.SetTagLevel("tag2", ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(WARN, tags)
	}
}

func BenchmarkCheckPassingTagLevelThree(b *testing.B) {
	logger := GetLogger("BenchmarkCheckPassingTagLevel")
	logger.SetLogLevel(ERROR)
	logger.SetLogLevel(ERROR)
	tags := []string{"alpha", "beta", "phi"}
	logger.SetTagLevel("alpha", ERROR)
	SetDefaultTagLogLevel("beta", ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(ERROR, tags)
	}
}

func BenchmarkCheckFailingTagLevelThree(b *testing.B) {
	logger := GetLogger("BenchmarkCheckFailingTagLevel")
	logger.SetLogLevel(ERROR)
	tags := []string{"alpha", "beta", "phi"}
	logger.SetTagLevel("alpha", ERROR)
	SetDefaultTagLogLevel("beta", ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(WARN, tags)
	}
}

func BenchmarkCheckPassingTagLevelNine(b *testing.B) {
	logger := GetLogger("BenchmarkCheckPassingTagLevel")
	logger.SetLogLevel(ERROR)
	logger.SetLogLevel(ERROR)
	tags := []string{"alpha", "beta", "gamma", "delta", "epsilon", "tau", "pi", "phi", "psi"}
	logger.SetTagLevel("alpha", ERROR)
	SetDefaultTagLogLevel("beta", ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(ERROR, tags)
	}
}

func BenchmarkCheckFailingTagLevelNine(b *testing.B) {
	logger := GetLogger("BenchmarkCheckFailingTagLevel")
	logger.SetLogLevel(ERROR)
	tags := []string{"alpha", "beta", "gamma", "delta", "epsilon", "tau", "pi", "phi", "psi"}
	logger.SetTagLevel("alpha", ERROR)
	SetDefaultTagLogLevel("beta", ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(WARN, tags)
	}
}

func BenchmarkCheckFailingTagLevelEightteen(b *testing.B) {
	logger := GetLogger("BenchmarkCheckFailingTagLevel")
	logger.SetLogLevel(ERROR)
	tags := []string{"alpha", "beta", "gamma", "delta", "epsilon", "tau", "pi", "phi", "psi",
		"zeta", "omega", "upsilon", "one", "two", "three", "four", "five", "six",
	}
	logger.SetTagLevel("alpha", ERROR)
	SetDefaultTagLogLevel("beta", ERROR)
	for n := 0; n < b.N; n++ {
		logger.CheckLevel(WARN, tags)
	}
}

func BenchmarkTagRangeList(b *testing.B) {
	theMap := make(map[string]string, 0)
	theMap["alpha"] = "alpha"
	tags := []string{"alpha", "beta", "gamma", "delta", "epsilon", "tau", "pi", "phi", "psi",
		"zeta", "omega", "upsilon", "one", "two", "three", "four", "five", "six",
	}
	for n := 0; n < b.N; n++ {
		for _, tag := range tags {
			_, _ = theMap[tag]
		}
	}
}

func BenchmarkTagMapLookup(b *testing.B) {
	theMap := make(map[string]string, 0)
	theMap["alpha"] = "a"
	for n := 0; n < b.N; n++ {
		_, _ = theMap["alpha"]
	}
}

func BenchmarkTagFailedMapLookup(b *testing.B) {
	theMap := make(map[string]string, 0)
	theMap["beta"] = "b"
	for n := 0; n < b.N; n++ {
		_, _ = theMap["a"]
	}
}
