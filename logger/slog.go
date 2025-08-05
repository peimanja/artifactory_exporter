package logger

import (
	"log/slog"
	"os"
)

// New returns configured instance of `log/slog`
//
// Expects configuration in the Config type structure.
// In the absence of appropriate information in the provided configuration,
// values `FormatDefault` or `LevelDefault` will be assumed.
func New(c Config) *slog.Logger {

	lvl := lvlFromConfig(c)

	switch lf := fmtFromConfig(c); lf {
	case fmtTXT:
		return newTXTLogger(lvl)
	case fmtJSON:
		return newJSONLogger(lvl)
	default:
		l := newTXTLogger(lvl)
		l.Error("We should never have ended up here! Plase report an issue")
		os.Exit(1)
	}

	/*
	 * The following should never happen and is only there
	 * to satisfy the compiler's formal requirements.
	 */
	return newTXTLogger(lvl)
}

// formatsAvailableMap provides O(1) lookup for format validation
var formatsAvailableMap = map[string]bool{
	FormatDefault: true,
	fmtJSON:       true,
}

func fmtFromConfig(c Config) string {
	if c.Format != "" {
		// Validate format using map lookup for better performance
		if formatsAvailableMap[c.Format] {
			return c.Format
		}
		// Return default for invalid formats
		return FormatDefault
	}
	return FormatDefault
}

func lvlFromConfig(c Config) slog.Level {
	lvlFromFlag := LevelDefault
	if c.Level != "" {
		lvlFromFlag = c.Level
	}
	return levelsFlagToSlog[lvlFromFlag]
}

func newJSONLogger(l slog.Level) *slog.Logger {
	h := slog.NewJSONHandler(
		os.Stderr, // Please read the explanation in the `doc.go` file.
		&slog.HandlerOptions{
			Level: l,
		},
	)
	return slog.New(h)
}

func newTXTLogger(l slog.Level) *slog.Logger {
	h := slog.NewTextHandler(
		os.Stderr, // Please read the explanation in the `doc.go` file.
		&slog.HandlerOptions{
			Level: l,
		},
	)
	return slog.New(h)
}
