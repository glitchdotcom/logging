package logging

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFormatFromString(t *testing.T) {
	assert.Equal(t, FormatFromString("FuLl"), FULL, "formats are case insensitive")
	assert.Equal(t, FormatFromString("SimplE"), SIMPLE, "formats are case insensitive")
	assert.Equal(t, FormatFromString("MinimalTagged"), MINIMALTAGGED, "formats are case insensitive")
	assert.Equal(t, FormatFromString("Minimal"), MINIMAL, "formats are case insensitive")
	assert.Equal(t, FormatFromString("foo"), SIMPLE, "default is simple")
}

func TestFormatGetFormatter(t *testing.T) {
	assert.Equal(t, GetFormatter(FULL), LogFormatter(fullFormat), "should be full")
	assert.Equal(t, GetFormatter(SIMPLE), LogFormatter(simpleFormat), "should be simple")
	assert.Equal(t, GetFormatter(MINIMALTAGGED), LogFormatter(minimalWithTagsFormat), "should be minimal tagged")
	assert.Equal(t, GetFormatter(MINIMAL), LogFormatter(minimalFormat), "should be minimal")
	assert.Equal(t, GetFormatter(LogFormat("foo")), LogFormatter(simpleFormat), "should be simple")
}

func TestFormatFull(t *testing.T) {

	at := time.Unix(1000, 0)
	original := at.AddDate(0, 0, 1)

	expected := "[Dec 31 16:16:40.000] [INFO] [one two] [replayed from Jan  1 16:16:40.000] hello"
	assert.Equal(t, fullFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))

	expected = "[Dec 31 16:16:40.000] [INFO] [replayed from Jan  1 16:16:40.000] hello"
	assert.Equal(t, fullFormat(INFO, nil, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))

	expected = "[Dec 31 16:16:40.000] [INFO] [one two] hello"
	assert.Equal(t, fullFormat(INFO, []string{"one", "two"}, "hello", at, at), expected, fmt.Sprintf("should equal %s", expected))

	expected = "[Dec 31 16:16:40.000] [INFO] [one two] [replayed from Jan  1 16:16:40.000] hello"
	assert.Equal(t, fullFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
}

func TestFormatSimple(t *testing.T) {

	at := time.Unix(1000, 0)
	original := at.AddDate(0, 0, 1)

	expected := "[Dec 31 16:16:40] [INFO] hello"
	assert.Equal(t, simpleFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
	assert.Equal(t, simpleFormat(INFO, nil, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
	assert.Equal(t, simpleFormat(INFO, []string{"one", "two"}, "hello", at, at), expected, fmt.Sprintf("should equal %s", expected))
	assert.Equal(t, simpleFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
}

func TestFormatMinimal(t *testing.T) {

	at := time.Unix(1000, 0)
	original := at.AddDate(0, 0, 1)

	expected := "hello"
	assert.Equal(t, minimalFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
	assert.Equal(t, minimalFormat(INFO, nil, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
	assert.Equal(t, minimalFormat(INFO, []string{"one", "two"}, "hello", at, at), expected, fmt.Sprintf("should equal %s", expected))
	assert.Equal(t, minimalFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
}

func TestFormatMinimalWithTags(t *testing.T) {

	at := time.Unix(1000, 0)
	original := at.AddDate(0, 0, 1)

	expected := "[INFO] [one two] hello"
	assert.Equal(t, minimalWithTagsFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))

	expected = "[INFO] hello"
	assert.Equal(t, minimalWithTagsFormat(INFO, nil, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))

	expected = "[INFO] [one two] hello"
	assert.Equal(t, minimalWithTagsFormat(INFO, []string{"one", "two"}, "hello", at, at), expected, fmt.Sprintf("should equal %s", expected))

	expected = "[INFO] [one two] hello"
	assert.Equal(t, minimalWithTagsFormat(INFO, []string{"one", "two"}, "hello", at, original), expected, fmt.Sprintf("should equal %s", expected))
}
