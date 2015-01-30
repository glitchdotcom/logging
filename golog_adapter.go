package logging

import (
	"log"
)

type goLogAdapter struct {
	level LogLevel
	tags  []string
}

func (adapter *goLogAdapter) Write(p []byte) (n int, err error) {
	s := string(p[:])
	defaultLogger.log(adapter.level, adapter.tags, s)
	return len(p), nil
}

//AdaptStandardLogging points the standard logging to fog creek logging
//using the provided level and tags. The default will be info with no tags.
func AdaptStandardLogging(level LogLevel, tags []string) {
	adapter := goLogAdapter{
		level: level,
		tags:  tags,
	}

	log.SetFlags(0)
	log.SetOutput(&adapter)
}
