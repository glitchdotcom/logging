package logging

import "strings"

//LogLevel is the type used to indicate the importance of a logging request
type LogLevel uint8

const (
	//DEFAULT is the default log level, loggers with default level will use the default loggers level
	DEFAULT LogLevel = iota
	//VERBOSE is the wordiest log level, useful for very big text output that may
	//be the last result during testing or debugging
	VERBOSE
	//DEBUG is generally the lowest level used when testing
	DEBUG
	//INFO is used for generally helpful but not important messages
	INFO
	//WARN is provided for warnings that do not represent a major program error
	WARN
	//ERROR is the highest log level and should only be used for exceptional conditions
	ERROR
)

//String converts a log level to an upper case string
func (level LogLevel) String() string {
	switch {
	case level >= ERROR:
		return "ERROR"
	case level >= WARN:
		return "WARN"
	case level >= INFO:
		return "INFO"
	case level >= DEBUG:
		return "DEBUG"
	default:
		return "VERBOSE"
	}
}

/*
LevelFromString converts a level in any case to a LogLevel, valid values are
error, warning, warn, info, informative, debug and verbose.
*/
func LevelFromString(str string) LogLevel {
	str = strings.ToLower(str)

	switch str {
	case "error":
		return ERROR
	case "warning", "warn":
		return WARN
	case "info", "informative":
		return INFO
	case "debug":
		return DEBUG
	case "verbose":
		return VERBOSE
	default:
		return DEFAULT
	}
}
