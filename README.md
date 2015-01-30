Logging is generally broken into two pieces: what to log and where to log it.

These values are often specified by named loggers with formatters that specify the format of log messages,
a level that specifies how much to log and some sort of target to specify where to log. The fog creek logging package for Go takes a slightly different approach based on our experience with real world logging.

A default logger provides an easy path to logging. The logging package provides methods to directly log ot the default logger without accessing it

	logging.Info("information message")

is the same as

	logging.DefaultLogger().Info("information message")

You can also use named loggers, to scope log levels to packages.

	logging.GetLogger("my logger").Infof("%v", aVar)

Tags provide a way to set levels across concepts, perpendicular to logger names.

	tags := ["network", "server"]
	logging.DebugWithTags(tags, "tagged message")

Levels are used to filter logging. There are five levels:

 * VERBOSE
 * DEBUG
 * INFO
 * WARN
 * ERROR

By default INFO messages and above are allowed. Each logger can have a level. A method to set the default level is available, and actually sets the level on the default logger.

logging.SetDefaultLogLevel(logging.INFO)

is equivalent to

logging.DefaultLogger().SetLogLevel(logging.INFO)

Levels can also be set by tag, which overrides the loggers level. By default loggers use the default loggers level.

Loggers have a four methods for each log level, one for formatted messages, and one for just a simple array of values, plus a version of each of these with tags.

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

Log messages are formatted by a format function. Which can be set per logger, by default loggers copy the default loggers format.

Logging ultimately goes through one or more appenders to get the messages to the console, a file or wherever.
All loggers share the same appenders - but appenders can be associated with a level which is unrelated to tags.

Each logger has an optional buffer, that will be flushed whenever its level/tags change.
This buffer contains un-passed messages. So that it is possible to configure the system to capture messages and replay them latter.

Replayed messages are tagged and have a double time stamp.

To use go vet with this package you can use the form:

    % go tool vet -printfuncs "ErrorWithTagsf,Errorf,WarnWithTagsf,Warnf,InfoWithTagsf,Infof,DebugWithTagsf,Debugf" <package>

To use logging with your project simply:

    % go get github.com/fogcreek/logging

