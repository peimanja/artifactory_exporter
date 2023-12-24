package logger

import "log/slog"

type Config struct {
	Format string
	Level  string
}

const (
	FormatDefault  = fmtTXT
	FormatFlagHelp = "Output format of log messages. One of: [logfmt, json]"
	FormatFlagName = "log.format"
	fmtJSON        = "json"
	fmtTXT         = "logfmt"
)
const (
	lvlFNameDebug = "debug"
	lvlFNameInfo  = "info"
	lvlFNameWarn  = "warn"
	lvlFNameError = "error"
	LevelDefault  = lvlFNameInfo
	LevelFlagHelp = "Only log messages with the given severity or above. One of: [debug, info, warn, error]"
	LevelFlagName = "log.level"
)

var (
	EmptyConfig      = Config{}
	FormatsAvailable = []string{
		fmtTXT,
		fmtJSON,
	}
	LevelsAvailable = []string{
		// Deliberately not in alphabetical order, but according to the significance of the levels.
		lvlFNameDebug,
		lvlFNameInfo,
		lvlFNameWarn,
		lvlFNameError,
	}
	levelsFlagToSlog = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
)
