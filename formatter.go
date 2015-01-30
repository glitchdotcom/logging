package logging

import (
	"fmt"
	"strings"
	"time"
)

//LogFormat is the name of a known formatting function.
type LogFormat string

//MINIMAL describes a formatter that just prints the message, replays are not indicated
const MINIMAL LogFormat = "minimal"

//MINIMALTAGGED describes a formatter that just prints the level, tags and message, replays are not indicated
const MINIMALTAGGED LogFormat = "minimaltagged"

//SIMPLE describes a formatter that just prints the date, level and message, replays are not indicated
const SIMPLE LogFormat = "simple"

//FULL formats messages with the date to ms accuracy, the level, tags and message. Replayed messages have a special field added.
const FULL LogFormat = "full"

//FormatFromString converts a string name to a LogFormat. Valid
//arguemnts include full, simple, minimaltagged and minimal. An
//unknown string will be treated like simple.
func FormatFromString(formatName string) LogFormat {
	formatName = strings.ToLower(formatName)
	switch formatName {
	case "full":
		return FULL
	case "simple":
		return SIMPLE
	case "minimaltagged":
		return MINIMALTAGGED
	case "minimal":
		return MINIMAL
	default:
		return SIMPLE
	}
}

//GetFormatter returns the function associated with a named format.
func GetFormatter(formatName LogFormat) LogFormatter {
	switch formatName {
	case FULL:
		return fullFormat
	case SIMPLE:
		return simpleFormat
	case MINIMALTAGGED:
		return minimalWithTagsFormat
	case MINIMAL:
		return minimalFormat
	default:
		return simpleFormat
	}
}

//LogFormatter is a function type used to convert a log record into a string.
//Original time is provided times when the formatter has to construct a replayed message from the buffer
type LogFormatter func(level LogLevel, tags []string, message string, t time.Time, original time.Time) string

func fullFormat(level LogLevel, tags []string, message string, t time.Time, original time.Time) string {

	if original != t {
		message = fmt.Sprintf("[replayed from %v] %v", original.Format(time.StampMilli), message)
	}

	if tags != nil && len(tags) > 0 {
		return fmt.Sprintf("[%v] [%v] %v %v", t.Format(time.StampMilli), level, tags, message)
	}
	return fmt.Sprintf("[%v] [%v] %v", t.Format(time.StampMilli), level, message)
}

func simpleFormat(level LogLevel, tags []string, message string, t time.Time, original time.Time) string {
	return fmt.Sprintf("[%v] [%v] %v", t.Format(time.Stamp), level, message)
}

func minimalFormat(level LogLevel, tags []string, message string, t time.Time, original time.Time) string {
	return message
}

func minimalWithTagsFormat(level LogLevel, tags []string, message string, t time.Time, original time.Time) string {
	if tags != nil && len(tags) > 0 {
		return fmt.Sprintf("[%v] %v %v", level, tags, message)
	}
	return fmt.Sprintf("[%v] %v", level, message)
}
