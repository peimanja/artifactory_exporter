// Package logger wraps configuration of `log/slog`.
//
// It is used by "github.com/peimanja/artifactory_exporter/config"
// and is not expected to be used outside the artifactory exporter.
//
// To maintain backward compatibility, log to os.Stderr
// as already used "github.com/prometheus/common/promlog".

package logger
